package db

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

// Connect opens a SQLite database and applies performance tuning PRAGMAs
func Connect(dbPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	// Performance Tuning
	tuning := []string{
		"PRAGMA journal_mode=WAL",
		"PRAGMA synchronous=NORMAL",
		"PRAGMA cache_size=-64000",
		"PRAGMA temp_store=MEMORY",
		"PRAGMA foreign_keys=ON",
		"PRAGMA threads=4",
	}

	for _, pragma := range tuning {
		if _, err := db.Exec(pragma); err != nil {
			db.Close()
			return nil, fmt.Errorf("failed to apply pragma %q: %w", pragma, err)
		}
	}

	// Connection Pool Limits
	// SQLite handles concurrent reads well in WAL mode, but concurrent writes
	// can lead to "database is locked" errors. Limiting to 1 open connection
	// ensures serialization and avoids many common SQLite concurrency issues.
	db.SetMaxOpenConns(1)

	return db, nil
}
