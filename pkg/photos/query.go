package photos

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/dgraph-io/badger"
)

func WithLocation(ctx context.Context) ([]Photo, error) {
	db := ctx.Value("badger").(*badger.DB)
	photoDir := ctx.Value("photoDir").(string)

	ps := make([]Photo, 0, 20)
	err := db.View(func(tx *badger.Txn) error {
		pfx := []byte{primaryRecord, SourceEXIF, DataRecord}
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false
		opts.Prefix = pfx
		it := tx.NewIterator(opts)
		defer it.Close()
		it.Rewind()
		for it.Seek(pfx); it.ValidForPrefix(pfx); it.Next() {
			item := it.Item()
			ex := make(map[string]interface{})
			if err := item.Value(func(v []byte) error { return json.Unmarshal(v, &ex) }); err != nil {
				return err
			}

			loni, ok := ex["GPSLongitude"]
			if !ok {
				continue
			}
			lon, ok := loni.(string)
			if !ok {
				continue
			}
			lati, ok := ex["GPSLatitude"]
			if !ok {
				continue
			}
			lat, ok := lati.(string)
			if !ok {
				continue
			}

			if strings.TrimSpace(lon) == "" || strings.TrimSpace(lat) == "" {
				continue
			}

			k := item.Key()
			fname := string(k[3:])
			p, err := FromSrc(ctx, photoDir+"/"+fname)
			if err != nil {
				return err
			}
			p.exifRead = true
			p.exif = ex
			ps = append(ps, p)

		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return ps, nil
}
