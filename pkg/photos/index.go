package photos

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/dgraph-io/badger"
	"github.com/fsnotify/fsnotify"
	"github.com/saracen/walker"
	"github.com/sirupsen/logrus"
)

type Indexer struct {
	photoDir string
	watcher  *fsnotify.Watcher
	ctx      context.Context
	log      logrus.FieldLogger
	db       *badger.DB
}

// index files if necessary (write times checked)
// relative path given
func (idx *Indexer) Index(path string, recur bool) {
	l := idx.log.WithField("path", path)
	fullpath := filepath.Join(idx.photoDir, path)

	// if file, do the business
	fi, err := os.Stat(fullpath)
	if err != nil {
		l.WithError(err).Error("unable to stat path")
		return
	}

	if !fi.IsDir() {
		if err := idx.indexFileIfNeeded(path, nil); err != nil {
			l.WithError(err).Error("error indexing file")
		}
		return
	}

	if idx.db == nil {
		l.Error("db not connected. Cannot index")
		return
	}

	// if dir, do the business for each file
	var wg sync.WaitGroup
	batch := idx.db.NewWriteBatch()
	err = walker.WalkWithContext(idx.ctx, fullpath, func(name string, fi os.FileInfo) error {
		if fi.IsDir() {
			if recur {
				return nil
			} else {
				return filepath.SkipDir
			}
		}
		if strings.HasSuffix(name, ".xmp") {
			return nil
		}
		filename := idx.relpath(name)
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := idx.indexFileIfNeeded(filename, batch); err != nil {
				l.WithField("file", filename).WithError(err).Error("erroring indexing file in path")
			}
		}()
		return nil
	})
	if err != nil {
		l.WithError(err).Error("error when crawling path")
	}
	wg.Wait()
	batch.Flush()
}

// writes file info to DB if either XMP or EXIF are out of date.
// checks write times. expects relative path
// batcher is optional db batchwriter. nil is acceptable
func (idx *Indexer) indexFileIfNeeded(file string, batcher *badger.WriteBatch) error {
	xmp, exif, err := idx.needsIndex(file)
	if err != nil {
		return err
	}
	return idx.indexFile(file, xmp, exif, batcher)
}

// indexes a file. expects relative path
func (idx *Indexer) indexFile(file string, xmp bool, exif bool, batcher *badger.WriteBatch) error {
	l := idx.log.WithField("file", file)
	fullpath := filepath.Join(idx.photoDir, file)
	var wg sync.WaitGroup
	if xmp {
		wg.Add(1)
		go func() {

			defer wg.Done()
			if x, err := ReadXMPFile(fullpath + ".xmp"); err != nil {
				l.WithError(err).Error("error reading XMP file")
			} else {
				l.Debug("indexing XMP data")
				if err := Write(idx.ctx, SourceXMP, file, x, batcher); err != nil {
					l.WithError(err).Error("error writing XMP to db")
				}
				toIndex := [][2]string{
					[2]string{"derived_from", x.DerivedFromFile},
					[2]string{"rating", strconv.Itoa(x.Rating)},
					[2]string{"auto_presets_applied", strconv.FormatBool(x.AutoPresets)},
					[2]string{"xmp_version", strconv.Itoa(x.XMPVersion)},
					[2]string{"creator", x.Creator},
					[2]string{"rights", x.Rights},
					[2]string{"title", x.Title},
				}
				if x.Location != nil {
					toIndex = append(toIndex, [2]string{"loc.lat", x.Location.Lat})
					toIndex = append(toIndex, [2]string{"loc.lon", x.Location.Lon})
					toIndex = append(toIndex, [2]string{"loc.alt", x.Location.Altitude})
				}
				for _, c := range x.ColorLabels {
					toIndex = append(toIndex, [2]string{"color_labels", c})
				}
				for _, t := range x.Tags {
					toIndex = append(toIndex, [2]string{"tags", t})
				}
				for _, h := range x.History {
					if !h.Enabled {
						continue
					}
					toIndex = append(toIndex, [2]string{"history", h.OpName})
				}

				for _, ti := range toIndex {
					if ti[1] == "" {
						continue // skip blank values
					}
					if err := WriteIdxField(idx.ctx, SourceXMP, file, ti[0], []byte(ti[1]), batcher); err != nil {
						l.WithError(err).WithField("field", ti[0]).WithField("value", ti[1]).Error("error writing XMP field to index")
					}
				}
			}
		}()
	}

	if exif {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if data, err := ReadExifFile(fullpath); err != nil {
				l.WithError(err).Error("error reading exif")
			} else {
				l.Debug("indexing EXIF data")
				if err := Write(idx.ctx, SourceEXIF, file, data, batcher); err != nil {
					l.WithError(err).Error("error writing exif to db")
				}
				for k, v := range data {
					var s string
					switch tv := v.(type) {
					case int:
						s = strconv.Itoa(tv)
					case int64:
						s = strconv.FormatInt(tv, 10)
					case float64:
						s = strconv.FormatFloat(tv, 'g', -1, 64)
					case bool:
						s = strconv.FormatBool(tv)
					case string:
						s = tv
					default:
						s = fmt.Sprintf("unhandled :: %T", v)
					}
					if s == "(none)" || s == "" || s == "n/a" {
						continue
					}
					// truncate to avoid using space
					if len(s) > 120 {
						s = s[:120]
					}
					if err := WriteIdxField(idx.ctx, SourceEXIF, file, k, []byte(s), batcher); err != nil {
						l.WithError(err).WithField("field", k).WithField("value", s).Error("error writing EXIF field to index")
					}
				}
			}
		}()
	}

	wg.Wait()
	return nil
}

// whether (xmp, exif) need to be reindexed for being out-of-date or missing in DB
func (idx *Indexer) needsIndex(file string) (bool, bool, error) {
	fullpath := filepath.Join(idx.photoDir, file)
	l := idx.log.WithField("file", file)
	if idx.db == nil {
		l.Error("db not connected")
		return true, true, errors.New("db not connected")
	}
	type metaState struct {
		needIndex bool // needs to be re-read from file and written to db
		exists    bool // if files still exist on disk
		sourceMod time.Time
		dbMod     time.Time
	}

	xmp := metaState{}
	exif := metaState{}

	if fi, err := os.Stat(fullpath + ".xmp"); err != nil {
		if !os.IsNotExist(err) {
			l.WithError(err).Trace("problem looking at associated XMP file")
		}
		xmp.exists = false
	} else {
		xmp.exists = true
		xmp.sourceMod = fi.ModTime()
	}

	if fi, err := os.Stat(fullpath); err != nil {
		l.WithError(err).Error("unable to find source file for info")
		return true, true, err // source file must exist
	} else {
		exif.exists = true
		exif.sourceMod = fi.ModTime()
	}

	// check db record exists, & timestamp
	err := idx.db.View(func(tx *badger.Txn) error {
		// process XMP
		if xmp.exists {
			// get and check db write time against file mod time
			if t, err := getAsTime(tx, TimeKey(file, SourceXMP)); err != nil {
				if err != badger.ErrKeyNotFound {
					l.WithError(err).Trace("cannot read db XMP.time write time. Will read XMP from file")
				}
				xmp.needIndex = true
			} else {
				xmp.dbMod = t
				l.WithField("file mod", xmp.sourceMod).WithField("db mod", xmp.dbMod).Trace("comparing modification times of XMP")
				if xmp.sourceMod.After(xmp.dbMod) {
					l.Trace("XMP source file modified. Will read from file")
					xmp.needIndex = true
				}
			}
		}

		if t, err := getAsTime(tx, TimeKey(file, SourceEXIF)); err != nil {
			if err != badger.ErrKeyNotFound {
				l.WithError(err).Trace("error reading exif mod time from db. Will read exif from file")
			}
			exif.needIndex = true
		} else {
			exif.dbMod = t
			l.WithField("file mod", exif.sourceMod).WithField("db mod", exif.dbMod).Trace("comparing modification times of EXIF data")
			if exif.sourceMod.After(exif.dbMod) {
				l.Trace("EXIF file modified. Will read from file")
				exif.needIndex = true
			}
		}
		return nil
	})
	if err != nil {
		l.WithError(err).Error("error from DB transaction")
		return true, true, err
	}

	return xmp.needIndex, exif.needIndex, nil
}

// relative path needed
func (idx *Indexer) dropIndex(file string) error {
	idx.log.WithField("path", file).Debug("dropping index")
	key := DataKey(file, SourceXMP)

	del := [][2]byte{
		[2]byte{SourceEXIF, DataRecord},
		[2]byte{SourceEXIF, TimestampRecord},
		[2]byte{SourceXMP, DataRecord},
		[2]byte{SourceXMP, TimestampRecord},
	}

	return idx.db.Update(func(tx *badger.Txn) error {
		for _, t := range del {
			nk := make([]byte, len(key))
			copy(nk, key)
			copy(nk[1:], t[:])
			if err := tx.Delete(nk); err != nil {
				idx.log.WithError(err).Error("error deleting index")
			}
		}

		pfx := []byte{indexRecord}
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false
		opts.Prefix = pfx
		fk := []byte(file)
		it := tx.NewIterator(opts)
		defer it.Close()
		it.Rewind()
		idx.log.WithField("path", file).WithField("filekey", string(fk)).Trace("deleting search indexes")
		for it.Seek(pfx); it.ValidForPrefix(pfx); it.Next() {
			k := it.Item().KeyCopy(nil) // https://github.com/dgraph-io/badger/issues/494#issuecomment-390831885
			if bytes.HasSuffix(k, fk) {
				if err := tx.Delete(k); err != nil {
					idx.log.WithError(err).WithField("k", string(k)).Error("error deleting query index")
				}
			}
		}

		return nil
	})
}

func (idx *Indexer) StartWatcher(ctx context.Context) error {
	idx.ctx = ctx
	w, err := fsnotify.NewWatcher()
	if err != nil {
		idx.log.WithError(err).Error("error creating fsnotify watcher")
		return err
	}
	idx.watcher = w

	idx.log.Debug("beginning indexer watch loop")
	go func() {
		defer func() {
			w.Close()
			idx.watcher = nil
		}()

		wdb := writeDebouncer(idx.Index, 1500*time.Millisecond)
		for {
			select {
			case event := <-w.Events:
				if !eventIs(event, fsnotify.Write) { // may get LOTS of write events per chunk, way too much for logging
					idx.log.WithField("event", event).Trace("got watch event")
				}
				if eventIs(event, fsnotify.Remove) || eventIs(event, fsnotify.Rename) {
					go idx.dropIndex(idx.relpath(event.Name)) // nolint
				}
				if eventIs(event, fsnotify.Create) || eventIs(event, fsnotify.Write) {
					wdb <- idx.relpath(event.Name)
				}

				// if a directory is added, we should add it
				if eventIs(event, fsnotify.Create) {
					if fi, err := os.Stat(event.Name); err == nil && fi.IsDir() {
						if err := idx.Watch(event.Name); err != nil {
							idx.log.WithError(err).WithField("name", event.Name).Error("error watching directory")
						}
					}
				}
			case err := <-w.Errors:
				idx.log.WithError(err).Error("got fsnotify error")
			case <-ctx.Done():
				idx.log.Info("context canceled. Exiting watcher loop")
				return
			}
		}
	}()
	return nil
}

func writeDebouncer(idx func(string, bool), timeout time.Duration) chan string {
	incoming := make(chan string)

	go func() {
		var s string
		t := time.NewTimer(timeout)
		t.Stop()

		for {
			select {
			case s = <-incoming:
				t.Stop()
				t.Reset(timeout)
			case <-t.C:
				go idx(s, true)
			}
		}
	}()

	return incoming

}

// *ABSOLUTE* path expected
func (idx *Indexer) Watch(dir string) error {
	l := idx.log.WithField("path", dir)
	if idx.watcher == nil {
		l.Error("unable to watch. watcher not created")
		return errors.New("watcher not activated")
	}
	l.Debug("indexer requested to watch path")
	return walker.WalkWithContext(idx.ctx, dir, func(name string, fi os.FileInfo) error {
		if fi.IsDir() {
			idx.log.WithField("path", name).Trace("watching path")
			return idx.watcher.Add(name)
		}
		return nil
	})
}

// *ABSOLUTE* path expected
func (idx *Indexer) UnWatch(dir string) error {
	if idx.watcher == nil {
		return errors.New("watcher not activated")
	}
	return idx.watcher.Remove(dir)
}

func (idx Indexer) relpath(p string) string        { return strings.TrimPrefix(p, idx.photoDir+"/") }
func eventIs(e fsnotify.Event, o fsnotify.Op) bool { return e.Op&o == o }
