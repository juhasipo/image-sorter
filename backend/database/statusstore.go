package database

import (
	"github.com/upper/db/v4"
	"time"
	"vincit.fi/image-sorter/common/logger"
)

type StatusKey string

const (
	SimilarityIndexUpdated StatusKey = "similarity_index_updated"
	ImageIndexUpdated      StatusKey = "image_index_updated"
)

type StatusStore struct {
	database   *Database
	collection db.Collection
}

func NewStatusStore(database *Database) *StatusStore {
	return &StatusStore{
		database: database,
	}
}

func (s *StatusStore) getCollection() db.Collection {
	if s.collection == nil {
		s.collection = s.database.Session().Collection("status")
	}
	return s.collection
}

func (s *StatusStore) GetStatus(key StatusKey) (*Status, error) {
	var status Status
	if err := s.getCollection().Find(db.Cond{"key": key}).One(&status); err != nil {
		return nil, err
	} else {
		return &status, nil
	}
}

func (s *StatusStore) UpdateTimestamp(key StatusKey, timestamp time.Time) error {
	logger.Debug.Printf("Updating %s to %s", key, timestamp)
	return s.getCollection().Find(db.Cond{"key": key}).Update(&Status{
		Key:       key,
		Timestamp: timestamp,
	})
}
