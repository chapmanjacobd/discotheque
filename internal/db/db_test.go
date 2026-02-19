package db

import (
	"os"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func TestConnect(t *testing.T) {
	f, err := os.CreateTemp("", "db-test-*.db")
	if err != nil {
		t.Fatal(err)
	}
	dbPath := f.Name()
	f.Close()
	defer os.Remove(dbPath)

	db, err := Connect(dbPath)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		t.Fatalf("Ping failed: %v", err)
	}
}
