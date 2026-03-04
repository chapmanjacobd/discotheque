package commands

import (
	"database/sql"
	"fmt"
	"strings"
)

func runMigrations(db *sql.DB) error {
	// 1. Column migrations
	if err := migrateColumns(db); err != nil {
		return err
	}

	// 2. Table migrations
	if err := migrateTables(db); err != nil {
		return err
	}

	// 3. Index migrations
	if err := migrateIndexes(db); err != nil {
		return err
	}

	return nil
}

func migrateColumns(db *sql.DB) error {
	cols := []struct {
		table  string
		column string
		schema string
	}{
		{"playlists", "title", "TEXT"},
		{"playlist_items", "time_added", "INTEGER DEFAULT 0"},
	}

	for _, c := range cols {
		rows, err := db.Query(fmt.Sprintf("PRAGMA table_info(%s)", c.table))
		if err != nil {
			if strings.Contains(err.Error(), "no such table") {
				continue
			}
			return err
		}

		exists := false
		for rows.Next() {
			var cid int
			var name, dtype string
			var notnull, pk int
			var dfltValue any
			if err := rows.Scan(&cid, &name, &dtype, &notnull, &dfltValue, &pk); err != nil {
				rows.Close()
				return err
			}
			if strings.EqualFold(name, c.column) {
				exists = true
				break
			}
		}
		rows.Close()

		if !exists {
			if _, err := db.Exec(fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s", c.table, c.column, c.schema)); err != nil {
				if !strings.Contains(err.Error(), "no such table") {
					return fmt.Errorf("failed to add column %s to table %s: %w", c.column, c.table, err)
				}
			}
		}
	}
	return nil
}

func migrateTables(db *sql.DB) error {
	// Create custom_keywords table if it doesn't exist
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS custom_keywords (
		category TEXT NOT NULL,
		keyword TEXT NOT NULL,
		PRIMARY KEY (category, keyword)
	)`); err != nil {
		return fmt.Errorf("failed to create custom_keywords table: %w", err)
	}
	return nil
}

func migrateIndexes(db *sql.DB) error {
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_type ON media(type)",
		"CREATE INDEX IF NOT EXISTS idx_genre ON media(genre)",
		"CREATE INDEX IF NOT EXISTS idx_artist ON media(artist)",
		"CREATE INDEX IF NOT EXISTS idx_album ON media(album)",
		"CREATE INDEX IF NOT EXISTS idx_categories ON media(categories)",
		"CREATE INDEX IF NOT EXISTS idx_uploader ON media(uploader)",
		"CREATE INDEX IF NOT EXISTS idx_score ON media(score)",
		"CREATE INDEX IF NOT EXISTS idx_view_count ON media(view_count)",
		"CREATE INDEX IF NOT EXISTS idx_time_created ON media(time_created)",
		"CREATE INDEX IF NOT EXISTS idx_time_modified ON media(time_modified)",
		"CREATE INDEX IF NOT EXISTS idx_time_uploaded ON media(time_uploaded)",
		"CREATE INDEX IF NOT EXISTS idx_time_downloaded ON media(time_downloaded)",
	}

	for _, idx := range indexes {
		if _, err := db.Exec(idx); err != nil {
			return fmt.Errorf("failed to create index: %w", err)
		}
	}
	return nil
}
