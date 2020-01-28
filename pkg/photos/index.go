package photos

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
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

// non-recursively index files if necessary (write times checked)
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
		if err := idx.indexFileIfNeeded(filename, batch); err != nil {
			l.WithField("file", filename).WithError(err).Error("erroring indexing file in path")
		}
		return nil
	})
	if err != nil {
		l.WithError(err).Error("error when crawling path")
	}
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
	if xmp {
		if x, err := ReadXMPFile(fullpath + ".xmp"); err != nil {
			l.WithError(err).Error("error reading XMP file")
		} else {
			l.Debug("indexing XMP data")
			if err := Write(idx.ctx, SourceXMP, file, x, batcher); err != nil {
				l.WithError(err).Error("error writing XMP to db")
			}
		}
	}

	if exif {
		if data, err := ReadExifFile(fullpath); err != nil {
			l.WithError(err).Error("error reading exif")
		} else {
			l.Debug("indexing EXIF data")
			if err := Write(idx.ctx, SourceEXIF, file, data, batcher); err != nil {
				l.WithError(err).Error("error writing exif to db")
			}
		}
	}
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
	return idx.db.Update(func(tx *badger.Txn) error {
		tx.Delete(key) // nolint
		key[2] = TimestampRecord
		tx.Delete(key) // nolint
		key[1] = SourceEXIF
		tx.Delete(key) // nolint
		key[2] = DataRecord
		tx.Delete(key) // nolint
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
		for {
			select {
			case event := <-w.Events:
				idx.log.WithField("event", event).Trace("got watch event")
				if eventIs(event, fsnotify.Remove) || eventIs(event, fsnotify.Rename) {
					go idx.dropIndex(idx.relpath(event.Name)) // nolint
				}
				if eventIs(event, fsnotify.Create) || eventIs(event, fsnotify.Write) {
					go idx.Index(idx.relpath(event.Name), true)
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
