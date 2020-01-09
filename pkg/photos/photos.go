package photos

import (
	"context"

	"github.com/dgraph-io/badger"
	"github.com/sirupsen/logrus"
)

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
