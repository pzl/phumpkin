package photos

import (
	"bytes"
	"context"

	"github.com/dgraph-io/badger"
)

func WithLocation(ctx context.Context) ([]Photo, error) {
	db := ctx.Value("badger").(*badger.DB)
	photoDir := ctx.Value("photoDir").(string)

	pmap := make(map[string]struct{})
	err := db.View(func(tx *badger.Txn) error {
		pfx := []byte{indexRecord, SourceEXIF, 'G', 'P', 'S', 'L', 'a', 't'}
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false
		opts.Prefix = pfx
		it := tx.NewIterator(opts)
		defer it.Close()
		it.Rewind()
		for it.Seek(pfx); it.ValidForPrefix(pfx); it.Next() {
			k := it.Item().Key()

			spl := bytes.LastIndexByte(k, 0)
			if spl == -1 {
				continue
			}

			fname := string(k[spl+1:])
			pmap[fname] = struct{}{}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	ps := make([]Photo, 0, len(pmap))
	for k := range pmap {
		p, err := FromSrc(ctx, photoDir+"/"+k)
		if err != nil {
			return nil, err
		}
		ps = append(ps, p)
	}

	return ps, nil
}
