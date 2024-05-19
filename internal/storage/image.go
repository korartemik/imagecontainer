package storage

import (
	"encoding/json"
	"github.com/dgraph-io/badger/v4"
	"tagestest/internal/dto"
	"tagestest/internal/lib/clock"
	"time"
)

type ImageStorage struct {
	db    *badger.DB
	clock clock.Clock
}

type ImageModel struct {
	Data      []byte    `json:"data"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (s *ImageStorage) convertToModel(image *dto.Image) *ImageModel {
	return &ImageModel{
		Data:      image.Data,
		CreatedAt: s.clock.Now(),
		UpdatedAt: s.clock.Now(),
	}
}

func NewImageStorage(db *badger.DB, clk clock.Clock) *ImageStorage {
	return &ImageStorage{db: db, clock: clk}
}

func (s *ImageStorage) Get(name string) (*dto.Image, error) {
	model := ImageModel{}
	err := s.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(name))
		if err != nil {
			return err
		}

		valCopy, err := item.ValueCopy(nil)
		if err != nil {
			return err
		}

		err = json.Unmarshal(valCopy, &model)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}
	return &dto.Image{
		Name: name,
		Data: model.Data,
	}, nil
}

func (s *ImageStorage) Insert(image *dto.Image) error {
	return s.db.Update(func(txn *badger.Txn) error {
		data, err := json.Marshal(s.convertToModel(image))
		if err != nil {
			return err
		}
		err = txn.Set([]byte(image.Name), data)
		return err
	})
}

func (s *ImageStorage) Select() ([]dto.ImageListFormat, error) {
	ans := make([]dto.ImageListFormat, 0)
	err := s.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			k := item.Key()
			err := item.Value(func(v []byte) error {
				model := ImageModel{}
				err := json.Unmarshal(v, &model)
				if err != nil {
					return err
				}

				ans = append(ans, dto.ImageListFormat{
					Name:      string(k),
					CreatedAt: model.CreatedAt,
					UpdatedAt: model.UpdatedAt,
				})
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return ans, nil
}

func (s *ImageStorage) Close() error {
	return s.db.Close()
}
