package photos

import (
	"encoding/json"
	"time"

	"github.com/dgraph-io/badger"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

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

func writeXMP(log logrus.FieldLogger, db *badger.DB, file string, m Meta) {
	d, t, err := marshalForWrite(m)
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

func writeXMPBatch(log logrus.FieldLogger, batch *badger.WriteBatch, file string, m Meta) {
	d, t, err := marshalForWrite(m)
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



func writeEXIF(log logrus.FieldLogger, db *badger.DB, file string, e map[string]interface{}) {
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
