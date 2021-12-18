package database

import (
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"testing"
	"vincit.fi/image-sorter/backend/dbapi"
)

func TestDatabase_InitializeForDirectory(t *testing.T) {
	a := require.New(t)

	sut := NewDatabase()

	dir, err := ioutil.TempDir("", "test_dir")
	a.Nil(err)

	err = sut.InitializeForDirectory(dir, "test.db")
	a.Nil(err)

	err = sut.session.Ping()
	a.Nil(err)

	sut.Close()
}

func TestDatabase_MigrateDB(t *testing.T) {
	a := require.New(t)

	sut := NewDatabase()

	dir, err := ioutil.TempDir("", "test_dir")
	a.Nil(err)

	err = sut.InitializeForDirectory(dir, "test.db")
	a.Nil(err)

	t.Run("First migration", func(t *testing.T) {
		a.Equal(dbapi.TableNotExist, sut.Migrate())
	})
	t.Run("Second migration", func(t *testing.T) {
		a.Equal(dbapi.TableExists, sut.Migrate())
	})

	sut.Close()
}
