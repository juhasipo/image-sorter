package database

import (
	"github.com/upper/db/v4"
)

type Store struct {
	database                *Database
}

func NewStore(database *Database) *Store {
	return &Store{
		database:                database,
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
