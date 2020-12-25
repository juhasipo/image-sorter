package database

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"vincit.fi/image-sorter/common/logger"
)

type Database struct {
	instance *sql.DB
}

func NewDatabase(file string) *Database {
	logger.Info.Printf("Initializing database %s", file)
	db, err := sql.Open("sqlite3", file)
	if err != nil {
		logger.Error.Fatal("Error opening database", err)
	}

	return &Database{
		instance: db,
	}
}

func (s *Database) GetInstance() *sql.DB {
	return s.instance
}

func (s *Database) Exec(statement string, args ...interface{}) (sql.Result, error) {
	tx, err := s.instance.Begin()

	if err != nil {
		return nil, err
	}

	stmt, err := tx.Prepare(statement)
	if err != nil {
		_ = tx.Rollback()
		return nil, err
	}
	defer stmt.Close()

	if r, err := stmt.Exec(args); err != nil {
		_ = tx.Rollback()
		return r, err
	} else {
		return r, tx.Commit()
	}
}

func (s *Database) Migrate() {
	logger.Info.Printf("Running migrations")
	// TODO: Actually migrate rather than just creating database
	tablesExists := s.doesTablesExists()

	if !tablesExists {
		logger.Info.Print("Initial databases don't exist. Creating...")
		s.instance.Exec(`
			CREATE TABLE migrations (
				id TEXT PRIMARY KEY
			)
		`)
	}

	logger.Info.Print("Start migrations...")
	err := s.migrate()
	if err != nil {
		logger.Error.Fatal("Error while running migrations", err)
	}
	logger.Info.Print("All migrations done")
}

func (s *Database) doesTablesExists() bool {
	rows, err := s.instance.Query(`
		SELECT name FROM sqlite_master WHERE type='table' AND name= 'migrations';
	`)

	if err != nil {
		return false
	}

	defer rows.Close()
	return rows.Next()
}

func (s *Database) migrate() error {
	if tx, err := s.instance.Begin(); err != nil {
		return err
	} else {
		migrationId := "0001"
		logger.Info.Printf("Prepare migration %s", migrationId)

		if statement, err := tx.Prepare(`SELECT count(*) FROM migrations WHERE id = ?`); err != nil {
			return err
		} else {
			numFound := 0
			statement.QueryRow(migrationId).Scan(&numFound)

			if numFound > 0 {
				logger.Info.Printf("Migration %s already run", migrationId)
				_ = tx.Rollback()
				return nil
			}
		}

		if statement, err := tx.Prepare(`INSERT INTO migrations (id) VALUES (?)`); err != nil {
			_ = tx.Rollback()
			return err
		} else {
			logger.Debug.Printf("Mark %s as run", migrationId)
			defer statement.Close()
			if _, err := statement.Exec(migrationId); err != nil {
				_ = tx.Rollback()
				return err
			}
		}

		logger.Info.Printf("Running migration %s", migrationId)
		query := `
			CREATE TABLE image (
			    id INTEGER PRIMARY KEY,
			    name TEXT,
				absolute_path TEXT
			);

			CREATE TABLE category (
			    id INTEGER PRIMARY KEY,
			    name TEXT,
				sub_path TEXT,
				shortcut INTEGER
			);

			CREATE TABLE image_category (
			    image_id INTEGER,
			    category_id INTEGER,
			    
			    FOREIGN KEY(image_id) REFERENCES image(id),
			    FOREIGN KEY(category_id) REFERENCES category(id)
			);

			CREATE TABLE image_hash (
			    image_id INTEGER,
			    
			    FOREIGN KEY(image_id) REFERENCES image(id)
			)
		`

		if _, err := tx.Exec(query); err != nil {
			_ = tx.Rollback()
			return err
		}

		logger.Debug.Printf("Commit migration")
		return tx.Commit()
	}
}

func (s *Database) Close() error {
	logger.Info.Printf("Closing database")
	return s.instance.Close()
}
