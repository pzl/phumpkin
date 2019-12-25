package photos

import (
	"encoding/json"
	"time"

	"github.com/dgraph-io/badger"
	"github.com/sirupsen/logrus"
)

func getModTime(tx *badger.Txn, key []byte) (*time.Time, error) {
	t, err := tx.Get(key)
	if err != nil {
		return nil, err
	}

	tm := &time.Time{}
	v, err := t.ValueCopy(nil)
	if err != nil {
		return nil, err
	}
	if err := tm.UnmarshalBinary(v); err != nil {
		return nil, err
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

func writeXMP(log logrus.FieldLogger, db *badger.DB, file string, m Meta) {
	m.EXIF = map[string]interface{}{}

	d, err := json.Marshal(m)
	if err != nil {
		log.WithError(err).Error("error marshalling XMP to json. unable to write to DB")
		return
	}

	t, err := time.Now().MarshalBinary()
	if err != nil {
		log.WithError(err).Error("unable to marshal timestamp for XMP. unable to write to DB")
		return
	}

	log.Info("writing XMP to db")
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
		log.Info("wrote XMP data to db")
	}
}

func writeEXIF(log logrus.FieldLogger, db *badger.DB, file string, e map[string]interface{}) {
	d, err := json.Marshal(e)
	if err != nil {
		log.WithError(err).Error("failed marshalling exif data to json. cannot write to db")
		return
	}

	t, err := time.Now().MarshalBinary()
	if err != nil {
		log.WithError(err).Error("failed marshalling exif time to json. cannot write to db")
		return
	}

	log.Info("writing exif to db")
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
		log.Info("wrote EXIF to db")
	}
}
