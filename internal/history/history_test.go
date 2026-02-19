package history

import (
	"context"
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func setupTestDB(t *testing.T) *sql.DB {
	database, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}

	// Create schema
	schema := `
	CREATE TABLE media (
		path TEXT PRIMARY KEY,
		title TEXT,
		duration INTEGER,
		size INTEGER,
		time_created INTEGER,
		time_modified INTEGER,
		time_deleted INTEGER DEFAULT 0,
		time_first_played INTEGER DEFAULT 0,
		time_last_played INTEGER DEFAULT 0,
		play_count INTEGER DEFAULT 0,
		playhead INTEGER DEFAULT 0
	);
	`
	if _, err := database.Exec(schema); err != nil {
		t.Fatal(err)
	}

	// Insert test data
	insert := `INSERT INTO media (path, title, duration, size) VALUES (?, ?, ?, ?)`
	database.Exec(insert, "/test/movie.mp4", "Test Movie", 7200, 1000000)

	return database
}

func TestUpdatePlayback(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	tracker := NewTracker(database)
	ctx := context.Background()

	err := tracker.UpdatePlayback(ctx, "/test/movie.mp4", 3600)
	if err != nil {
		t.Fatalf("UpdatePlayback failed: %v", err)
	}

	// Verify update
	var playCount int64
	var playhead int64
	err = database.QueryRow("SELECT play_count, playhead FROM media WHERE path = ?", "/test/movie.mp4").
		Scan(&playCount, &playhead)
	if err != nil {
		t.Fatal(err)
	}

	if playCount != 1 {
		t.Errorf("Expected play_count 1, got %d", playCount)
	}

	if playhead != 3600 {
		t.Errorf("Expected playhead 3600, got %d", playhead)
	}

	// Update again
	err = tracker.UpdatePlayback(ctx, "/test/movie.mp4", 7200)
	if err != nil {
		t.Fatal(err)
	}

	database.QueryRow("SELECT play_count FROM media WHERE path = ?", "/test/movie.mp4").Scan(&playCount)

	if playCount != 2 {
		t.Errorf("Expected play_count 2, got %d", playCount)
	}
}

func TestMarkDeleted(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	tracker := NewTracker(database)
	ctx := context.Background()

	err := tracker.MarkDeleted(ctx, "/test/movie.mp4")
	if err != nil {
		t.Fatalf("MarkDeleted failed: %v", err)
	}

	var timeDeleted int64
	err = database.QueryRow("SELECT time_deleted FROM media WHERE path = ?", "/test/movie.mp4").
		Scan(&timeDeleted)
	if err != nil {
		t.Fatal(err)
	}

	if timeDeleted == 0 {
		t.Error("Expected time_deleted to be set")
	}
}
