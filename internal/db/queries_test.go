package db

import (
	"context"
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func TestQueries(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	// Create schema
	_, err = db.Exec(`
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
		playhead INTEGER DEFAULT 0,
		type TEXT,
		width INTEGER,
		height INTEGER,
		fps REAL,
		video_codecs TEXT,
		audio_codecs TEXT,
		subtitle_codecs TEXT,
		video_count INTEGER DEFAULT 0,
		audio_count INTEGER DEFAULT 0,
		subtitle_count INTEGER DEFAULT 0,
		album TEXT,
		artist TEXT,
		genre TEXT,
		mood TEXT,
		bpm INTEGER,
		key TEXT,
		decade TEXT,
		categories TEXT,
		city TEXT,
		country TEXT,
		description TEXT,
		language TEXT,
		webpath TEXT,
		uploader TEXT,
		time_uploaded INTEGER,
		time_downloaded INTEGER,
		view_count INTEGER,
		num_comments INTEGER,
		favorite_count INTEGER,
		score REAL,
		upvote_ratio REAL,
		latitude REAL,
		longitude REAL
	);
	`)
	if err != nil {
		t.Fatal(err)
	}

	q := New(db)
	ctx := context.Background()

	t.Run("UpsertAndGet", func(t *testing.T) {
		err := q.UpsertMedia(ctx, UpsertMediaParams{
			Path:  "test.mp4",
			Title: sql.NullString{String: "Test Title", Valid: true},
			Size:  sql.NullInt64{Int64: 1000, Valid: true},
		})
		if err != nil {
			t.Errorf("UpsertMedia failed: %v", err)
		}

		m, err := q.GetMediaByPathExact(ctx, "test.mp4")
		if err != nil {
			t.Errorf("GetMediaByPathExact failed: %v", err)
		}
		if m.Title.String != "Test Title" {
			t.Errorf("Expected Test Title, got %s", m.Title.String)
		}
	})

	t.Run("CategoryStats", func(t *testing.T) {
		db.Exec("UPDATE media SET categories = ';comedy;' WHERE path = 'test.mp4'")
		stats, err := q.GetCategoryStats(ctx)
		if err != nil {
			t.Fatal(err)
		}

		found := false
		for _, s := range stats {
			if s.Category == "comedy" && s.Count == 1 {
				found = true
				break
			}
		}
		if !found {
			t.Error("Comedy category stat not found")
		}
	})
}
