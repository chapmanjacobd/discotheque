package commands

import (
	"database/sql"
	"fmt"
	"strings"
)

func runMigrations(db *sql.DB) error {
	// Check if 'title' column exists in 'playlists' table
	rows, err := db.Query("PRAGMA table_info(playlists)")
	if err != nil {
		return err
	}
	defer rows.Close()

	hasTitle := false
	for rows.Next() {
		var cid int
		var name, dtype string
		var notnull, pk int
		var dfltValue any
		if err := rows.Scan(&cid, &name, &dtype, &notnull, &dfltValue, &pk); err != nil {
			return err
		}
		if strings.ToLower(name) == "title" {
			hasTitle = true
			break
		}
	}

	if !hasTitle {
		if _, err := db.Exec("ALTER TABLE playlists ADD COLUMN title TEXT"); err != nil {
			// If it fails because the table doesn't exist yet, that's fine, InitDB will create it
			if !strings.Contains(err.Error(), "no such table") {
				return fmt.Errorf("failed to add title column to playlists: %w", err)
			}
		}
	}

	return nil
}
