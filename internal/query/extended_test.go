package query

import (
	"context"
	"database/sql"
	"os"
	"testing"

	"github.com/chapmanjacobd/discotheque/internal/models"
)

func TestFileCountsFiltering(t *testing.T) {
	f, _ := os.CreateTemp("", "episodic-test-*.db")
	dbPath := f.Name()
	f.Close()
	defer os.Remove(dbPath)

	dbConn, _ := sql.Open("sqlite3", dbPath)
	schema := `CREATE TABLE media (path TEXT PRIMARY KEY, time_deleted INTEGER DEFAULT 0, size INTEGER, duration INTEGER, title TEXT, type TEXT, time_created INTEGER, time_modified INTEGER, time_first_played INTEGER, time_last_played INTEGER, play_count INTEGER, playhead INTEGER, album TEXT, artist TEXT, genre TEXT, mood TEXT, bpm INTEGER, key TEXT, decade TEXT, categories TEXT, city TEXT, country TEXT, description TEXT, language TEXT, video_codecs TEXT, audio_codecs TEXT, subtitle_codecs TEXT, width INTEGER, height INTEGER);`
	dbConn.Exec(schema)
	dbConn.Exec("INSERT INTO media (path) VALUES ('/show/s1e1.mp4')")
	dbConn.Exec("INSERT INTO media (path) VALUES ('/show/s1e2.mp4')")
	dbConn.Exec("INSERT INTO media (path) VALUES ('/movie/m1.mp4')")
	dbConn.Close()

	ctx := context.Background()
	dbs := []string{dbPath}

	// Filter for directories with > 1 file
	got, err := MediaQuery(ctx, dbs, models.GlobalFlags{AggregateFlags: models.AggregateFlags{FileCounts: ">1"}})
	if err != nil {
		t.Fatalf("MediaQuery failed: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("Expected 2 results, got %d", len(got))
	}

	// Filter for directories with 1 file (Specials)
	got, err = MediaQuery(ctx, dbs, models.GlobalFlags{AggregateFlags: models.AggregateFlags{FileCounts: "1"}})
	if err != nil {
		t.Fatalf("MediaQuery failed: %v", err)
	}
	if len(got) != 1 {
		t.Errorf("Expected 1 result, got %d", len(got))
	}
	if got[0].Path != "/movie/m1.mp4" {
		t.Errorf("Expected movie file, got %s", got[0].Path)
	}
}

func TestFileCountsMediaQueryCount(t *testing.T) {
	f, _ := os.CreateTemp("", "episodic-count-test-*.db")
	dbPath := f.Name()
	f.Close()
	defer os.Remove(dbPath)

	dbConn, _ := sql.Open("sqlite3", dbPath)
	schema := `CREATE TABLE media (path TEXT PRIMARY KEY, time_deleted INTEGER DEFAULT 0, size INTEGER, duration INTEGER, title TEXT, type TEXT, time_created INTEGER, time_modified INTEGER, time_first_played INTEGER, time_last_played INTEGER, play_count INTEGER, playhead INTEGER, album TEXT, artist TEXT, genre TEXT, mood TEXT, bpm INTEGER, key TEXT, decade TEXT, categories TEXT, city TEXT, country TEXT, description TEXT, language TEXT, video_codecs TEXT, audio_codecs TEXT, subtitle_codecs TEXT, width INTEGER, height INTEGER);`
	dbConn.Exec(schema)
	dbConn.Exec("INSERT INTO media (path) VALUES ('/show/s1e1.mp4')")
	dbConn.Exec("INSERT INTO media (path) VALUES ('/show/s1e2.mp4')")
	dbConn.Exec("INSERT INTO media (path) VALUES ('/show/s1e3.mp4')")
	dbConn.Exec("INSERT INTO media (path) VALUES ('/movie/m1.mp4')")
	dbConn.Close()

	ctx := context.Background()
	dbs := []string{dbPath}

	// Total matching count should be 3 (show files)
	flags := models.GlobalFlags{
		AggregateFlags: models.AggregateFlags{FileCounts: "3"},
		QueryFlags:     models.QueryFlags{Limit: 1},
	}
	count, err := MediaQueryCount(ctx, dbs, flags)
	if err != nil {
		t.Fatalf("MediaQueryCount failed: %v", err)
	}
	if count != 3 {
		t.Errorf("Expected count 3, got %d", count)
	}

	// Verify MediaQuery actually respects the limit
	results, err := MediaQuery(ctx, dbs, flags)
	if err != nil {
		t.Fatalf("MediaQuery failed: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("Expected 1 result due to limit, got %d", len(results))
	}
}

func TestFetchSiblings(t *testing.T) {
	f, _ := os.CreateTemp("", "siblings-test-*.db")
	dbPath := f.Name()
	f.Close()
	defer os.Remove(dbPath)

	dbConn, _ := sql.Open("sqlite3", dbPath)
	schema := `CREATE TABLE media (path TEXT PRIMARY KEY, time_deleted INTEGER DEFAULT 0, size INTEGER, duration INTEGER, title TEXT, type TEXT, time_created INTEGER, time_modified INTEGER, time_first_played INTEGER, time_last_played INTEGER, play_count INTEGER, playhead INTEGER, album TEXT, artist TEXT, genre TEXT, mood TEXT, bpm INTEGER, key TEXT, decade TEXT, categories TEXT, city TEXT, country TEXT, description TEXT, language TEXT, video_codecs TEXT, audio_codecs TEXT, subtitle_codecs TEXT, width INTEGER, height INTEGER);`
	dbConn.Exec(schema)
	dbConn.Exec("INSERT INTO media (path) VALUES ('/dir/file1.mp4')")
	dbConn.Exec("INSERT INTO media (path) VALUES ('/dir/file2.mp4')")
	dbConn.Exec("INSERT INTO media (path) VALUES ('/other/file3.mp4')")
	dbConn.Close()

	media := []models.MediaWithDB{
		{Media: models.Media{Path: "/dir/file1.mp4"}, DB: dbPath},
	}

	// Fetch all siblings in the same directory
	got, err := FetchSiblings(context.Background(), media, models.GlobalFlags{FilterFlags: models.FilterFlags{FetchSiblings: "all"}})
	if err != nil {
		t.Fatalf("FetchSiblings failed: %v", err)
	}

	if len(got) != 2 {
		t.Errorf("Expected 2 siblings, got %d", len(got))
	}
}
