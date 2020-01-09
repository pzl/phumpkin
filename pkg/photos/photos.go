package photos

import (
	"context"

	"github.com/dgraph-io/badger"
	"github.com/sirupsen/logrus"
)

type Meta struct {
	DerivedFromFile string        `json:"derived_from"`
	Rating          int           `json:"rating"`
	Location        *Location     `json:"loc,omitempty"`
	AutoPresets     bool          `json:"auto_presets_applied"`
	XMPVersion      int           `json:"xmp_version"`
	ColorLabels     []string      `json:"color_labels,omitempty"`
	Creator         string        `json:"creator,omitempty"`
	History         []DTOperation `json:"history,omitempty"`
	Rights          string        `json:"rights"`
	Tags            []string      `json:"tags,omitempty"`
	Title           string        `json:"title,omitempty"`
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
	indexer Indexer
}

func New() *Mgr {
	return &Mgr{}
}

func (m *Mgr) Start(ctx context.Context) error {
	photoDir := ctx.Value("photoDir").(string)
	m.indexer.photoDir = photoDir
	m.indexer.log = ctx.Value("log").(logrus.FieldLogger)

	m.indexer.db = ctx.Value("badger").(*badger.DB)
	if err := m.indexer.StartWatcher(ctx); err != nil {
		return err
	}
	if err := m.indexer.Watch(photoDir); err != nil {
		return err
	}
	go m.indexer.Index("", true) // recursively index the photoDir
	return nil
}
