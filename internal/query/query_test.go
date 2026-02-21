package query

import (
	"context"
	"database/sql"
	"os"
	"testing"

	"github.com/chapmanjacobd/discotheque/internal/models"
	"github.com/chapmanjacobd/discotheque/internal/utils"
	_ "github.com/mattn/go-sqlite3"
)

func TestNewQueryBuilder(t *testing.T) {
	flags := models.GlobalFlags{QueryFlags: models.QueryFlags{Query: "SELECT 1"}}
	qb := NewQueryBuilder(flags)
	if qb.Flags.Query != "SELECT 1" {
		t.Errorf("Expected query SELECT 1, got %s", qb.Flags.Query)
	}
}

func TestQueryBuilder_Build(t *testing.T) {
	tests := []struct {
		name     string
		flags    models.GlobalFlags
		expected string
	}{
		{
			"Raw Query",
			models.GlobalFlags{QueryFlags: models.QueryFlags{Query: "SELECT * FROM test"}},
			"SELECT * FROM test",
		},
		{
			"Default Query",
			models.GlobalFlags{
				SortFlags:   models.SortFlags{SortBy: "path"},
				QueryFlags:  models.QueryFlags{Limit: 100},
				FilterFlags: models.FilterFlags{HideDeleted: true},
			},
			"SELECT * FROM media WHERE COALESCE(time_deleted, 0) = 0 ORDER BY path ASC LIMIT 100",
		},
		{
			"Search Query",
			models.GlobalFlags{
				FilterFlags: models.FilterFlags{Search: []string{"term"}, HideDeleted: true},
				SortFlags:   models.SortFlags{SortBy: "path"},
				QueryFlags:  models.QueryFlags{Limit: 100},
			},
			"SELECT * FROM media WHERE COALESCE(time_deleted, 0) = 0 AND ((path LIKE ? OR title LIKE ?)) ORDER BY path ASC LIMIT 100",
		},
		{
			"Video Only",
			models.GlobalFlags{
				FilterFlags: models.FilterFlags{VideoOnly: true, HideDeleted: true},
				SortFlags:   models.SortFlags{SortBy: "path"},
				QueryFlags:  models.QueryFlags{Limit: 100},
			},
			"SELECT * FROM media WHERE COALESCE(time_deleted, 0) = 0 AND (" + utils.ExtensionsToLike(utils.VideoExtensions) + ") ORDER BY path ASC LIMIT 100",
		},
		{
			"Reverse Sort",
			models.GlobalFlags{
				SortFlags:   models.SortFlags{SortBy: "path", Reverse: true},
				QueryFlags:  models.QueryFlags{Limit: 10},
				FilterFlags: models.FilterFlags{HideDeleted: true},
			},
			"SELECT * FROM media WHERE COALESCE(time_deleted, 0) = 0 ORDER BY path DESC LIMIT 10",
		},
		{
			"Random Sort",
			models.GlobalFlags{
				SortFlags:   models.SortFlags{Random: true},
				QueryFlags:  models.QueryFlags{Limit: 10},
				FilterFlags: models.FilterFlags{HideDeleted: true},
			},
			"SELECT * FROM media WHERE COALESCE(time_deleted, 0) = 0 AND rowid IN (SELECT rowid FROM media WHERE COALESCE(time_deleted, 0) = 0 ORDER BY RANDOM() LIMIT 160) ORDER BY RANDOM() LIMIT 10",
		},
		{
			"FTS Search",
			models.GlobalFlags{
				FTSFlags:    models.FTSFlags{FTS: true, FTSTable: "media_fts"},
				FilterFlags: models.FilterFlags{Search: []string{"term"}, HideDeleted: true},
				QueryFlags:  models.QueryFlags{Limit: 10},
			},
			"SELECT * FROM media WHERE COALESCE(time_deleted, 0) = 0 AND rowid IN (SELECT rowid FROM media_fts WHERE media_fts MATCH ?) LIMIT 10",
		},
		{
			"Only Deleted",
			models.GlobalFlags{
				FilterFlags: models.FilterFlags{OnlyDeleted: true},
				QueryFlags:  models.QueryFlags{Limit: 10},
			},
			"SELECT * FROM media WHERE COALESCE(time_deleted, 0) > 0 LIMIT 10",
		},
		{
			"Portrait",
			models.GlobalFlags{
				FilterFlags: models.FilterFlags{Portrait: true, HideDeleted: true},
				QueryFlags:  models.QueryFlags{Limit: 10},
			},
			"SELECT * FROM media WHERE COALESCE(time_deleted, 0) = 0 AND width < height LIMIT 10",
		},
		{
			"Online Only",
			models.GlobalFlags{
				FilterFlags: models.FilterFlags{OnlineMediaOnly: true, HideDeleted: true},
				QueryFlags:  models.QueryFlags{Limit: 10},
			},
			"SELECT * FROM media WHERE COALESCE(time_deleted, 0) = 0 AND path LIKE 'http%' LIMIT 10",
		},
		{
			"Custom Where",
			models.GlobalFlags{
				FilterFlags: models.FilterFlags{Where: []string{"play_count > 5"}, HideDeleted: true},
				QueryFlags:  models.QueryFlags{Limit: 10},
			},
			"SELECT * FROM media WHERE COALESCE(time_deleted, 0) = 0 AND play_count > 5 LIMIT 10",
		},
		{
			"Partial Skip",
			models.GlobalFlags{
				FilterFlags: models.FilterFlags{Partial: "s", HideDeleted: true},
				QueryFlags:  models.QueryFlags{Limit: 10},
			},
			"SELECT * FROM media WHERE COALESCE(time_deleted, 0) = 0 AND COALESCE(time_first_played, 0) = 0 LIMIT 10",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			qb := NewQueryBuilder(tt.flags)
			query, _ := qb.Build()
			if query != tt.expected {
				t.Errorf("Build() query = %q, want %q", query, tt.expected)
			}
		})
	}
}

func TestFilterMedia(t *testing.T) {
	var size100 int64 = 100
	var size200 int64 = 200
	media := []models.MediaWithDB{
		{Media: models.Media{Path: "test1.mp4", Size: &size100}, DB: "db1"},
		{Media: models.Media{Path: "test2.mkv", Size: &size200}, DB: "db1"},
	}

	tests := []struct {
		name     string
		flags    models.GlobalFlags
		expected int
	}{
		{"No filters", models.GlobalFlags{}, 2},
		{"Include filter", models.GlobalFlags{FilterFlags: models.FilterFlags{Include: []string{"%.mp4"}}}, 1},
		{"Exclude filter", models.GlobalFlags{FilterFlags: models.FilterFlags{Exclude: []string{"%.mp4"}}}, 1},
		{"Size filter", models.GlobalFlags{FilterFlags: models.FilterFlags{Size: []string{">150B"}}}, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FilterMedia(media, tt.flags)
			if len(got) != tt.expected {
				t.Errorf("FilterMedia() len = %v, want %v", len(got), tt.expected)
			}
		})
	}
}

func TestSortMedia(t *testing.T) {
	var size100 int64 = 100
	var size200 int64 = 200
	media := []models.MediaWithDB{
		{Media: models.Media{Path: "b.mp4", Size: &size200}},
		{Media: models.Media{Path: "a.mp4", Size: &size100}},
	}

	SortMedia(media, models.PlaybackFlags{GlobalFlags: models.GlobalFlags{
		SortFlags: models.SortFlags{SortBy: "path"},
	}})
	if media[0].Path != "a.mp4" {
		t.Errorf("SortMedia by path failed, got %s", media[0].Path)
	}

	SortMedia(media, models.PlaybackFlags{GlobalFlags: models.GlobalFlags{
		SortFlags: models.SortFlags{SortBy: "size"},
	}})
	if *media[0].Size != 100 {
		t.Errorf("SortMedia by size failed, got %d", *media[0].Size)
	}

	// Test that SortBy is respected even when PlayInOrder is set to default
	media = []models.MediaWithDB{
		{Media: models.Media{Path: "a.mp4", Size: &size200}},
		{Media: models.Media{Path: "b.mp4", Size: &size100}},
	}
	SortMedia(media, models.PlaybackFlags{
		GlobalFlags: models.GlobalFlags{
			SortFlags: models.SortFlags{SortBy: "size"},
		},
		PlaybackOptions: models.PlaybackOptions{PlayInOrder: "natural_ps"},
	})
	if *media[0].Size != 100 {
		t.Errorf("SortMedia by size with natural_ps failed, got %d", *media[0].Size)
	}
}

func TestSortFolders(t *testing.T) {
	folders := []models.FolderStats{
		{Path: "b", Count: 2},
		{Path: "a", Count: 1},
	}

	SortFolders(folders, "path", false)
	if folders[0].Path != "a" {
		t.Errorf("SortFolders by path failed, got %s", folders[0].Path)
	}

	SortFolders(folders, "count", true)
	if folders[0].Count != 2 {
		t.Errorf("SortFolders by count desc failed, got %d", folders[0].Count)
	}
}

func TestQueryDatabase(t *testing.T) {
	// Create a temporary database file
	f, err := os.CreateTemp("", "query-test-*.db")
	if err != nil {
		t.Fatal(err)
	}
	dbPath := f.Name()
	f.Close()
	defer os.Remove(dbPath)

	database, err := sql.Open("sqlite3", dbPath)
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
		playhead INTEGER DEFAULT 0,
		type TEXT
	);
	CREATE TABLE history (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		media_path TEXT NOT NULL,
		time_played INTEGER,
		playhead INTEGER,
		done INTEGER
	);
	CREATE TABLE captions (
		media_path TEXT NOT NULL,
		time REAL,
		text TEXT
	);
	`
	if _, err := database.Exec(schema); err != nil {
		database.Close()
		t.Fatal(err)
	}

	// Insert test data
	insert := `INSERT INTO media (path, title, duration, size, type) VALUES (?, ?, ?, ?, ?)`
	database.Exec(insert, "/test/movie.mp4", "Test Movie", 7200, 1000000, "video")
	database.Close()

	ctx := context.Background()
	results, err := QueryDatabase(ctx, dbPath, "SELECT * FROM media", nil)
	if err != nil {
		t.Fatalf("QueryDatabase failed: %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}
	if results[0].Path != "/test/movie.mp4" {
		t.Errorf("Expected path /test/movie.mp4, got %s", results[0].Path)
	}
}

func TestSummarizeMedia(t *testing.T) {
	var size100 int64 = 100
	var size200 int64 = 200
	var size300 int64 = 300
	var dur10 int64 = 10
	var dur20 int64 = 20
	var dur30 int64 = 30

	media := []models.MediaWithDB{
		{Media: models.Media{Size: &size100, Duration: &dur10}},
		{Media: models.Media{Size: &size200, Duration: &dur20}},
		{Media: models.Media{Size: &size300, Duration: &dur30}},
	}

	got := SummarizeMedia(media)
	if len(got) != 2 {
		t.Fatalf("SummarizeMedia() returned %d items, want 2", len(got))
	}

	total := got[0]
	if total.Label != "Total" || total.Count != 3 || total.TotalSize != 600 || total.TotalDuration != 60 {
		t.Errorf("Total stats incorrect: %+v", total)
	}

	median := got[1]
	if median.Label != "Median" || median.TotalSize != 200 || median.TotalDuration != 20 {
		t.Errorf("Median stats incorrect: %+v", median)
	}
}

func TestMediaQuery(t *testing.T) {
	// Create two temporary databases
	f1, _ := os.CreateTemp("", "query-test1-*.db")
	dbPath1 := f1.Name()
	f1.Close()
	defer os.Remove(dbPath1)

	f2, _ := os.CreateTemp("", "query-test2-*.db")
	dbPath2 := f2.Name()
	f2.Close()
	defer os.Remove(dbPath2)

	schema := `
	CREATE TABLE media (path TEXT PRIMARY KEY, time_deleted INTEGER DEFAULT 0);
	CREATE TABLE history (id INTEGER PRIMARY KEY AUTOINCREMENT, media_path TEXT NOT NULL, time_played INTEGER, playhead INTEGER, done INTEGER);
	CREATE TABLE captions (media_path TEXT NOT NULL, time REAL, text TEXT);
	`
	for _, dbPath := range []string{dbPath1, dbPath2} {
		db, _ := sql.Open("sqlite3", dbPath)
		db.Exec(schema)
		db.Exec("INSERT INTO media (path) VALUES (?)", dbPath)
		db.Close()
	}

	ctx := context.Background()
	flags := models.GlobalFlags{
		QueryFlags: models.QueryFlags{Limit: 10},
		SortFlags:  models.SortFlags{SortBy: "path"},
	}
	results, err := MediaQuery(ctx, []string{dbPath1, dbPath2}, flags)
	if err != nil {
		t.Fatalf("MediaQuery failed: %v", err)
	}

	if len(results) != 2 {
		t.Fatalf("Expected 2 results, got %d", len(results))
	}
}
