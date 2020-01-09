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

func prepXMP(m Meta) ([]byte, []byte, error) {
	m.EXIF = map[string]interface{}{}

	d, err := json.Marshal(m)
	if err != nil {
		return nil, nil, errors.Wrap(err, "error marshalling XMP to json")
	}

	t, err := time.Now().MarshalBinary()
	if err != nil {
		return nil, nil, errors.Wrap(err, "unable to marshal timestamp for XMP")
	}

	return d, t, nil
}

func writeXMP(log logrus.FieldLogger, db *badger.DB, file string, m Meta) {
	d, t, err := prepXMP(m)
	if err != nil {
		log.WithError(err).Error("error preparing XMP for write. Unable to write to DB")
		return
	}

	log.WithField("file", file).Trace("writing XMP to db")
	err = db.Update(func(tx *badger.Txn) error {
		err := tx.SetEntry(badger.NewEntry([]byte(file+".XMP"), d).WithDiscard())
		if err != nil {
			log.WithError(err).Error("unable to set XMP record to db")
			return err
		}

		err = tx.SetEntry(badger.NewEntry([]byte(file+".XMP.time"), t).WithDiscard())
		if err != nil {
			log.WithError(err).Error("unable to write XMP timestamp to db")
			return err
		}

		return nil
	})

	if err != nil {
		log.WithError(err).Error("failed writing XMP to db")
	} else {
		log.WithField("file", file).Trace("wrote XMP data to db")
	}
}

func writeXMPBatch(log logrus.FieldLogger, batch *badger.WriteBatch, file string, m Meta) {
	d, t, err := prepXMP(m)
	if err != nil {
		log.WithError(err).Error("unable to prep XMP for write. Not writing to batch")
		return
	}

	log.WithField("file", file).Trace("writing XMP to db-batch")
	err = batch.SetEntry(badger.NewEntry([]byte(file+".XMP"), d).WithDiscard())
	if err != nil {
		log.WithError(err).Error("unable to add XMP record to batch")
		return
	}
	err = batch.SetEntry(badger.NewEntry([]byte(file+".XMP.time"), t).WithDiscard())
	if err != nil {
		log.WithError(err).Error("unable to add XMP time record to batch")
		return
	}
	log.WithField("file", file).Trace("wrote XMP data to batch")
}

func prepEXIF(e map[string]interface{}) ([]byte, []byte, error) {
	d, err := json.Marshal(e)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed marshalling exif data to json")
	}

	t, err := time.Now().MarshalBinary()
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed marshalling exif time")
	}

	return d, t, nil
}

func writeEXIF(log logrus.FieldLogger, db *badger.DB, file string, e map[string]interface{}) {
	d, t, err := prepEXIF(e)
	if err != nil {
		log.WithError(err).Error("unable to prep EXIF data for writing")
		return
	}

	log.WithField("file", file).Trace("writing exif to db")
	err = db.Update(func(tx *badger.Txn) error {
		err := tx.SetEntry(badger.NewEntry([]byte(file+".EXIF"), d).WithDiscard())
		if err != nil {
			return err
		}

		err = tx.SetEntry(badger.NewEntry([]byte(file+".EXIF.time"), t).WithDiscard())
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		log.WithError(err).Error("failed writing EXIF to db")
	} else {
		log.WithField("file", file).Trace("wrote EXIF to db")
	}
}

func writeEXIFBatch(log logrus.FieldLogger, batch *badger.WriteBatch, file string, e map[string]interface{}) {
	d, t, err := prepEXIF(e)
	if err != nil {
		log.WithError(err).Error("unable to prep EXIF data for writing to batch")
		return
	}

	log.WithField("file", file).Trace("writing EXIF to db-batch")
	err = batch.SetEntry(badger.NewEntry([]byte(file+".EXIF"), d).WithDiscard())
	if err != nil {
		log.WithError(err).Error("unable to add EXIF record to batch")
		return
	}
	err = batch.SetEntry(badger.NewEntry([]byte(file+".EXIF.time"), t).WithDiscard())
	if err != nil {
		log.WithError(err).Error("unable to add EXIF time record to batch")
		return
	}
	log.WithField("file", file).Trace("wrote EXIF data to batch")
}
