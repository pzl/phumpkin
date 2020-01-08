package photos

import (
	"context"
	"encoding/json"

	"github.com/dgraph-io/badger"
	"github.com/sirupsen/logrus"
)

type Meta struct {
	DerivedFromFile string                 `json:"derived_from"`
	Rating          int                    `json:"rating"`
	Location        *Location              `json:"loc,omitempty"`
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

type Location struct {
	Lat      string `json:"lat"`
	Lon      string `json:"lon"`
	Altitude string `json:"alt"`
}

type DTOperation struct {
	Name           string `json:"name"`
	Number         string `json:"num"`
	Enabled        bool   `json:"enabled"`
	ModVersion     int    `json:"modversion"`
	Params         string `json:"params"`
	MultiName      string `json:"multi_name"`
	MultiPriority  int    `json:"multi_priority"`
	BlendOpVersion int    `json:"blendop_version"`
	BlendOpParams  string `json:"blendop_params"`
	IOPOrder       string `json:"iop_order"`
}

type Mgr struct {
	log     logrus.FieldLogger
	db      *badger.DB
	indexer Indexer
}

func New() *Mgr {
	return &Mgr{}
}

func (m *Mgr) Start(ctx context.Context) error {
	m.log = ctx.Value("log").(logrus.FieldLogger)
	photoDir := ctx.Value("photoDir").(string)
	m.indexer.photoDir = photoDir
	m.indexer.log = m.log

	m.db = ctx.Value("badger").(*badger.DB)
	m.indexer.db = m.db
	if err := m.indexer.StartWatcher(ctx); err != nil {
		return err
	}
	if err := m.indexer.Watch(photoDir); err != nil {
		return err
	}
	go m.indexer.Index("", true) // recursively index the photoDir
	return nil
}

// gets data, from DB if possible, otherwise reads directly
func (m *Mgr) Load(log logrus.FieldLogger, file string) (Meta, error) {
	l := log.WithField("file", file)
	meta := Meta{}

	// check index status, read directly if needed
	x, e, err := m.indexer.needsIndex(file)
	if err != nil {
		log.WithError(err).Error("error checking file for current index status")
		return meta, err
	}

	if x || e {
		if err := m.indexer.indexFile(file, x, e, nil); err != nil {
			log.WithError(err).Error("error loading file into index")
			return meta, err
		}
	}

	err = m.db.View(func(tx *badger.Txn) error {
		if data, err := getValue(tx, []byte(file+".XMP")); err != nil {
			l.WithError(err).Error("error reading XMP from database")
		} else {
			if err := json.Unmarshal(data, &meta); err != nil {
				l.WithError(err).Error("error unmarshalling XMP data from db")
			}
		}

		if data, err := getValue(tx, []byte(file+".EXIF")); err != nil {
			l.WithError(err).Error("error getting EXIF from db")
			return err
		} else {
			if err := json.Unmarshal(data, &meta.EXIF); err != nil {
				l.WithError(err).Error("error unmarshalling EXIF data from db")
				return err
			}
		}
		return nil
	})
	if err != nil {
		return meta, err
	}

	return meta, nil
}
