package photos

import (
	"bytes"
	"context"
	"sort"

	"github.com/dgraph-io/badger"
)

func GetFields(ctx context.Context, source byte, partial string) ([]string, error) {
	db := ctx.Value("badger").(*badger.DB)

	keymap := make(map[string]struct{})
	pfx := append([]byte{indexRecord, source}, []byte(partial)...)
	opts := badger.DefaultIteratorOptions
	opts.PrefetchValues = false
	opts.Prefix = pfx
	err := db.View(func(tx *badger.Txn) error {
		it := tx.NewIterator(opts)
		defer it.Close()
		it.Rewind()
		for it.Seek(pfx); it.ValidForPrefix(pfx); it.Next() {
			k := it.Item().Key()
			spl := bytes.IndexByte(k, 0)
			if spl == -1 {
				continue
			}

			keymap[string(k[2:spl])] = struct{}{}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	keys := make([]string, 0, len(keymap))
	for k := range keymap {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	return keys, nil
}

func GetValues(ctx context.Context, source byte, field string, partial string) ([]string, error) {
	db := ctx.Value("badger").(*badger.DB)

	vmap := make(map[string]struct{})
	pfx := append([]byte{indexRecord, source}, []byte(field)...)
	pfx = append(pfx, 0)
	pfx = append(pfx, []byte(partial)...)
	opts := badger.DefaultIteratorOptions
	opts.PrefetchValues = false
	opts.Prefix = pfx
	err := db.View(func(tx *badger.Txn) error {
		it := tx.NewIterator(opts)
		defer it.Close()
		it.Rewind()
		for it.Seek(pfx); it.ValidForPrefix(pfx); it.Next() {
			k := it.Item().Key()
			start := bytes.IndexByte(k, 0)
			if start == -1 {
				continue
			}
			start++ // skip null byte itself
			end := bytes.IndexByte(k[start:], 0)
			if end == -1 {
				continue
			}

			vmap[string(k[start:start+end])] = struct{}{}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	values := make([]string, 0, len(vmap))
	for v := range vmap {
		values = append(values, v)
	}
	sort.Strings(values)

	return values, nil
}

func WithLocation(ctx context.Context) ([]Photo, error) {
	db := ctx.Value("badger").(*badger.DB)
	photoDir := ctx.Value("photoDir").(string)

	pmap := make(map[string]struct{})
	pfx := []byte{indexRecord, SourceEXIF, 'G', 'P', 'S', 'L', 'a', 't'}
	opts := badger.DefaultIteratorOptions
	opts.PrefetchValues = false
	opts.Prefix = pfx
	err := db.View(func(tx *badger.Txn) error {
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
