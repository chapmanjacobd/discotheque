package main

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	"github.com/chapmanjacobd/discotheque/internal/db"
	"github.com/chapmanjacobd/discotheque/internal/history"
	"github.com/chapmanjacobd/discotheque/internal/models"
	"github.com/chapmanjacobd/discotheque/internal/query"
	_ "github.com/mattn/go-sqlite3"
)

// TestFixture provides a complete test environment
type TestFixture struct {
	DB      *sql.DB
	Queries *db.Queries
	Tracker *history.Tracker
	TempDir string
}

func setupIntegrationTest(t *testing.T) *TestFixture {
	// Create temp directory for test files
	tempDir, err := os.MkdirTemp("", "lb-test-*")
	if err != nil {
		t.Fatal(err)
	}

	// Create in-memory database
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
	CREATE INDEX idx_time_deleted ON media(time_deleted);
	CREATE INDEX idx_time_last_played ON media(time_last_played);

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
		t.Fatal(err)
	}

	// Insert comprehensive test data
	testData := []struct {
		path           string
		title          string
		duration       int32
		size           int64
		playCount      int32
		playhead       int32
		timeLastPlayed int64
	}{
		// TV Shows - Season 1
		{filepath.Join(tempDir, "tv/Show/S01E01.mp4"), "Pilot", 2700, 500_000_000, 2, 2700, 1700000000},
		{filepath.Join(tempDir, "tv/Show/S01E02.mp4"), "Episode 2", 2700, 480_000_000, 1, 1200, 1700000100},
		{filepath.Join(tempDir, "tv/Show/S01E10.mp4"), "Finale", 3600, 520_000_000, 0, 0, 0},

		// TV Shows - Season 2
		{filepath.Join(tempDir, "tv/Show/S02E01.mp4"), "New Season", 2700, 490_000_000, 0, 0, 0},

		// Movies - Action
		{filepath.Join(tempDir, "movies/action/BigMovie.2024.1080p.mp4"), "Big Action", 7200, 2_000_000_000, 1, 7200, 1700000200},
		{filepath.Join(tempDir, "movies/action/SmallMovie.720p.mp4"), "Small Action", 5400, 800_000_000, 0, 0, 0},
		{filepath.Join(tempDir, "movies/action/sample.mp4"), "Sample", 300, 50_000_000, 0, 0, 0},

		// Movies - Comedy
		{filepath.Join(tempDir, "movies/comedy/Funny.2023.mp4"), "Funny Movie", 6000, 1_200_000_000, 3, 0, 1700000300},
		{filepath.Join(tempDir, "movies/comedy/Short.mp4"), "Short Comedy", 1800, 300_000_000, 0, 900, 0},

		// Documentaries
		{filepath.Join(tempDir, "docs/nature/Wildlife.mp4"), "Wildlife Doc", 5400, 1_500_000_000, 1, 2700, 1700000400},
		{filepath.Join(tempDir, "docs/history/Ancient.mp4"), "Ancient History", 7200, 1_800_000_000, 0, 0, 0},

		// Audiobooks
		{filepath.Join(tempDir, "audiobooks/Fiction/Book1-01.m4a"), "Chapter 1", 3600, 100_000_000, 2, 3600, 1700000500},
		{filepath.Join(tempDir, "audiobooks/Fiction/Book1-02.m4a"), "Chapter 2", 3600, 100_000_000, 1, 1800, 1700000600},
		{filepath.Join(tempDir, "audiobooks/Fiction/Book1-03.m4a"), "Chapter 3", 3600, 100_000_000, 0, 0, 0},
	}

	for _, td := range testData {
		_, err := database.Exec(`
			INSERT INTO media (path, title, duration, size, play_count, playhead, time_last_played)
			VALUES (?, ?, ?, ?, ?, ?, ?)
		`, td.path, td.title, td.duration, td.size, td.playCount, td.playhead, td.timeLastPlayed)
		if err != nil {
			t.Fatal(err)
		}
	}

	return &TestFixture{
		DB:      database,
		Queries: db.New(database),
		Tracker: history.NewTracker(database),
		TempDir: tempDir,
	}
}

func (f *TestFixture) Cleanup() {
	f.DB.Close()
	os.RemoveAll(f.TempDir)
}

// Integration Test 1: Filter + Sort + Aggregate Pipeline
func TestIntegration_FilterSortAggregate(t *testing.T) {
	fixture := setupIntegrationTest(t)
	defer fixture.Cleanup()

	ctx := context.Background()

	// Get all media
	dbMedia, err := fixture.Queries.GetMedia(ctx, 100)
	if err != nil {
		t.Fatal(err)
	}

	var allMedia []models.MediaWithDB
	for _, m := range dbMedia {
		allMedia = append(allMedia, models.FromDBWithDB(m, "test.db"))
	}

	// Filter: Large video files (> 450MB) that are not samples
	flags := models.GlobalFlags{
		FilterFlags: models.FilterFlags{
			Size:    []string{">450MB"},
			Exclude: []string{"*sample*"},
		},
	}
	filtered := query.FilterMedia(allMedia, flags)

	if len(filtered) != 9 { // Should exclude small files and sample
		t.Errorf("Expected 9 filtered results, got %d", len(filtered))
	}

	// Sort by natural order (important for TV shows)
	query.SortMedia(filtered, models.PlaybackFlags{GlobalFlags: models.GlobalFlags{
		SortFlags: models.SortFlags{SortBy: "path", NatSort: true},
	}})

	// Verify S01E01 comes before S01E10
	var s01e01Idx, s01e10Idx int = -1, -1
	for i, m := range filtered {
		if filepath.Base(m.Path) == "S01E01.mp4" {
			s01e01Idx = i
		}
		if filepath.Base(m.Path) == "S01E10.mp4" {
			s01e10Idx = i
		}
	}
	if s01e01Idx == -1 || s01e10Idx == -1 || s01e01Idx >= s01e10Idx {
		t.Error("Natural sort failed: S01E01 should come before S01E10")
	}

	// Aggregate by folder
	folders := query.AggregateMedia(filtered, models.GlobalFlags{
		DisplayFlags: models.DisplayFlags{BigDirs: true},
	})

	// Find TV show folder
	var tvFolder *models.FolderStats
	for i := range folders {
		if filepath.Base(folders[i].Path) == "Show" {
			tvFolder = &folders[i]
			break
		}
	}

	if tvFolder == nil {
		t.Fatal("TV Show folder not found")
	}

	if tvFolder.Count != 4 { // S01E01, S01E02, S01E10, S02E01
		t.Errorf("Expected 4 files in TV folder, got %d", tvFolder.Count)
	}

	// Sort folders by size
	query.SortFolders(folders, "size", true)

	// Largest folder should be action movies
	if filepath.Base(folders[0].Path) != "action" {
		t.Errorf("Expected action folder to be largest, got %s", folders[0].Path)
	}
}

// Integration Test 2: Watch History + Unfinished Content
func TestIntegration_WatchHistoryAndUnfinished(t *testing.T) {
	fixture := setupIntegrationTest(t)
	defer fixture.Cleanup()

	ctx := context.Background()

	// Get unfinished media (has playhead but not complete)
	dbUnfinished, err := fixture.Queries.GetUnfinishedMedia(ctx, 100)
	if err != nil {
		t.Fatal(err)
	}

	var unfinished []models.Media
	for _, m := range dbUnfinished {
		unfinished = append(unfinished, models.FromDB(m))
	}

	// Should find S01E02, Short Comedy, Wildlife Doc, Book1-02.m4a
	if len(unfinished) != 4 {
		t.Errorf("Expected 4 unfinished items, got %d", len(unfinished))
	}

	// Play one of the unfinished items to completion
	if len(unfinished) > 0 {
		path := unfinished[0].Path
		duration := int32(*unfinished[0].Duration)

		err := fixture.Tracker.UpdatePlayback(ctx, path, duration)
		if err != nil {
			t.Fatal(err)
		}

		// Verify play count increased
		var playCount int32
		err = fixture.DB.QueryRow("SELECT play_count FROM media WHERE path = ?", path).Scan(&playCount)
		if err != nil {
			t.Fatal(err)
		}

		expectedCount := int32(*unfinished[0].PlayCount) + 1
		if playCount != expectedCount {
			t.Errorf("Expected play count %d, got %d", expectedCount, playCount)
		}
	}

	// Get most watched content
	watched, err := fixture.Queries.GetMediaByPlayCount(ctx, db.GetMediaByPlayCountParams{
		PlayCount:   sql.NullInt64{Int64: 2, Valid: true},
		PlayCount_2: sql.NullInt64{Int64: 100, Valid: true},
		Limit:       10,
	})
	if err != nil {
		t.Fatal(err)
	}

	// Should find items with play_count >= 2
	if len(watched) < 2 {
		t.Errorf("Expected at least 2 highly watched items, got %d", len(watched))
	}
}

// Integration Test 3: Complex Query - Unwatched HD Content
func TestIntegration_UnwatchedHDContent(t *testing.T) {
	fixture := setupIntegrationTest(t)
	defer fixture.Cleanup()

	ctx := context.Background()

	// Get unwatched media
	dbUnwatched, err := fixture.Queries.GetUnwatchedMedia(ctx, 100)
	if err != nil {
		t.Fatal(err)
	}

	var unwatched []models.MediaWithDB
	for _, m := range dbUnwatched {
		unwatched = append(unwatched, models.FromDBWithDB(m, "test.db"))
	}

	// Filter for unwatched files > 500MB and longer than 1 hour
	flags := models.GlobalFlags{
		FilterFlags: models.FilterFlags{
			Size:     []string{">500MB"},
			Duration: []string{">1hour"},
		},
	}
	hdUnwatched := query.FilterMedia(unwatched, flags)

	// Should find S01E10 and possibly others
	if len(hdUnwatched) == 0 {
		t.Error("Expected to find some unwatched HD content")
	}

	// Verify all results match criteria
	for _, m := range hdUnwatched {
		if *m.Size < 500_000_000 {
			t.Errorf("File %s too small: %d bytes", m.Path, *m.Size)
		}
		if *m.Duration < 3600 {
			t.Errorf("File %s too short: %d seconds", m.Path, *m.Duration)
		}
	}
}

// Integration Test 4: Regex Filter + Natural Sort + Size Limits
func TestIntegration_RegexNaturalSortSize(t *testing.T) {
	fixture := setupIntegrationTest(t)
	defer fixture.Cleanup()

	ctx := context.Background()

	allMediaRaw, err := fixture.Queries.GetMedia(ctx, 100)
	if err != nil {
		t.Fatal(err)
	}

	var allMedia []models.MediaWithDB
	for _, m := range allMediaRaw {
		allMedia = append(allMedia, models.FromDBWithDB(m, "test.db"))
	}

	// Find all Season 1 episodes using regex
	flags := models.GlobalFlags{
		FilterFlags: models.FilterFlags{
			Regex: `S01E\d+`,
			Size:  []string{">400MB"}, // Exclude samples
		},
	}
	season1 := query.FilterMedia(allMedia, flags)

	if len(season1) != 3 {
		t.Errorf("Expected 3 Season 1 episodes, got %d", len(season1))
	}

	// Natural sort to get correct episode order
	query.SortMedia(season1, models.PlaybackFlags{GlobalFlags: models.GlobalFlags{
		SortFlags: models.SortFlags{SortBy: "path", NatSort: true},
	}})

	// Verify episode order
	expectedOrder := []string{"S01E01.mp4", "S01E02.mp4", "S01E10.mp4"}
	for i, expected := range expectedOrder {
		if i >= len(season1) {
			break
		}
		actual := filepath.Base(season1[i].Path)
		if actual != expected {
			t.Errorf("Position %d: expected %s, got %s", i, expected, actual)
		}
	}

	// Calculate total watch time for season
	var totalDuration int32
	for _, m := range season1 {
		totalDuration += int32(*m.Duration)
	}

	expectedTotal := int32(2700 + 2700 + 3600) // Episode durations
	if totalDuration != expectedTotal {
		t.Errorf("Expected total duration %d, got %d", expectedTotal, totalDuration)
	}
}

// Integration Test 5: Multi-Database Query Simulation
func TestIntegration_MultiDatabaseScenario(t *testing.T) {
	// Simulate querying multiple databases with different content

	// Database 1: Movies
	fixture1 := setupIntegrationTest(t)
	defer fixture1.Cleanup()

	// Database 2: TV Shows (new instance)
	fixture2 := setupIntegrationTest(t)
	defer fixture2.Cleanup()

	ctx := context.Background()

	// Get media from both databases
	dbMedia1, err := fixture1.Queries.GetMedia(ctx, 100)
	if err != nil {
		t.Fatal(err)
	}

	dbMedia2, err := fixture2.Queries.GetMedia(ctx, 100)
	if err != nil {
		t.Fatal(err)
	}

	// Combine results
	var allMedia []models.MediaWithDB
	for _, m := range dbMedia1 {
		allMedia = append(allMedia, models.FromDBWithDB(m, "db1.db"))
	}
	for _, m := range dbMedia2 {
		allMedia = append(allMedia, models.FromDBWithDB(m, "db2.db"))
	}

	// Filter for videos only (exclude audiobooks)
	flags := models.GlobalFlags{
		FilterFlags: models.FilterFlags{
			Exclude: []string{"*.m4a", "*.mp3", "*.flac"},
		},
	}
	videos := query.FilterMedia(allMedia, flags)

	// Should have twice the video content (minus audiobooks)
	if len(videos) < 20 { // Each fixture has 11 video files
		t.Errorf("Expected at least 20 video files from both databases, got %d", len(videos))
	}

	// Aggregate by folder across both databases
	folders := query.AggregateMedia(videos, models.GlobalFlags{
		DisplayFlags: models.DisplayFlags{BigDirs: true},
	})

	// Sort by total size
	query.SortFolders(folders, "size", true)

	// Verify we have folders from both databases
	uniqueFolders := make(map[string]bool)
	for _, f := range folders {
		uniqueFolders[f.Path] = true
	}

	if len(uniqueFolders) < 5 {
		t.Errorf("Expected at least 5 unique folders, got %d", len(uniqueFolders))
	}
}

// Integration Test 6: Complete Workflow - Watch, Track, Filter Unwatched
func TestIntegration_CompleteWatchWorkflow(t *testing.T) {
	fixture := setupIntegrationTest(t)
	defer fixture.Cleanup()

	ctx := context.Background()

	// Step 1: Get unwatched content
	dbUnwatched, err := fixture.Queries.GetUnwatchedMedia(ctx, 100)
	if err != nil {
		t.Fatal(err)
	}

	var unwatched []models.MediaWithDB
	for _, m := range dbUnwatched {
		unwatched = append(unwatched, models.FromDBWithDB(m, "test.db"))
	}

	initialUnwatchedCount := len(unwatched)

	// Step 2: Filter for short content (easier to "watch" in test)
	flags := models.GlobalFlags{
		FilterFlags: models.FilterFlags{
			Duration: []string{"<1hour"},
			Size:     []string{">100MB"},
		},
	}
	toWatch := query.FilterMedia(unwatched, flags)

	if len(toWatch) == 0 {
		t.Fatal("No content to watch")
	}

	// Step 3: Natural sort (watch in order)
	query.SortMedia(toWatch, models.PlaybackFlags{GlobalFlags: models.GlobalFlags{
		SortFlags: models.SortFlags{SortBy: "path", NatSort: true},
	}})

	// Step 4: "Watch" first item
	firstItem := toWatch[0]
	err = fixture.Tracker.UpdatePlayback(ctx, firstItem.Path, int32(*firstItem.Duration))
	if err != nil {
		t.Fatal(err)
	}

	// Step 5: Verify it's marked as watched
	var playCount int32
	var timeLastPlayed int64
	err = fixture.DB.QueryRow(
		"SELECT play_count, time_last_played FROM media WHERE path = ?",
		firstItem.Path,
	).Scan(&playCount, &timeLastPlayed)
	if err != nil {
		t.Fatal(err)
	}

	if playCount != int32(*firstItem.PlayCount)+1 {
		t.Errorf("Play count not incremented: expected %d, got %d",
			int32(*firstItem.PlayCount)+1, playCount)
	}

	if timeLastPlayed == 0 {
		t.Error("time_last_played not set")
	}

	// Step 6: Query unwatched again - should have one less
	unwatchedAfter, err := fixture.Queries.GetUnwatchedMedia(ctx, 100)
	if err != nil {
		t.Fatal(err)
	}

	if len(unwatchedAfter) != initialUnwatchedCount-1 {
		t.Errorf("Expected %d unwatched items, got %d",
			initialUnwatchedCount-1, len(unwatchedAfter))
	}

	// Step 7: Get recently watched
	recentlyWatched, err := fixture.Queries.GetWatchedMedia(ctx, 10)
	if err != nil {
		t.Fatal(err)
	}

	// First item should be what we just watched
	if len(recentlyWatched) == 0 || recentlyWatched[0].Path != firstItem.Path {
		t.Error("Recently watched query doesn't show our watched item first")
	}
}

// Integration Test 7: Folder Aggregation with Mixed Content
func TestIntegration_FolderStatsAccuracy(t *testing.T) {
	fixture := setupIntegrationTest(t)
	defer fixture.Cleanup()

	ctx := context.Background()

	// Get all media
	dbMedia, err := fixture.Queries.GetMedia(ctx, 100)
	if err != nil {
		t.Fatal(err)
	}

	var allMedia []models.MediaWithDB
	for _, m := range dbMedia {
		allMedia = append(allMedia, models.FromDBWithDB(m, "test.db"))
	}

	// Aggregate by folder
	folders := query.AggregateMedia(allMedia, models.GlobalFlags{
		DisplayFlags: models.DisplayFlags{BigDirs: true},
	})

	// Find action movies folder
	var actionFolder *models.FolderStats
	for i := range folders {
		if filepath.Base(folders[i].Path) == "action" {
			actionFolder = &folders[i]
			break
		}
	}

	if actionFolder == nil {
		t.Fatal("Action folder not found")
	}

	// Verify folder statistics
	if actionFolder.Count != 3 {
		t.Errorf("Expected 3 action movies, got %d", actionFolder.Count)
	}

	// Calculate expected totals
	var expectedSize int64
	var expectedDuration int32
	for _, f := range actionFolder.Files {
		expectedSize += *f.Size
		expectedDuration += int32(*f.Duration)
	}

	if actionFolder.TotalSize != expectedSize {
		t.Errorf("Total size mismatch: expected %d, got %d",
			expectedSize, actionFolder.TotalSize)
	}

	if actionFolder.TotalDuration != int64(expectedDuration) {
		t.Errorf("Total duration mismatch: expected %d, got %d",
			expectedDuration, actionFolder.TotalDuration)
	}

	if actionFolder.AvgSize != expectedSize/int64(actionFolder.Count) {
		t.Errorf("Average size calculation incorrect")
	}
}

// Benchmark: Filter + Sort on Large Dataset
func BenchmarkIntegration_FilterSort(b *testing.B) {
	fixture := setupIntegrationTest(&testing.T{})
	defer fixture.Cleanup()

	ctx := context.Background()
	dbMedia, _ := fixture.Queries.GetMedia(ctx, 100)
	var allMedia []models.MediaWithDB
	for _, m := range dbMedia {
		allMedia = append(allMedia, models.FromDBWithDB(m, "test.db"))
	}

	flags := models.GlobalFlags{
		FilterFlags: models.FilterFlags{
			Size:    []string{">400MB"},
			Exclude: []string{"*sample*"},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		filtered := query.FilterMedia(allMedia, flags)
		query.SortMedia(filtered, models.PlaybackFlags{GlobalFlags: models.GlobalFlags{
			SortFlags: models.SortFlags{SortBy: "path", NatSort: true},
		}})
	}
}

// Run all integration tests:
// go test -v -run Integration
// go test -v -run TestIntegration_FilterSortAggregate
// go test -bench=. -benchmem
