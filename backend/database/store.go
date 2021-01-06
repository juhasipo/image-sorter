package database

import (
	"github.com/upper/db/v4"
)

type Store struct {
	database                *Database
	imageCollection         db.Collection
	categoryCollection      db.Collection
	imageCategoryCollection db.Collection
	imageSimilarCollection  db.Collection
}

func NewStore(database *Database) *Store {
	return &Store{
		database:                database,
		imageCollection:         database.Session().Collection("image"),
		categoryCollection:      database.Session().Collection("category"),
		imageCategoryCollection: database.Session().Collection("image_category"),
		imageSimilarCollection:  database.Session().Collection("image_similar"),
	}
}

func NewInMemoryStore() *Store {
	memoryDb := NewInMemoryDatabase()
	memoryDb.Migrate()
	memoryStore := NewStore(memoryDb)
	return memoryStore
}

func (s *Store) DoInTransaction(fn func(session db.Session) error) error {
	return s.database.Session().Tx(fn)
}
