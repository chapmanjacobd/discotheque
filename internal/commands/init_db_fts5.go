//go:build fts5

package commands

import (
	"database/sql"
)

func InitDB(sqlDB *sql.DB) error {
	schema, err := schemaFS.ReadFile("schema.sql")
	if err != nil {
		return err
	}

	tx, err := sqlDB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(string(schema)); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return runMigrations(sqlDB)
}
