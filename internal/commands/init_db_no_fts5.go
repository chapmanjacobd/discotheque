//go:build !fts5

package commands

import (
	"database/sql"
	"strings"
)

func InitDB(sqlDB *sql.DB) error {
	schemaBytes, err := schemaFS.ReadFile("schema.sql")
	if err != nil {
		return err
	}
	schema := string(schemaBytes)

	// Filter out FTS5 specific commands
	var filteredSchema strings.Builder
	lines := strings.SplitSeq(schema, ";")
	for line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if strings.Contains(strings.ToUpper(trimmed), "FTS5") || strings.Contains(strings.ToUpper(trimmed), "_FTS") {
			continue
		}
		filteredSchema.WriteString(trimmed)
		filteredSchema.WriteString(";")
	}

	tx, err := sqlDB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err = tx.Exec(filteredSchema.String()); err != nil {
		return err
	}

	return tx.Commit()
}
