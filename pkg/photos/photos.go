package photos

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"time"

	"github.com/dgraph-io/badger"
	"github.com/sirupsen/logrus"
)

type Meta struct {
	DerivedFromFile string                 `json:"derived_from"`
	Rating          int                    `json:"rating"`
	AutoPresets     bool                   `json:"auto_presets_applied"`
	XMPVersion      int                    `json:"xmp_version"`
	ColorLabels     []string               `json:"color_labels,omitempty"`
	Creator         string                 `json:"creator,omitempty"`
	History         []DTOperation          `json:"history,omitempty"`
	Rights          string                 `json:"rights"`
	Tags            []string               `json:"tags,omitempty"`
	Title           string                 `json:"title,omitempty"`
	EXIF            map[string]interface{} `json:"exif"`
}

type DTOperation struct {
	Name           string `json:"name"`
	Enabled        bool   `json:"enabled"`
	ModVersion     int    `json:"modversion"`
	Params         string `json:"params"`
	MultiName      string `json:"multi_name"`
	MultiPriority  int    `json:"multi_priority"`
	BlendOpVersion int    `json:"blendop_version"`
	BlendOpParams  string `json:"blendop_params"`
}

type Mgr struct {
	dataDir string
	db      *badger.DB
}

func New(dir string) *Mgr {
	return &Mgr{
		dataDir: dir,
		db:      nil,
	}
}

func (m *Mgr) Start(ctx context.Context) error {
	db, err := badger.Open(badger.DefaultOptions(m.dataDir))
	if err != nil {
		return err
	}
	m.db = db
	go func() {
		<-ctx.Done()
		m.Close()
	}()
	return nil
}

func (m *Mgr) Close() {
	if m.db != nil {
		m.db.Close()
		m.db = nil
	}
}

// store in db as bytes:
//  - fileID.resolution (from exif)
//  - fileID.XMP
//    + fileID.XMP.time
//  - fileID.EXIF
//    + fileID.EXIF.time

func (m *Mgr) Load(log logrus.FieldLogger, file string) (Meta, error) {
	l := log.WithField("file", file)
	if m.db == nil {
		l.Error("db not connected")
		return Meta{}, errors.New("db not connected")
	}
	type metaState struct {
		generate  bool // needs to be re-read from file and written to db
		exists    bool // if files still exist on disk
		sourceMod time.Time
		dbMod     time.Time
	}

	meta := Meta{}
	exif := make(map[string]interface{})
	xmp := metaState{}
	exifInf := metaState{}

	if fi, err := os.Stat(file + ".xmp"); err != nil {
		l.WithError(err).Trace("problem looking at associated XMP file")
		xmp.exists = false
	} else {
		xmp.exists = true
		xmp.sourceMod = fi.ModTime()
	}

	if fi, err := os.Stat(file); err != nil {
		l.WithError(err).Error("unable to find source file for info")
		return Meta{}, err // source file must exist
	} else {
		exifInf.exists = true
		exifInf.sourceMod = fi.ModTime()
	}

	// check db record exists, & timestamp
	err := m.db.View(func(tx *badger.Txn) error {
		// process XMP
		if xmp.exists {
			// get and check db write time against file mod time
			if t, err := getModTime(tx, []byte(file+".XMP.time")); err != nil {
				l.WithError(err).Trace("cannot read db XMP.time write time. Will read XMP from file")
				xmp.generate = true
			} else {
				xmp.dbMod = *t
				l.WithField("file mod", xmp.sourceMod).WithField("db mod", xmp.dbMod).Trace("comparing modification times of XMP")
				if xmp.sourceMod.After(xmp.dbMod) {
					l.Trace("XMP source file modified. Will read from file")
					xmp.generate = true
				}
			}

			// if we don't need to read file, use DB copy
			if !xmp.generate {
				l.Trace("reading XMP data from database")
				if x, err := getValue(tx, []byte(file+".XMP")); err != nil {
					l.WithError(err).Error("error reading XMP from database, will read from file")
					xmp.generate = true
				} else {
					if err := json.Unmarshal(x, &meta); err != nil {
						l.WithError(err).Error("error unmarshaling XMP data from db, will read from file.")
						xmp.generate = true
					}
				}
			}
		}

		if t, err := getModTime(tx, []byte(file+".EXIF.time")); err != nil {
			l.WithError(err).Trace("error reading exif mod time from db. Will read exif from file")
			exifInf.generate = true
		} else {
			exifInf.dbMod = *t
			l.WithField("file mod", exifInf.sourceMod).WithField("db mod", exifInf.dbMod).Trace("comparing modification times of EXIF data")
			if exifInf.sourceMod.After(exifInf.dbMod) {
				l.Trace("EXIF file modified. Will read from file")
				exifInf.generate = true
			}
		}

		if !exifInf.generate {
			l.Trace("reading EXIF data from database")
			if x, err := getValue(tx, []byte(file+".EXIF")); err != nil {
				l.WithError(err).Error("error reading EXIF from db. Will read from file")
				exifInf.generate = true
			} else {
				if err := json.Unmarshal(x, &exif); err != nil {
					l.WithError(err).Error("error unmarshaling EXIF db info. Will read from file")
					exifInf.generate = true
				} else {
					l.Trace("exif data read from db successfully")
					// meta.EXIF = exif
				}
			}
		}

		return nil
	})
	if err != nil {
		l.WithError(err).Error("error from DB transaction")
		return Meta{}, err
	}

	if xmp.generate {
		l.Info("reading XMP from file")
		if meta, err = ReadXMP(file + ".xmp"); err != nil {
			l.WithError(err).Error("error reading XMP file")
		}
		l.Trace("writing XMP to db")
		go writeXMP(l, m.db, file, meta)
	}

	if exifInf.generate {
		l.Info("reading EXIF from file")
		if exif, err = ReadExif(file); err != nil {
			l.WithError(err).Error("error reading exif")
		}
		l.Trace("writing EXIF to db")
		go writeEXIF(l, m.db, file, exif)
	}
	meta.EXIF = exif
	return meta, nil
}
