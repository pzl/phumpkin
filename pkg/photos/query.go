package photos

import (
	"bytes"
	"context"
	"path/filepath"
	"sort"

	"github.com/dgraph-io/badger"
	"github.com/pzl/mstk/logger"
	"github.com/sahilm/fuzzy"
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

type SearchResults struct {
	Total   int            `json:"total"`
	Results []SearchResult `json:"results"`
}

type SearchResult struct {
	Score   int    `json:"score"`
	Matches []int  `json:"matches"`
	Str     string `json:"str"`
	Photo   Photo  `json:"photo"`
}

func GetNames(ctx context.Context, partial string) (SearchResults, error) {
	db := ctx.Value("badger").(*badger.DB)
	photoDir := ctx.Value("photoDir").(string)

	files := make([]string, 0, 1000)
	pfx := []byte{primaryRecord, SourceEXIF, DataRecord}
	opts := badger.DefaultIteratorOptions
	opts.PrefetchValues = false
	opts.Prefix = pfx
	err := db.View(func(tx *badger.Txn) error {
		it := tx.NewIterator(opts)
		defer it.Close()
		it.Rewind()
		for it.Seek(pfx); it.ValidForPrefix(pfx); it.Next() {

			k := it.Item().Key()

			file := string(k[3:])
			files = append(files, file)
		}

		return nil
	})
	if err != nil {
		return SearchResults{}, err
	}

	matches := fuzzy.Find(partial, files)

	page := min(len(matches), 4)
	results := SearchResults{
		Total:   len(matches),
		Results: make([]SearchResult, page),
	}

	for i := 0; i < page; i++ {
		results.Results[i] = SearchResult{
			Score:   matches[i].Score,
			Matches: matches[i].MatchedIndexes,
			Str:     matches[i].Str,
		}

		p, err := FromSrc(ctx, filepath.Join(photoDir, matches[i].Str))
		if err != nil {
			return SearchResults{}, err
		}
		results.Results[i].Photo = p
	}

	return results, nil
}

func WithLocation(ctx context.Context) ([]Photo, error) {
	log := logger.LogFromCtx(ctx)
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
			log.WithError(err).Error("error converting index result to photo")
			continue
		}
		ps = append(ps, p)
	}

	return ps, nil
}

func ColorLabels(ctx context.Context, labels []string) ([]Photo, error) {
	log := logger.LogFromCtx(ctx)
	db := ctx.Value("badger").(*badger.DB)
	photoDir := ctx.Value("photoDir").(string)

	lmap := make(map[string]struct{}, len(labels))
	for _, l := range labels {
		lmap[l] = struct{}{}
	}

	pmap := make(map[string]struct{})
	pfx := []byte{indexRecord, SourceXMP, 'c', 'o', 'l', 'o', 'r', '_', 'l', 'a', 'b', 'e', 'l', 's', 0}
	opts := badger.DefaultIteratorOptions
	opts.PrefetchValues = false
	opts.Prefix = pfx
	err := db.View(func(tx *badger.Txn) error {
		it := tx.NewIterator(opts)
		defer it.Close()
		it.Rewind()
		for it.Seek(pfx); it.ValidForPrefix(pfx); it.Next() {
			k := it.Item().Key()

			vstart := bytes.IndexByte(k, 0)
			if vstart == -1 {
				continue
			}

			vend := bytes.LastIndexByte(k, 0)
			if vend == -1 {
				continue
			}

			v := string(k[vstart+1 : vend])
			if _, ok := lmap[v]; ok {
				fname := string(k[vend+1:])
				pmap[fname] = struct{}{}
			}
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
			log.WithError(err).Error("error converting index result to photo")
			continue
		}
		ps = append(ps, p)
	}

	return ps, nil
}

func ByTags(ctx context.Context, tags []string) ([]Photo, error) {
	log := logger.LogFromCtx(ctx)
	db := ctx.Value("badger").(*badger.DB)
	photoDir := ctx.Value("photoDir").(string)

	tb := make([][]byte, len(tags))
	for i := range tags {
		tb[i] = []byte(tags[i])
	}

	pmap := make(map[string]struct{})
	pfx := []byte{indexRecord, SourceXMP, 't', 'a', 'g', 's', 0}
	opts := badger.DefaultIteratorOptions
	opts.PrefetchValues = false
	opts.Prefix = pfx
	err := db.View(func(tx *badger.Txn) error {
		it := tx.NewIterator(opts)
		defer it.Close()
		it.Rewind()
	keyLoop:
		for it.Seek(pfx); it.ValidForPrefix(pfx); it.Next() {
			k := it.Item().Key()

			for _, t := range tb {
				pfxVal := append(pfx, t...)
				if bytes.HasPrefix(k, pfxVal) {
					vend := bytes.LastIndexByte(k, 0)
					if vend == -1 {
						continue keyLoop
					}
					pmap[string(k[vend+1:])] = struct{}{}
					continue keyLoop
				}
			}
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
			log.WithError(err).Error("error converting index result to photo")
			continue
		}
		ps = append(ps, p)
	}

	return ps, nil
}

func HasFace(ctx context.Context) ([]Photo, error) {
	log := logger.LogFromCtx(ctx)
	db := ctx.Value("badger").(*badger.DB)
	photoDir := ctx.Value("photoDir").(string)

	pmap := make(map[string]struct{})
	pfx := []byte{indexRecord, SourceEXIF, 'F', 'a', 'c', 'e', 's', 'D', 'e', 't', 'e', 'c', 't', 'e', 'd', 0}
	opts := badger.DefaultIteratorOptions
	opts.PrefetchValues = false
	opts.Prefix = pfx
	err := db.View(func(tx *badger.Txn) error {
		it := tx.NewIterator(opts)
		defer it.Close()
		it.Rewind()
		for it.Seek(pfx); it.ValidForPrefix(pfx); it.Next() {
			k := it.Item().Key()

			if k[len(pfx)] != '0' {
				spl := bytes.LastIndexByte(k, 0)
				if spl == -1 {
					continue
				}

				fname := string(k[spl+1:])
				pmap[fname] = struct{}{}
			}
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
			log.WithError(err).Error("error converting index result to photo")
			continue
		}
		ps = append(ps, p)
	}

	return ps, nil
}

func ByRating(ctx context.Context, ratings []string) ([]Photo, error) {
	log := logger.LogFromCtx(ctx)
	db := ctx.Value("badger").(*badger.DB)
	photoDir := ctx.Value("photoDir").(string)

	rmap := make(map[string]struct{}, len(ratings))
	for _, l := range ratings {
		rmap[l] = struct{}{}
	}

	pmap := make(map[string]struct{})
	pfx := []byte{indexRecord, SourceXMP, 'r', 'a', 't', 'i', 'n', 'g', 0} // @ todo: this is not checking EXIF
	opts := badger.DefaultIteratorOptions
	opts.PrefetchValues = false
	opts.Prefix = pfx
	err := db.View(func(tx *badger.Txn) error {
		it := tx.NewIterator(opts)
		defer it.Close()
		it.Rewind()
		for it.Seek(pfx); it.ValidForPrefix(pfx); it.Next() {
			k := it.Item().Key()

			vstart := bytes.IndexByte(k, 0)
			if vstart == -1 {
				continue
			}

			vend := bytes.LastIndexByte(k, 0)
			if vend == -1 {
				continue
			}

			v := string(k[vstart+1 : vend])
			if _, ok := rmap[v]; ok {
				fname := string(k[vend+1:])
				pmap[fname] = struct{}{}
			}
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
			log.WithError(err).Error("error converting index result to photo")
			continue
		}
		ps = append(ps, p)
	}

	return ps, nil
}
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
