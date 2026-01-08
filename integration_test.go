package main

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/chapmanjacobd/discotheque/internal/aggregate"
	"github.com/chapmanjacobd/discotheque/internal/db"
	"github.com/chapmanjacobd/discotheque/internal/filter"
	"github.com/chapmanjacobd/discotheque/internal/history"
	"github.com/chapmanjacobd/discotheque/internal/sort"

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
		playhead INTEGER DEFAULT 0
	);
	CREATE INDEX idx_time_deleted ON media(time_deleted);
	CREATE INDEX idx_time_last_played ON media(time_last_played);
	`
	if _, err := database.Exec(schema); err != nil {
		t.Fatal(err)
	}

	// Insert comprehensive test data
	testData := []struct {
		path      string
		title     string
		duration  int32
		size      int64
		playCount int32
		playhead  int32
	}{
		// TV Shows - Season 1
		{filepath.Join(tempDir, "tv/Show/S01E01.mp4"), "Pilot", 2700, 500_000_000, 2, 2700},
		{filepath.Join(tempDir, "tv/Show/S01E02.mp4"), "Episode 2", 2700, 480_000_000, 1, 1200},
		{filepath.Join(tempDir, "tv/Show/S01E10.mp4"), "Finale", 3600, 520_000_000, 0, 0},

		// TV Shows - Season 2
		{filepath.Join(tempDir, "tv/Show/S02E01.mp4"), "New Season", 2700, 490_000_000, 0, 0},

		// Movies - Action
		{filepath.Join(tempDir, "movies/action/BigMovie.2024.1080p.mp4"), "Big Action", 7200, 2_000_000_000, 1, 7200},
		{filepath.Join(tempDir, "movies/action/SmallMovie.720p.mp4"), "Small Action", 5400, 800_000_000, 0, 0},
		{filepath.Join(tempDir, "movies/action/sample.mp4"), "Sample", 300, 50_000_000, 0, 0},

		// Movies - Comedy
		{filepath.Join(tempDir, "movies/comedy/Funny.2023.mp4"), "Funny Movie", 6000, 1_200_000_000, 3, 0},
		{filepath.Join(tempDir, "movies/comedy/Short.mp4"), "Short Comedy", 1800, 300_000_000, 0, 900},

		// Documentaries
		{filepath.Join(tempDir, "docs/nature/Wildlife.mp4"), "Wildlife Doc", 5400, 1_500_000_000, 1, 2700},
		{filepath.Join(tempDir, "docs/history/Ancient.mp4"), "Ancient History", 7200, 1_800_000_000, 0, 0},

		// Audiobooks
		{filepath.Join(tempDir, "audiobooks/Fiction/Book1-01.m4a"), "Chapter 1", 3600, 100_000_000, 2, 3600},
		{filepath.Join(tempDir, "audiobooks/Fiction/Book1-02.m4a"), "Chapter 2", 3600, 100_000_000, 1, 1800},
		{filepath.Join(tempDir, "audiobooks/Fiction/Book1-03.m4a"), "Chapter 3", 3600, 100_000_000, 0, 0},
	}

	for _, td := range testData {
		_, err := database.Exec(`
			INSERT INTO media (path, title, duration, size, play_count, playhead)
			VALUES (?, ?, ?, ?, ?, ?)
		`, td.path, td.title, td.duration, td.size, td.playCount, td.playhead)
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
	allMedia, err := fixture.Queries.GetMedia(ctx, 100)
	if err != nil {
		t.Fatal(err)
	}

	// Filter: Large video files (> 500MB) that are not samples
	criteria := filter.Criteria{
		MinSize: 500_000_000,
		Exclude: []string{"*sample*"},
	}
	filtered := filter.Apply(allMedia, criteria)

	if len(filtered) != 7 { // Should exclude small files and sample
		t.Errorf("Expected 7 filtered results, got %d", len(filtered))
	}

	// Sort by natural order (important for TV shows)
	sort.Apply(filtered, sort.ByPath, false, true)

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
	folders := aggregate.ByFolder(filtered)

	// Find TV show folder
	var tvFolder *aggregate.FolderStats
	for i := range folders {
		if filepath.Base(folders[i].Path) == "Show" {
			tvFolder = &folders[i]
			break
		}
	}

	if tvFolder == nil {
		t.Fatal("TV Show folder not found")
	}

	if tvFolder.Count != 3 { // S01E01, S01E02, S01E10
		t.Errorf("Expected 3 files in TV folder, got %d", tvFolder.Count)
	}

	// Sort folders by size
	aggregate.SortFolders(folders, "size", true)

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
	unfinished, err := fixture.Queries.GetUnfinishedMedia(ctx, 100)
	if err != nil {
		t.Fatal(err)
	}

	// Should find S01E02 and Short Comedy
	if len(unfinished) != 2 {
		t.Errorf("Expected 2 unfinished items, got %d", len(unfinished))
	}

	// Play one of the unfinished items to completion
	if len(unfinished) > 0 {
		path := unfinished[0].Path
		duration := unfinished[0].Duration

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

		expectedCount := unfinished[0].PlayCount + 1
		if playCount != expectedCount {
			t.Errorf("Expected play count %d, got %d", expectedCount, playCount)
		}
	}

	// Get most watched content
	watched, err := fixture.Queries.GetMediaByPlayCount(ctx, db.GetMediaByPlayCountParams{
		PlayCount:   2,
		PlayCount_2: 100,
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
	unwatched, err := fixture.Queries.GetUnwatchedMedia(ctx, 100)
	if err != nil {
		t.Fatal(err)
	}

	// Filter for HD (1080p or larger files > 800MB) and longer than 1 hour
	criteria := filter.Criteria{
		PathContains: []string{"1080p"},
		MinSize:      800_000_000,
		MinDuration:  3600,
	}
	hdUnwatched := filter.Apply(unwatched, criteria)

	// Should find S01E10 and possibly others
	if len(hdUnwatched) == 0 {
		t.Error("Expected to find some unwatched HD content")
	}

	// Verify all results match criteria
	for _, m := range hdUnwatched {
		if m.Size < 800_000_000 {
			t.Errorf("File %s too small: %d bytes", m.Path, m.Size)
		}
		if m.Duration < 3600 {
			t.Errorf("File %s too short: %d seconds", m.Path, m.Duration)
		}
		if m.PlayCount > 0 {
			t.Errorf("File %s already watched", m.Path)
		}
	}
}

// Integration Test 4: Regex Filter + Natural Sort + Size Limits
func TestIntegration_RegexNaturalSortSize(t *testing.T) {
	fixture := setupIntegrationTest(t)
	defer fixture.Cleanup()

	ctx := context.Background()

	allMedia, err := fixture.Queries.GetMedia(ctx, 100)
	if err != nil {
		t.Fatal(err)
	}

	// Find all Season 1 episodes using regex
	criteria := filter.Criteria{
		Regex:   regexp.MustCompile(`S01E\d+`),
		MinSize: 400_000_000, // Exclude samples
	}
	season1 := filter.Apply(allMedia, criteria)

	if len(season1) != 3 {
		t.Errorf("Expected 3 Season 1 episodes, got %d", len(season1))
	}

	// Natural sort to get correct episode order
	sort.Apply(season1, sort.ByPath, false, true)

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
		totalDuration += m.Duration
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
	media1, err := fixture1.Queries.GetMedia(ctx, 100)
	if err != nil {
		t.Fatal(err)
	}

	media2, err := fixture2.Queries.GetMedia(ctx, 100)
	if err != nil {
		t.Fatal(err)
	}

	// Combine results
	allMedia := append(media1, media2...)

	// Filter for videos only (exclude audiobooks)
	criteria := filter.Criteria{
		Exclude: []string{"*.m4a", "*.mp3", "*.flac"},
	}
	videos := filter.Apply(allMedia, criteria)

	// Should have twice the video content (minus audiobooks)
	if len(videos) < 20 { // Each fixture has 11 video files
		t.Errorf("Expected at least 20 video files from both databases, got %d", len(videos))
	}

	// Aggregate by folder across both databases
	folders := aggregate.ByFolder(videos)

	// Sort by total size
	aggregate.SortFolders(folders, "size", true)

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
	unwatched, err := fixture.Queries.GetUnwatchedMedia(ctx, 100)
	if err != nil {
		t.Fatal(err)
	}

	initialUnwatchedCount := len(unwatched)

	// Step 2: Filter for short content (easier to "watch" in test)
	criteria := filter.Criteria{
		MaxDuration: 3600, // 1 hour or less
		MinSize:     100_000_000,
	}
	toWatch := filter.Apply(unwatched, criteria)

	if len(toWatch) == 0 {
		t.Fatal("No content to watch")
	}

	// Step 3: Natural sort (watch in order)
	sort.Apply(toWatch, sort.ByPath, false, true)

	// Step 4: "Watch" first item
	firstItem := toWatch[0]
	err = fixture.Tracker.UpdatePlayback(ctx, firstItem.Path, firstItem.Duration)
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

	if playCount != firstItem.PlayCount+1 {
		t.Errorf("Play count not incremented: expected %d, got %d",
			firstItem.PlayCount+1, playCount)
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
	allMedia, err := fixture.Queries.GetMedia(ctx, 100)
	if err != nil {
		t.Fatal(err)
	}

	// Aggregate by folder
	folders := aggregate.ByFolder(allMedia)

	// Find action movies folder
	var actionFolder *aggregate.FolderStats
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
		expectedSize += f.Size
		expectedDuration += f.Duration
	}

	if actionFolder.TotalSize != expectedSize {
		t.Errorf("Total size mismatch: expected %d, got %d",
			expectedSize, actionFolder.TotalSize)
	}

	if actionFolder.TotalDuration != expectedDuration {
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
	allMedia, _ := fixture.Queries.GetMedia(ctx, 100)

	criteria := filter.Criteria{
		MinSize: 400_000_000,
		Exclude: []string{"*sample*"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		filtered := filter.Apply(allMedia, criteria)
		sort.Apply(filtered, sort.ByPath, false, true)
	}
}

// Run all integration tests:
// go test -v -run Integration
// go test -v -run TestIntegration_FilterSortAggregate
// go test -bench=. -benchmem
