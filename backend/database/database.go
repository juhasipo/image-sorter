package database

import (
	"github.com/upper/db/v4"
	"github.com/upper/db/v4/adapter/sqlite"
	"vincit.fi/image-sorter/common/logger"
)

type Database struct {
	instance db.Session
}

func NewDatabase(file string) *Database {
	logger.Info.Printf("Initializing database %s", file)
	var settings = sqlite.ConnectionURL{
		Database: file,
	}

	session, err := sqlite.Open(settings)
	if err != nil {
		logger.Error.Fatal("Error opening database", err)
	}

	return &Database{
		instance: session,
	}
}

type TableExist bool

const (
	TableNotExist TableExist = false
	TableExists   TableExist = true
)

func (s *Database) Migrate() TableExist {
	logger.Info.Printf("Running migrations")
	// TODO: Actually migrate rather than just creating database
	tablesExists := s.doesTablesExists()

	if !tablesExists {
		logger.Info.Print("Initial databases don't exist. Creating...")
		_, err := s.instance.SQL().Exec(`
			CREATE TABLE migrations (
				id TEXT PRIMARY KEY
			)
		`)
		if err != nil {
			logger.Error.Fatal("Error while creating migration table", err)
		}
	}

	logger.Info.Print("Start migrations...")
	err := s.migrate()
	if err != nil {
		logger.Error.Fatal("Error while running migrations", err)
	}
	logger.Info.Print("All migrations done")

	if tablesExists {
		return TableExists
	} else {
		return TableNotExist
	}
}

func (s *Database) doesTablesExists() bool {
	rows, err := s.instance.SQL().Query(`
		SELECT name FROM sqlite_master WHERE type='table' AND name= 'migrations';
	`)

	if err != nil {
		return false
	}

	defer rows.Close()
	return rows.Next()
}

func (s *Database) Session() db.Session {
	return s.instance
}

func (s *Database) migrate() error {
	return s.instance.Tx(func(session db.Session) error {
		migrationId := "0001"
		logger.Info.Printf("Prepare migration %s", migrationId)

		if statement, err := session.SQL().Prepare(`SELECT count(*) FROM migrations WHERE id = ?`); err != nil {
			return err
		} else {
			numFound := 0
			statement.QueryRow(migrationId).Scan(&numFound)

			if numFound > 0 {
				logger.Info.Printf("Migration %s already run", migrationId)
				return nil
			}
		}

		if statement, err := session.SQL().Prepare(`INSERT INTO migrations (id) VALUES (?)`); err != nil {
			return err
		} else {
			logger.Debug.Printf("Mark %s as run", migrationId)
			if _, err := statement.Exec(migrationId); err != nil {
				return err
			}
		}

		logger.Info.Printf("Running migration %s", migrationId)
		query := `
			CREATE TABLE image (
			    id INTEGER PRIMARY KEY,
			    name TEXT,
			    file_name TEXT,
				directory TEXT,
				byte_size INT,
				
				UNIQUE (directory, file_name),
				UNIQUE (name)
			);

			CREATE TABLE category (
			    id INTEGER PRIMARY KEY,
			    name TEXT,
				sub_path TEXT,
				shortcut INTEGER,
				
				UNIQUE (name)
			);

			CREATE TABLE image_category (
			    image_id INTEGER,
			    category_id INTEGER,
			    operation INT,
			    
			    FOREIGN KEY(image_id) REFERENCES image(id) ON DELETE CASCADE,
			    FOREIGN KEY(category_id) REFERENCES category(id) ON DELETE CASCADE
			);

			CREATE TABLE image_hash (
			    image_id INTEGER,
			    
			    FOREIGN KEY(image_id) REFERENCES image(id) ON DELETE CASCADE
			)
		`

		if _, err := session.SQL().Exec(query); err != nil {
			return err
		}

		logger.Debug.Printf("Commit migration")
		return nil
	})
}

func (s *Database) Close() error {
	logger.Info.Printf("Closing database")
	return s.instance.Close()
}
