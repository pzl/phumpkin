package photos

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/dgraph-io/badger"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// returns the modification time for an EXIF or XMP entry
func ReadModTime(db *badger.DB, key string, file string) (time.Time, error) {
	if file[0:1] == "/" {
		return time.Time{}, fmt.Errorf("photoDir-relative path expected. Got %s in ReadModTime", file)
	}
	t := time.Time{}
	err := db.View(func(tx *badger.Txn) error {
		tm, err := getAsTime(tx, []byte(file+"."+key+".time"))
		if err != nil {
			return err
		}
		t = tm
		return nil
	})
	if err != nil {
		return time.Time{}, err
	}
	return t, nil
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

func fetchJSON(db *badger.DB, key string, v interface{}) error {
	return db.View(func(tx *badger.Txn) error {
		if data, err := getValue(tx, []byte(key)); err != nil {
			return err
		} else {
			if err := json.Unmarshal(data, v); err != nil {
				return err
			}
		}
		return nil
	})
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

type EntrySetter interface {
	SetEntry(*badger.Entry) error
}

func writeEntries(e EntrySetter, key string, d []byte, t []byte) error {
	err := e.SetEntry(badger.NewEntry([]byte(key), d).WithDiscard())
	if err != nil {
		return err
	}
	err = e.SetEntry(badger.NewEntry([]byte(key+".time"), t).WithDiscard())
	if err != nil {
		return err
	}
	return nil
}

/* ----------- XMP ------------- */

func readXMPDB(db *badger.DB, file string) (XMP, error) {
	if file[0:1] == "/" {
		return XMP{}, fmt.Errorf("photoDir-relative path expected. Got %s when calling readXMPDB", file)
	}
	x := XMP{}
	if err := fetchJSON(db, file+".XMP", &x); err != nil {
		return XMP{}, err
	}
	return x, nil
}

func writeXMP(log logrus.FieldLogger, db *badger.DB, file string, x XMP) {
	if file[0:1] == "/" {
		log.WithField("file", file).Error("photoDir-relative path expected")
	}
	d, t, err := marshalForWrite(x)
	if err != nil {
		log.WithError(err).Error("error preparing XMP for write. Unable to write to DB")
		return
	}

	log.WithField("file", file).Trace("writing XMP to db")
	err = db.Update(func(tx *badger.Txn) error { return writeEntries(tx, file+".XMP", d, t) })
	if err != nil {
		log.WithError(err).Error("failed writing XMP to db")
	} else {
		log.WithField("file", file).Trace("wrote XMP data to db")
	}
}

func writeXMPBatch(log logrus.FieldLogger, batch *badger.WriteBatch, file string, x XMP) {
	if file[0:1] == "/" {
		log.WithField("file", file).Error("photoDir-relative path expected")
	}
	d, t, err := marshalForWrite(x)
	if err != nil {
		log.WithError(err).Error("unable to prep XMP for write. Not writing to batch")
		return
	}

	log.WithField("file", file).Trace("writing XMP to db-batch")
	err = writeEntries(batch, file+".XMP", d, t)
	if err != nil {
		log.WithError(err).Error("unable to add XMP record to batch")
		return
	}
	log.WithField("file", file).Trace("wrote XMP data to batch")
}

/* ------------ EXIF ---------- */

func readExifDB(db *badger.DB, file string) (map[string]interface{}, error) {
	if file[0:1] == "/" {
		return nil, fmt.Errorf("photoDir-relative path expected. Got %s when calling readExifDB", file)
	}
	ex := make(map[string]interface{})

	if err := fetchJSON(db, file+".EXIF", &ex); err != nil {
		return nil, err
	}
	return ex, nil
}

func writeEXIF(log logrus.FieldLogger, db *badger.DB, file string, e map[string]interface{}) {
	if file[0:1] == "/" {
		log.WithField("file", file).Error("photoDir-relative path expected")
	}
	d, t, err := marshalForWrite(e)
	if err != nil {
		log.WithError(err).Error("unable to prep EXIF data for writing")
		return
	}

	log.WithField("file", file).Trace("writing exif to db")
	err = db.Update(func(tx *badger.Txn) error { return writeEntries(tx, file+".EXIF", d, t) })
	if err != nil {
		log.WithError(err).Error("failed writing EXIF to db")
	} else {
		log.WithField("file", file).Trace("wrote EXIF to db")
	}
}

func writeEXIFBatch(log logrus.FieldLogger, batch *badger.WriteBatch, file string, e map[string]interface{}) {
	if file[0:1] == "/" {
		log.WithField("file", file).Error("photoDir-relative path expected")
	}
	d, t, err := marshalForWrite(e)
	if err != nil {
		log.WithError(err).Error("unable to prep EXIF data for writing to batch")
		return
	}

	log.WithField("file", file).Trace("writing EXIF to db-batch")
	err = writeEntries(batch, file+".EXIF", d, t)
	if err != nil {
		log.WithError(err).Error("unable to add EXIF record to batch")
		return
	}
	log.WithField("file", file).Trace("wrote EXIF data to batch")
}
