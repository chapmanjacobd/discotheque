//go:build fts5

package db

import (
	"database/sql"
	"fmt"
	"os"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func TestFTSRepair(t *testing.T) {
	// 1. Setup
	f, err := os.CreateTemp("", "fts-repair-*.db")
	if err != nil {
		t.Fatal(err)
	}
	dbPath := f.Name()
	f.Close()
	defer os.Remove(dbPath)

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatal(err)
	}

	setupSQL := `
CREATE TABLE media (
    path TEXT PRIMARY KEY,
    title TEXT,
    time_deleted INTEGER DEFAULT 0
);

CREATE VIRTUAL TABLE media_fts USING fts5(
    path,
    title,
    content='media',
    content_rowid='rowid'
);

CREATE TRIGGER media_ai AFTER INSERT ON media BEGIN
    INSERT INTO media_fts(rowid, path, title)
    VALUES (new.rowid, new.path, new.title);
END;

CREATE TRIGGER media_au AFTER UPDATE ON media BEGIN
    INSERT INTO media_fts(media_fts, rowid, path, title) VALUES('delete', old.rowid, old.path, old.title);
    INSERT INTO media_fts(rowid, path, title) VALUES (new.rowid, new.path, new.title);
END;
`
	if _, err := db.Exec(setupSQL); err != nil {
		t.Fatal(err)
	}

	// Insert data
	for i := 0; i < 100; i++ {
		if _, err := db.Exec("INSERT INTO media (path, title) VALUES (?, ?)",
			fmt.Sprintf("file%d.mp4", i), fmt.Sprintf("Video %d", i)); err != nil {
			t.Fatal(err)
		}
	}
	db.Close()

	// 2. Verify Healthy
	if !isHealthy(dbPath) {
		t.Fatal("Database should be healthy initially")
	}

	// 3. Corrupt
	// We want to corrupt it in a way that triggers "database disk image is malformed"
	// but is still recoverable.
	// We'll try to find a spot that is NOT the header.
	stats, _ := os.Stat(dbPath)
	size := stats.Size()
	t.Logf("Database size: %d", size)

	file, err := os.OpenFile(dbPath, os.O_WRONLY, 0o644)
	if err != nil {
		t.Fatal(err)
	}
	// Write garbage in the middle of the file
	if size > 8192 {
		file.WriteAt([]byte("CORRUPT DATA CORRUPT DATA CORRUPT DATA"), 8192)
	} else {
		file.WriteAt([]byte("CORRUPT DATA CORRUPT DATA CORRUPT DATA"), 4096)
	}
	file.Close()

	// 4. Verify Corrupt
	if isHealthy(dbPath) {
		t.Log("isHealthy did NOT detect corruption in the middle, trying offset 100")
		// If middle didn't work, try hitting after the header
		file, _ = os.OpenFile(dbPath, os.O_WRONLY, 0o644)
		// Try not to destroy the whole schema. Offset 100 is still very early.
		// Let's try offset 1024.
		file.WriteAt([]byte("CORRUPT DATA"), 1024)
		file.Close()
		if isHealthy(dbPath) {
			t.Log("isHealthy still did NOT detect corruption, trying offset 100")
			file, _ = os.OpenFile(dbPath, os.O_WRONLY, 0o644)
			file.WriteAt([]byte("CORRUPT"), 100)
			file.Close()
			if isHealthy(dbPath) {
				t.Fatal("Failed to corrupt database in a way isHealthy detects")
			}
		}
	}

	// 5. Repair
	t.Log("Running Repair...")
	if err := Repair(dbPath); err != nil {
		t.Fatalf("Repair failed: %v", err)
	}

	// 6. Verify Healthy again
	if !isHealthy(dbPath) {
		t.Fatal("Database should be healthy after repair")
	}

	// 7. Verify Data and FTS
	db, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	var title string
	err = db.QueryRow("SELECT title FROM media WHERE path = ?", "file42.mp4").Scan(&title)
	if err != nil {
		t.Errorf("Failed to query media after repair: %v", err)
	}
	if title != "Video 42" {
		t.Errorf("Expected 'Video 42', got %q", title)
	}
	t.Logf("Recovered title: %q", title)

	// Check FTS
	var count int
	err = db.QueryRow("SELECT count(*) FROM media_fts WHERE media_fts MATCH 'Video'").Scan(&count)
	if err != nil {
		t.Errorf("FTS MATCH query failed after repair: %v", err)
	}
	if count != 100 {
		t.Errorf("Expected 100 FTS matches, got %d", count)
	}
	t.Logf("Recovered FTS MATCH count: %d", count)
}
