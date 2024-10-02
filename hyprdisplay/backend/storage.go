package backend

import (
	"database/sql"
	"errors"
	"fmt"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

const DB_NAME string = "hyprdisplay.db"

func DefaultDbPath() (string, error) {
	xdgDataHome := os.Getenv("XDG_DATA_HOME")

	if xdgDataHome == "" {
		return "", errors.New("$XDG_DATA_HOME not set")
	}

	if _, err := os.Stat(xdgDataHome); errors.Is(err, os.ErrNotExist) {
		return "", errors.New("$XDG_DATA_HOME not a directory")
	}

	return xdgDataHome + string(os.PathSeparator) + "hyprdisplay", nil
}

func InitDb(path string, dbName string) (*sql.DB, error) {
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		os.MkdirAll(path, 0777)
	}

	dbPath := path + string(os.PathSeparator) + dbName

	var db *sql.DB

	if _, err := os.Stat(dbPath); errors.Is(err, os.ErrNotExist) {
		println(dbPath)
		db, err = sql.Open("sqlite3", dbPath)

		if err != nil {
			return nil, fmt.Errorf("Unable to create DB file %w", err)
		}

		_, err = db.Exec("PRAGMA journal_mode=WAL")

		if err != nil {
			db.Close()
			return nil, fmt.Errorf("Unable to set WAL %w", err)
		}
	} else {
		db, err = sql.Open("sqlite3", dbPath)

		if err != nil {
			return nil, fmt.Errorf("Unable to open DB file %w", err)
		}
	}

	_, err := db.Exec("CREATE TABLE IF NOT EXISTS migrations (id INTEGER PRIMARY KEY)")

	if err != nil {
		return nil, fmt.Errorf("Unable to create 'migrations' table %w", err)
	}

	var lastMigration int
	err = db.QueryRow("SELECT Id FROM migrations ORDER BY Id DESC LIMIT 1").Scan(&lastMigration)

	if errors.Is(err, sql.ErrNoRows) {
		lastMigration = 0
	} else if err != nil {
		return nil, fmt.Errorf("Unable to get last migration %w", err)
	}

	err = runMigrations(db, lastMigration)
	if err != nil {
		return nil, fmt.Errorf("Unable to migrate db to current version %w", err)
	}

	return db, nil
}

func runMigrations(db *sql.DB, from int) error {
	lastMigration := len(migrations)

	for i := from + 1; i <= lastMigration; i++ {
		_, err := db.Exec("BEGIN TRANSACTION;")
		if err != nil {
			db.Exec("ROLLBACK;")
			return fmt.Errorf("Error running migration %d: %w", i, err)
		}

		_, err = db.Exec(migrations[i-1])
		if err != nil {
			db.Exec("ROLLBACK;")
			return fmt.Errorf("Error running migration %d: %w", i, err)
		}

		_, err = db.Exec("INSERT INTO migrations (Id) VALUES (?)", i)
		if err != nil {
			db.Exec("ROLLBACK;")
			return fmt.Errorf("Error running migration %d: %w", i, err)
		}

		_, err = db.Exec("COMMIT;")
		if err != nil {
			db.Exec("ROLLBACK;")
			return fmt.Errorf("Error running migration %d: %w", i, err)
		}
	}

	return nil
}

var migrations []string = []string{
	`
CREATE TABLE Monitor (
	Id INTEGER PRIMARY KEY AUTOINCREMENT,
	Name NVARCHAR(250),
	Description NVARCHAR(250),
	Width INTEGER,
	Height INTEGER,
	RefreshRate NVARCHAR(50),
	Disabled INTEGER
);

CREATE TABLE Setup (
	Id INTEGER PRIMARY KEY AUTOINCREMENT,
	Key TEXT
);

CREATE TABLE MonitorSetup (
	MonitorId BIGINT,
	SetupId BIGINT,
	PRIMARY KEY (MonitorId, SetupId),
	FOREIGN KEY (MonitorId) REFERENCES Monitor(Id),
	FOREIGN KEY (SetupId) REFERENCES Setup(Id)
);
`,
}
