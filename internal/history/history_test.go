package history

import (
	"context"
	"database/sql"
	"os"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func setupTestDB(t *testing.T) (*sql.DB, string) {
	f, err := os.CreateTemp("", "history-test-*.db")
	if err != nil {
		t.Fatal(err)
	}
	dbPath := f.Name()
	f.Close()

	sqlDB, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		os.Remove(dbPath)
		t.Fatal(err)
	}

	schema := `
	CREATE TABLE media (
		path TEXT PRIMARY KEY,
		time_first_played INTEGER,
		time_last_played INTEGER,
		playhead INTEGER,
		play_count INTEGER DEFAULT 0,
		time_deleted INTEGER DEFAULT 0
	);
	CREATE TABLE history (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		media_path TEXT NOT NULL,
		time_played INTEGER,
		playhead INTEGER,
		done INTEGER
	);
	`
	if _, err := sqlDB.Exec(schema); err != nil {
		sqlDB.Close()
		os.Remove(dbPath)
		t.Fatal(err)
	}

	return sqlDB, dbPath
}

func TestTracker_UpdatePlayback(t *testing.T) {
	sqlDB, dbPath := setupTestDB(t)
	defer sqlDB.Close()
	defer os.Remove(dbPath)

	path := "/test/video.mp4"
	if _, err := sqlDB.Exec("INSERT INTO media (path) VALUES (?)", path); err != nil {
		t.Fatal(err)
	}

	tracker := NewTracker(sqlDB)
	if err := tracker.UpdatePlayback(context.Background(), path, 100); err != nil {
		t.Fatalf("UpdatePlayback failed: %v", err)
	}

	// Verify media update
	var lastPlayed, playhead int64
	err := sqlDB.QueryRow("SELECT time_last_played, playhead FROM media WHERE path = ?", path).Scan(&lastPlayed, &playhead)
	if err != nil {
		t.Fatalf("Failed to query media: %v", err)
	}
	if lastPlayed == 0 {
		t.Error("Expected time_last_played to be set")
	}
	if playhead != 100 {
		t.Errorf("Expected playhead 100, got %d", playhead)
	}

	// Verify history record
	var hPath string
	var hPlayhead int64
	err = sqlDB.QueryRow("SELECT media_path, playhead FROM history WHERE media_path = ?", path).Scan(&hPath, &hPlayhead)
	if err != nil {
		t.Fatalf("Failed to query history: %v", err)
	}
	if hPath != path {
		t.Errorf("Expected history path %s, got %s", path, hPath)
	}
	if hPlayhead != 100 {
		t.Errorf("Expected history playhead 100, got %d", hPlayhead)
	}
}

func TestUpdateHistorySimple(t *testing.T) {
	sqlDB, dbPath := setupTestDB(t)
	sqlDB.Close() // UpdateHistorySimple opens its own connection
	defer os.Remove(dbPath)

	// Re-open to insert test data
	dbConn, _ := sql.Open("sqlite3", dbPath)
	path := "/test/audio.mp3"
	dbConn.Exec("INSERT INTO media (path) VALUES (?)", path)
	dbConn.Close()

	if err := UpdateHistorySimple(dbPath, []string{path}, 50, true); err != nil {
		t.Fatalf("UpdateHistorySimple failed: %v", err)
	}

	// Verify
	dbConn, _ = sql.Open("sqlite3", dbPath)
	defer dbConn.Close()

	var done int64
	err := dbConn.QueryRow("SELECT done FROM history WHERE media_path = ?", path).Scan(&done)
	if err != nil {
		t.Fatalf("Failed to query history: %v", err)
	}
	if done != 1 {
		t.Errorf("Expected done=1, got %d", done)
	}
}
