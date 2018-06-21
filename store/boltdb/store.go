package boltdb

import (
	"bytes"
	"encoding/gob"
	"log"
	"time"

	"github.com/ahmadmuzakkir/scrapenews/model"
	"github.com/boltdb/bolt"
	"github.com/getsentry/raven-go"
	"github.com/pkg/errors"
)

const bucket = "news"

type Store struct {
	db *bolt.DB
}

func NewStore() (*Store, error) {
	db, err := bolt.Open("news.db", 0600, nil)
	if err != nil {
		return nil, err
	}

	gob.Register(&model.News{})
	gob.Register(&model.Picture{})

	if err := db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(bucket))
		return err
	}); err != nil {
		db.Close()
		raven.CaptureErrorAndWait(err, map[string]string{"module": "boltdb"})
		return nil, err
	}

	return &Store{
		db: db,
	}, nil
}

func (s *Store) Insert(news []*model.News) error {
	log.Println("Add")
	var data = make(map[string][]byte)
	for _, v := range news {
		buf := &bytes.Buffer{}
		log.Println("News Id: ", v.Id)

		if err := gob.NewEncoder(buf).Encode(v); err != nil {
			err = errors.Wrap(err, "[boltdb] gob.Encode() error")
			raven.CaptureError(err, map[string]string{"module": "boltdb"})
			return err
		}

		data[v.Id] = buf.Bytes()
	}

	err := s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return bolt.ErrBucketNotFound
		}

		for key, v := range data {
			keyBytes := []byte(key)
			// Check if it already exists
			if b.Get(keyBytes) != nil {
				log.Println("already exist")
				continue
			}

			err := b.Put(keyBytes, v)
			if err != nil {
				return errors.Wrap(err, "[boltdb] Add() Put error")
			}
		}

		return nil
	})

	if err != nil {
		return nil
	}

	return nil
}

func (s *Store) GetByKeywords(keywords []string) ([]*model.News, error) {
	return nil, nil
}

func (s *Store) GetAll(from time.Time, until time.Time) ([]*model.News, error) {
	return s.get(from, until, "")
}

func (s *Store) get(from time.Time, until time.Time, providerId string) ([]*model.News, error) {
	//if days == 0 {
	//	// default to 1 day
	//	days -= 1
	//}
	//var after = time.Now().AddDate(0, 0, -days)

	var list = make([]*model.News, 0)

	//var data [][]byte
	//var keys [][]byte
	// The keys of the item that cannot be decoded. We should delete these.
	var deleteKeys [][]byte

	s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return bolt.ErrBucketNotFound
		}

		bs := b.Stats()
		log.Println(bs)
		return nil
	})

	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return bolt.ErrBucketNotFound
		}

		c := b.Cursor()

		for k, v := c.First(); k != nil && v != nil; k, v = c.Next() {

			n, err := s.decode(k, v)
			if err != nil {
				deleteKeys = append(deleteKeys, k)
				continue
			}

			if (!from.IsZero() && n.Datetime.After(from)) || (!until.IsZero() && n.Datetime.Before(until)) {
				continue
			}

			list = append(list, n)

			//data = append(data, v)
			//keys = append(keys, k)
		}
		return nil
	})

	if deleteKeys != nil && len(deleteKeys) > 0 {
		s.db.Update(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte(bucket))

			for _, val := range deleteKeys {
				b.Delete([]byte(val))
			}
			return nil
		})
	}

	if err != nil {
		return nil, err
	}

	//length := len(data)
	//for i := 0; i < length; i++ {
	//	n := &model.News{}
	//
	//	err = gob.NewDecoder(bytes.NewBuffer(data[i])).Decode(n)
	//	if err != nil {
	//		err = errors.Wrap(err, "[boltdb] GetAll() gob.Decode() error")
	//		raven.CaptureError(err, map[string]string{"module": "boltdb"})
	//
	//		// Delete the data if we could not decode it.
	//		s.delete(string(keys[i]))
	//		return nil, err
	//	}
	//
	//	if providerId != "" && n.Source.NewspaperId != providerId {
	//		continue
	//	}
	//
	//	log.Println("GetAll ", n.Title)
	//	list = append(list, n)
	//}

	return list, nil
}

//func (s *Store) get(providerId string) ([]*model.News, error) {
//	var data [][]byte
//	var keys [][]byte
//
//	err := s.db.View(func(tx *bolt.Tx) error {
//		b := tx.Bucket([]byte(bucket))
//		if b == nil {
//			return bolt.ErrBucketNotFound
//		}
//
//		c := b.Cursor()
//
//		for k, v := c.First(); k != nil && v != nil; k, v = c.Next() {
//			data = append(data, v)
//			keys = append(keys, k)
//		}
//		return nil
//	})
//
//	if err != nil {
//		return nil, err
//	}
//
//	var list = make([]*model.News, 0)
//
//	length := len(data)
//	for i := 0; i < length; i++ {
//		n := &model.News{}
//
//		err = gob.NewDecoder(bytes.NewBuffer(data[i])).Decode(n)
//		if err != nil {
//			err = errors.Wrap(err, "[boltdb] GetAll() gob.Decode() error")
//			raven.CaptureError(err, map[string]string{"module": "boltdb"})
//
//			// Delete the data if we could not decode it.
//			s.delete(string(keys[i]))
//			return nil, err
//		}
//
//		if providerId != "" && n.Source.NewspaperId != providerId {
//			continue
//		}
//
//		log.Println("GetAll ", n.Title)
//		list = append(list, n)
//	}
//
//	return list, nil
//}

func (s *Store) GetByProvider(provider string) ([]*model.News, error) {
	return nil, nil
}

//// Create a meta bucket to store the last update datetime.
//func (s *Store) GetLatestNewsTime(providerId string) (time.Time) {
//	list, err := s.get(providerId)
//	if err != nil {
//		return time.Time{}
//	}
//
//	if list == nil || len(list) < 1 {
//		return time.Time{}
//	}
//
//	var latest time.Time
//
//	for _, v := range list {
//		if v.Datetime.After(latest) {
//			latest = v.Datetime
//		}
//	}
//
//	return latest
//}

func (s *Store) delete(key string) error {
	err := s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))

		err := b.Delete([]byte(key))
		if err != nil {
			return err
		}

		return nil
	})
	return err
}

func (s *Store) decode(key []byte, value []byte) (*model.News, error) {
	n := &model.News{}

	err := gob.NewDecoder(bytes.NewBuffer(value)).Decode(n)
	if err != nil {
		err = errors.Wrap(err, "[boltdb] GetAll() gob.Decode() error")
		raven.CaptureError(err, map[string]string{"module": "boltdb"})

		// Delete the data if we could not decode it.
		s.delete(string(key))
		return nil, err
	}
	if err != nil {
		return nil, err
	}

	return n, nil
}
