package photos

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/dgraph-io/badger"
	"github.com/pkg/errors"
	"github.com/pzl/mstk/logger"
)

/*
	DB structure.
	Universal key structure:
		- first byte is recordType
		- second byte is XMP or EXIF source

	Primary data:
	--------------
	key: primaryRecord + <sourceType> + <DataType> + []byte(fileID)

		 x2, one for XMP, one for EXIF (usually)

	value:
		for the non-time records, JSON encoded.
			map[string]interface{} for EXIF, XMP for xmp
		for the time records, a binary marshalled time.Time



	Index Data:
	-------------
	key: indexRecord + <sourceType> + []byte(fieldname) + byte(0) + []byte(value) + ?? separator + []byte(fileID)
	value: []byte{}


	values seem to be one of:
		- string (some of them very long, like AFAreaXPosition)
		- int ( binary.BigEndian.PutUint64(buf[:], i) )
		- float64  ( binary.BigEndian.PutUint64(buf[:], math.Float64bits(f)) )
		- bool? or are these parsed strings. One bool in XMP
		- structs - from XMP. Location, and history Ops
		- []string in XMP -- just have multiple index records with differing values?


	Strategies
	===========

	- Delete all for fileID:
		+ DEL primaryRecord+EXIF+fileID, DEL primaryRecord+XMP+fileID (+times)
		+ DEL .. need to find indexes
	- Get all unique values for a field
		+ prefix = indexRecord+source+field+0, collect all the values up
		+ iteratorOptions.Prefix is a thing (on top of seek, validForPfx, etc)
	- sort by field
		+ if field is stored correctly, should be pre-sorted
		+ store numbers as BigEndian
		+ iterator.reverse is an option
	- get filterable fields
		+ populate dropdown: select EXIF/XMP, then get a dropdown with all keys
		+ typeahead filling, from the above, or seeking

*/

const (
	primaryRecord byte = iota + 1
	indexRecord
)

const (
	SourceEXIF byte = iota + 1
	SourceXMP
)

const (
	DataRecord byte = iota + 1
	TimestampRecord
)

// key helpers
func TimeKey(file string, source byte) []byte {
	return append([]byte{primaryRecord, source, TimestampRecord}, []byte(file)...)
}
func DataKey(file string, source byte) []byte {
	return append([]byte{primaryRecord, source, DataRecord}, []byte(file)...)
}

func Read(ctx context.Context, key []byte, into interface{}) error {
	warnIfAbsolute(ctx, key[3])
	db, err := dbHandle(ctx)
	if err != nil {
		return err
	}
	return fetchJSON(db, key, into)
}

func ReadTime(ctx context.Context, key []byte) (time.Time, error) {
	var t time.Time
	warnIfAbsolute(ctx, key[3])
	db, err := dbHandle(ctx)
	if err != nil {
		return t, err
	}
	return t, db.View(func(tx *badger.Txn) error {
		tv, err := tx.Get(key)
		if err != nil {
			return err
		}
		return tv.Value(func(b []byte) error {
			return t.UnmarshalBinary(b)
		})
	})
}

func Write(ctx context.Context, sourceType byte, file string, data interface{}, batch *badger.WriteBatch) error {
	warnIfAbsolute(ctx, []byte(file)[0])
	db, err := dbHandle(ctx)
	if err != nil {
		return err
	}

	d, tm, err := marshalForWrite(data)
	if err != nil {
		return fmt.Errorf("database Write -- unable to marshal data for file %s: %w", file, err)
	}

	key := DataKey(file, sourceType)
	if batch != nil {
		return writeRecords(batch, d, tm, key)
	}
	return db.Update(func(tx *badger.Txn) error {
		return writeRecords(tx, d, tm, key)
	})
}

func WriteIdxField(ctx context.Context, sourceType byte, file string, field string, v []byte, batch *badger.WriteBatch) error {
	warnIfAbsolute(ctx, []byte(file)[0])
	db, err := dbHandle(ctx)
	if err != nil {
		return err
	}

	fileb := []byte(file)
	fieldb := []byte(field)

	key := make([]byte, len(fileb)+len(fieldb)+len(v)+4) // extras are: recordType, sourceType, and 2*null-sep
	key[0] = indexRecord
	key[1] = sourceType
	copy(key[2:], fieldb)
	key[2+len(fieldb)] = 0
	copy(key[3+len(fieldb):], v)
	key[3+len(fieldb)+len(v)] = 0 // ?? separator v could contain ANYTHING
	copy(key[4+len(fieldb)+len(v):], fileb)

	if batch != nil {
		return batch.SetEntry(badger.NewEntry(key, nil).WithDiscard())
	}
	return db.Update(func(tx *badger.Txn) error {
		return tx.SetEntry(badger.NewEntry(key, nil).WithDiscard())
	})

}

/* ---- write helpers ----------- */

type EntrySetter interface {
	SetEntry(*badger.Entry) error
}

func writeRecords(e EntrySetter, d []byte, t []byte, k []byte) error {
	if err := e.SetEntry(badger.NewEntry(k, d).WithDiscard()); err != nil {
		return err
	}
	kt := make([]byte, len(k))
	copy(kt, k)
	kt[2] = TimestampRecord
	return e.SetEntry(badger.NewEntry(kt, t).WithDiscard())
}

func marshalForWrite(v interface{}) ([]byte, []byte, error) {
	d, err := json.Marshal(v)
	if err != nil {
		return nil, nil, errors.Wrap(err, "error marshalling data to json")
	}

	t, err := time.Now().MarshalBinary()
	if err != nil {
		return nil, nil, errors.Wrap(err, "unable to marshal timestamp for writing db record time")
	}

	return d, t, nil
}

/* ---- Read helpers ----- */

func getValue(tx *badger.Txn, key []byte) ([]byte, error) {
	i, err := tx.Get(key)
	if err != nil {
		return nil, err
	}
	v, err := i.ValueCopy(nil)
	if err != nil {
		return nil, err
	}
	return v, nil
}

func fetchJSON(db *badger.DB, key []byte, v interface{}) error {
	return db.View(func(tx *badger.Txn) error {
		if data, err := getValue(tx, key); err != nil {
			return err
		} else {
			if err := json.Unmarshal(data, v); err != nil {
				return err
			}
		}
		return nil
	})
}

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

// generic helpers

func dbHandle(ctx context.Context) (*badger.DB, error) {
	b := ctx.Value("badger")
	if b == nil {
		return nil, errors.New("db not present in context")
	}
	db, ok := b.(*badger.DB)
	if !ok || db == nil {
		return nil, errors.New("unable to parse database from context")
	}
	return db, nil
}

func warnIfAbsolute(ctx context.Context, key byte) {
	log := logger.LogFromCtx(ctx)
	if key == '/' {
		log.WithField("key", key).Warn("photoDir-relative path expected")
	}
}

// fetches a key that's expected to be a time.time
func getAsTime(tx *badger.Txn, key []byte) (time.Time, error) {
	t, err := tx.Get(key)
	if err != nil {
		return time.Time{}, err
	}

	tm := time.Time{}
	v, err := t.ValueCopy(nil)
	if err != nil {
		return time.Time{}, err
	}
	if err := tm.UnmarshalBinary(v); err != nil {
		return time.Time{}, err
	}

	return tm, nil
}
