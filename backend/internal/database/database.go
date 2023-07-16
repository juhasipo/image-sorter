package database

import (
	"github.com/upper/db/v4"
	"github.com/upper/db/v4/adapter/sqlite"
	"path/filepath"
	"vincit.fi/image-sorter/backend/dbapi"
	"vincit.fi/image-sorter/backend/internal/util"
	"vincit.fi/image-sorter/common/constants"
	"vincit.fi/image-sorter/common/logger"
)

type Database struct {
	session  db.Session
	dbPath   string
	basePath string
}

func NewInMemoryDatabase(basePath string) *Database {
	logger.Info.Printf("Initializing in-memory database")
	var settings = sqlite.ConnectionURL{
		Database: "memory.db",
		Options: map[string]string{
			"mode": "memory",
		},
	}

	session, err := sqlite.Open(settings)
	if err != nil {
		logger.Error.Fatal("Error opening database ", err)
	}

	database := Database{session: session, basePath: basePath}

	database.Migrate()

	return &database
}

func NewDatabase() *Database {
	return &Database{}
}

func (s *Database) BasePath() string {
	return s.basePath
}

func (s *Database) InitializeForDirectory(directory string, file string) error {
	if err := util.MakeDirectoriesIfNotExist(directory, filepath.Join(directory, constants.ImageSorterDir)); err != nil {
		return err
	}
	s.basePath = directory

	s.dbPath = filepath.Join(directory, constants.ImageSorterDir, file)
	logger.Info.Printf("Initializing database %s", s.dbPath)
	var settings = sqlite.ConnectionURL{
		Database: s.dbPath,
	}

	session, err := sqlite.Open(settings)
	if err != nil {
		return err
	}

	s.session = session
	// Get database engine version from session
	var version map[string]interface{}
	err = s.session.SQL().Select(db.Func("sqlite_version")).One(&version)

	logger.Info.Printf("Database initialized. Using SQLite version %s", version["sqlite_version()"])

	return nil
}

func (s *Database) Migrate() dbapi.TableExist {
	logger.Info.Printf("Running migrations")
	tablesExists := s.doesTablesExists()

	if !tablesExists {
		logger.Info.Print("Initial databases don't exist. Creating...")
		err := s.session.Tx(func(session db.Session) error {
			_, err := session.SQL().Exec(`
				CREATE TABLE migration (
					id TEXT PRIMARY KEY
				)
			`)
			if err != nil {
				logger.Error.Fatal("Error while creating migration table", err)
			}
			return err
		})

		if err != nil {
			logger.Error.Fatal("Error while running migrations", err)
		}

	}

	logger.Info.Print("Start migrations...")
	err := s.migrate()
	if err != nil {
		logger.Error.Fatal("Error while running migrations", err)
	}
	logger.Info.Print("All migrations done")

	if tablesExists {
		return dbapi.TableExists
	} else {
		return dbapi.TableNotExist
	}
}

func (s *Database) doesTablesExists() bool {
	rows, err := s.session.SQL().Query(`
		SELECT name FROM sqlite_master WHERE type='table' AND name= 'migration';
	`)

	if err != nil {
		return false
	}

	defer rows.Close()
	return rows.Next()
}

func (s *Database) Session() db.Session {
	return s.session
}

func (s *Database) migrate() error {
	return s.session.Tx(func(session db.Session) error {
		if migrationStatusesById, err := s.findAlreadyRunMigrations(session); err != nil {
			return err
		} else {
			for _, migration := range migrations {
				if err := s.runMigration(session, migration, migrationStatusesById); err != nil {
					logger.Error.Print("Print failed to run migration ", err)
					return err
				}
			}

			logger.Debug.Printf("Commit migrations")
			return nil
		}
	})
}

func (s *Database) runMigration(session db.Session, migration migration, migrationStatusesById map[MigrationId]bool) error {
	migrationId := migration.id

	logger.Info.Printf("Prepare migration %d: %s", migrationId, migration.description)

	if _, found := migrationStatusesById[migrationId]; found {
		logger.Info.Printf("Migration %d is already done", migrationId)
	} else {
		if statement, err := session.SQL().Prepare(`INSERT INTO migration (id) VALUES (?)`); err != nil {
			return err
		} else {
			logger.Debug.Printf("Mark %d as run", migrationId)
			if _, err := statement.Exec(migrationId); err != nil {
				return err
			}
		}

		logger.Info.Printf("Running migration %d", migration.id)

		if _, err := session.SQL().Exec(migration.query); err != nil {
			return err
		}
	}
	return nil
}

func (s *Database) findAlreadyRunMigrations(session db.Session) (map[MigrationId]bool, error) {
	var runMigrationIds []Migration
	if err := session.Collection("migration").Find().All(&runMigrationIds); err != nil {
		return nil, err
	} else {
		var migrationStatusesById = map[MigrationId]bool{}
		for _, migration := range runMigrationIds {
			migrationStatusesById[migration.Id] = true
		}
		return migrationStatusesById, nil
	}
}

func (s *Database) Close() {
	logger.Info.Printf("Closing database %s", s.dbPath)
	if s.session != nil {
		if err := s.session.Close(); err != nil {
			logger.Error.Print("Error while trying to close database ", err)
		}
	} else {
		logger.Warn.Printf("No database instance to close")
	}
}
