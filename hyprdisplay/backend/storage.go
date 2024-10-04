package backend

import (
	"database/sql"
	"errors"
	"fmt"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

const DB_NAME string = "hyprdisplay.db"

func SaveSetup(db *sql.DB, key string, ms []MonitorStatus) error {
	_, err := db.Exec("BEGIN TRANSACTION;", key)

	if err != nil {
		return fmt.Errorf("Error during SaveSetup 1 %w", err)
	}
	_, err = db.Exec(`
INSERT INTO Setup(Key) SELECT ?
WHERE NOT EXISTS(SELECT 1 FROM Setup WHERE Key = ?);`,
		key, key)
	if err != nil {
		_, err := db.Exec("ROLLBACK;", key)
		return fmt.Errorf("Error during SaveSetup 2 %w", err)
	}

	_, err = db.Exec(`
DELETE FROM MonitorSetup WHERE SetupId = (SELECt Id FROM Setup WHERE Key = ?);
DELETE FROM Monitor WHERE NOT eXISTs (SELECt 1 FROM MonitorSetup ms WHERE ms.MonitorId = Id);`,
		key)
	if err != nil {
		_, err := db.Exec("ROLLBACK;", key)
		return fmt.Errorf("Error during SaveSetup 4 %w", err)
	}

	for _, m := range ms {
		_, err = db.Exec(`
INSERT INTO Monitor (
	Name,
	Description,
	Disabled,
	Width,
	Height,
	RefreshRate,
	X,
	Y,
	Scale,
	Transform
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?);
INSERT INTO MonitorSetup (SetupId, MonitorId)
VALUES (
	(SELECT Id FROM setup WHERE Key=?),
	(SELECT max(Id) FROM monitor)
);`,
			m.Name,
			m.Description,
			m.Disabled,
			m.Width,
			m.Height,
			m.RefreshRate,
			m.X,
			m.Y,
			m.Scale,
			m.Transform,
			key)

		if err != nil {
			db.Exec("ROLLBACK;", key)
			return fmt.Errorf("Error during SaveSetup 3 %w", err)
		}
	}

	_, err = db.Exec("COMMIT TRANSACTION;", key)

	return nil //errors.New("SaveSetup not implemented")
}

func FindSetup(db *sql.DB, key string) ([]MonitorStatus, error) {
	rows, err := db.Query(`
SELECT 
	m.Name,
	m.Description,
	m.Disabled,
	m.Width,
	m.Height,
	m.RefreshRate,
	m.X,
	m.Y,
	m.Scale,
	m.Transform
FROM Setup s
JOIN MonitorSetup ms ON s.Id = ms.SetupId
JOIN Monitor m ON m.Id = ms.MonitorId
WHERE s.Key = ?
		`, key)

	if errors.Is(err, sql.ErrNoRows) {
		return []MonitorStatus{}, nil
	} else if err != nil {
		return nil, fmt.Errorf("Error running FindSetup query %w", err)
	}

	ms := make([]MonitorStatus, 0)
	for rows.Next() {
		m := MonitorStatus{}
		err = rows.Scan(
			&m.Name,
			&m.Description,
			&m.Disabled,
			&m.Width,
			&m.Height,
			&m.RefreshRate,
			&m.X,
			&m.Y,
			&m.Scale,
			&m.Transform,
		)
		if err != nil {
			return nil, fmt.Errorf("Error running FindSetup query %w", err)
		}
		ms = append(ms, m)
	}
	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("Error running FindSetup query %w", err)
	}

	return ms, nil
}

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
	X INTEGER,
	Y INTEGER,
	Disabled INTEGER,
	Scale NVARCHAR(50),
	Transform INTEGER
);

CREATE TABLE Setup (
	Id INTEGER PRIMARY KEY AUTOINCREMENT,
	Key TEXT
);

CREATE TABLE MonitorSetup (
	MonitorId BIGINT NOT NULL,
	SetupId BIGINT NOT NULL,
	PRIMARY KEY (MonitorId, SetupId),
	FOREIGN KEY (MonitorId) REFERENCES Monitor(Id),
	FOREIGN KEY (SetupId) REFERENCES Setup(Id)
);
`,
}
