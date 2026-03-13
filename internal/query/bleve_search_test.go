//go:build bleve

package query

import (
	"testing"

	"github.com/chapmanjacobd/discoteca/internal/bleve"
	"github.com/chapmanjacobd/discoteca/internal/testutils"
	_ "github.com/mattn/go-sqlite3"
)

func TestBleveSearch(t *testing.T) {
	fixture := testutils.Setup(t)
	defer fixture.Cleanup()

	dbPath := fixture.DBPath

	// Initialize bleve index
	if err := bleve.InitIndex(dbPath); err != nil {
		t.Fatalf("InitIndex failed: %v", err)
	}
	defer bleve.CloseIndex()

	// Index test documents
	docs := []*bleve.MediaDocument{
		{
			ID:            "doc-1",
			Path:          "/media/videos/action-video1.mp4",
			PathTokenized: "/media/videos/action-video1",
			Title:         "Sample Action Video",
			Description:   "An exciting action video with great scenes",
			Type:          "video",
		},
		{
			ID:            "doc-2",
			Path:          "/media/videos/comedy-video2.mp4",
			PathTokenized: "/media/videos/comedy-video2",
			Title:         "Sample Comedy Video",
			Description:   "A hilarious comedy video clip",
			Type:          "video",
		},
		{
			ID:            "doc-3",
			Path:          "/media/music/rock-audio.mp3",
			PathTokenized: "/media/music/rock-audio",
			Title:         "Sample Rock Audio",
			Description:   "A great rock music track",
			Type:          "audio",
		},
	}

	for _, doc := range docs {
		if err := bleve.IndexDocument(doc); err != nil {
			t.Fatalf("IndexDocument failed: %v", err)
		}
	}

	tests := []struct {
		name        string
		searchTerms []string
		limit       int
		expectCount int
		expectIDs   []string
	}{
		{
			name:        "Single term search",
			searchTerms: []string{"action"},
			limit:       10,
			expectCount: 1,
			expectIDs:   []string{"doc-1"},
		},
		{
			name:        "Multiple search terms",
			searchTerms: []string{"video", "action"},
			limit:       10,
			expectCount: 1,
			expectIDs:   []string{"doc-1"},
		},
		{
			name:        "Search with limit",
			searchTerms: []string{"media"},
			limit:       2,
			expectCount: 2,
		},
		{
			name:        "No results",
			searchTerms: []string{"nonexistent"},
			limit:       10,
			expectCount: 0,
		},
		{
			name:        "Empty search terms",
			searchTerms: []string{},
			limit:       10,
			expectCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ids, err := BleveSearch(tt.searchTerms, tt.limit)
			if err != nil {
				t.Fatalf("BleveSearch failed: %v", err)
			}

			if len(ids) != tt.expectCount {
				t.Errorf("Expected %d results, got %d: %v", tt.expectCount, len(ids), ids)
			}

			if tt.expectIDs != nil && len(tt.expectIDs) > 0 {
				if len(ids) != len(tt.expectIDs) {
					t.Errorf("Expected IDs %v, got %v", tt.expectIDs, ids)
				} else {
					for i, id := range tt.expectIDs {
						if ids[i] != id {
							t.Errorf("Expected ID %s at position %d, got %s", id, i, ids[i])
						}
					}
				}
			}
		})
	}
}

func TestBleveSearchWithoutIndex(t *testing.T) {
	// Ensure index is closed
	bleve.CloseIndex()

	ids, err := BleveSearch([]string{"test"}, 10)
	if err != nil {
		t.Errorf("Expected no error when searching without index, got %v", err)
	}
	if len(ids) != 0 {
		t.Errorf("Expected empty ids, got %v", ids)
	}
}

func TestBleveSearchTermJoining(t *testing.T) {
	fixture := testutils.Setup(t)
	defer fixture.Cleanup()

	dbPath := fixture.DBPath

	if err := bleve.InitIndex(dbPath); err != nil {
		t.Fatalf("InitIndex failed: %v", err)
	}
	defer bleve.CloseIndex()

	// Index a document
	doc := &bleve.MediaDocument{
		ID:            "doc-1",
		Path:          "/test/path.mp4",
		PathTokenized: "/test/path",
		Title:         "Test Document",
		Description:   "This is a test document with multiple words",
		Type:          "video",
	}
	if err := bleve.IndexDocument(doc); err != nil {
		t.Fatalf("IndexDocument failed: %v", err)
	}

	// Test that multiple search terms are joined correctly
	ids, err := BleveSearch([]string{"test", "document"}, 10)
	if err != nil {
		t.Fatalf("BleveSearch with multiple terms failed: %v", err)
	}
	if len(ids) < 1 {
		t.Errorf("Expected at least 1 result for 'test document', got %d", len(ids))
	}
}

func TestHasBleveIndex(t *testing.T) {
	fixture := testutils.Setup(t)
	defer fixture.Cleanup()

	dbPath := fixture.DBPath

	// Test without initialization
	if HasBleveIndex() {
		t.Error("Expected HasBleveIndex to return false before initialization")
	}

	// Initialize index
	if err := bleve.InitIndex(dbPath); err != nil {
		t.Fatalf("InitIndex failed: %v", err)
	}
	defer bleve.CloseIndex()

	// Test with initialization
	if !HasBleveIndex() {
		t.Error("Expected HasBleveIndex to return true after initialization")
	}
}

func TestBleveSearchIntegration(t *testing.T) {
	fixture := testutils.Setup(t)
	defer fixture.Cleanup()

	dbPath := fixture.DBPath

	// Initialize bleve index
	if err := bleve.InitIndex(dbPath); err != nil {
		t.Fatalf("InitIndex failed: %v", err)
	}
	defer bleve.CloseIndex()

	// Create a realistic set of test documents
	docs := []*bleve.MediaDocument{
		{
			ID:            "video-1",
			Path:          "/home/media/videos/sample-video-1999.mp4",
			PathTokenized: "/home/media/videos/sample-video-1999",
			Title:         "Sample Video Title",
			Description:   "A sample video file with test content",
			Type:          "video",
			Size:          1073741824, // 1GB
			Duration:      8160,       // 136 minutes
		},
		{
			ID:            "video-2",
			Path:          "/home/media/videos/another-video-2010.mp4",
			PathTokenized: "/home/media/videos/another-video-2010",
			Title:         "Another Video",
			Description:   "Another sample video file",
			Type:          "video",
			Size:          2147483648, // 2GB
			Duration:      8880,       // 148 minutes
		},
		{
			ID:            "show-1",
			Path:          "/home/media/shows/sample-show/S01E01.mp4",
			PathTokenized: "/home/media/shows/sample-show/S01E01",
			Title:         "Sample Show S01E01",
			Description:   "A sample show episode",
			Type:          "video",
			Size:          536870912, // 512MB
			Duration:      3480,      // 58 minutes
		},
		{
			ID:            "music-1",
			Path:          "/home/media/music/sample-artist/sample-album.mp3",
			PathTokenized: "/home/media/music/sample-artist/sample-album",
			Title:         "Sample Track",
			Description:   "Classic sample album by sample artist",
			Type:          "audio",
			Size:          10485760, // 10MB
			Duration:      382,      // 6:22 minutes
		},
	}

	for _, doc := range docs {
		if err := bleve.IndexDocument(doc); err != nil {
			t.Fatalf("IndexDocument failed: %v", err)
		}
	}

	tests := []struct {
		name        string
		searchTerms []string
		limit       int
		minResults  int
		maxResults  int
	}{
		{
			name:        "Search by path component",
			searchTerms: []string{"videos"},
			limit:       10,
			minResults:  2,
			maxResults:  2,
		},
		{
			name:        "Search by path keyword",
			searchTerms: []string{"shows"},
			limit:       10,
			minResults:  1,
			maxResults:  1,
		},
		{
			name:        "Search with zero limit",
			searchTerms: []string{"media"},
			limit:       0,
			minResults:  0,
			maxResults:  0,
		},
		{
			name:        "Search with limit 1",
			searchTerms: []string{"media"},
			limit:       1,
			minResults:  1,
			maxResults:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ids, err := BleveSearch(tt.searchTerms, tt.limit)
			if err != nil {
				t.Fatalf("BleveSearch failed: %v", err)
			}

			if len(ids) < tt.minResults {
				t.Errorf("Expected at least %d results, got %d", tt.minResults, len(ids))
			}
			if len(ids) > tt.maxResults {
				t.Errorf("Expected at most %d results, got %d", tt.maxResults, len(ids))
			}
		})
	}
}

func TestBleveSearchPathIntegration(t *testing.T) {
	fixture := testutils.Setup(t)
	defer fixture.Cleanup()

	dbPath := fixture.DBPath

	if err := bleve.InitIndex(dbPath); err != nil {
		t.Fatalf("InitIndex failed: %v", err)
	}
	defer bleve.CloseIndex()

	// Index documents with hierarchical paths
	docs := []*bleve.MediaDocument{
		{
			ID:            "doc-1",
			Path:          "/data/videos/movies/action/movie1.mp4",
			PathTokenized: "/data/videos/movies/action/movie1",
			Title:         "Action Movie 1",
		},
		{
			ID:            "doc-2",
			Path:          "/data/videos/movies/action/movie2.mp4",
			PathTokenized: "/data/videos/movies/action/movie2",
			Title:         "Action Movie 2",
		},
		{
			ID:            "doc-3",
			Path:          "/data/videos/movies/comedy/movie3.mp4",
			PathTokenized: "/data/videos/movies/comedy/movie3",
			Title:         "Comedy Movie",
		},
		{
			ID:            "doc-4",
			Path:          "/data/music/rock/song1.mp3",
			PathTokenized: "/data/music/rock/song1",
			Title:         "Rock Song",
		},
	}

	for _, doc := range docs {
		if err := bleve.IndexDocument(doc); err != nil {
			t.Fatalf("IndexDocument failed: %v", err)
		}
	}

	// Use bleve.SearchPath directly for path-specific tests
	ids, err := bleve.SearchPath("/data/videos/movies/action", 10)
	if err != nil {
		t.Fatalf("SearchPath failed: %v", err)
	}
	if len(ids) != 2 {
		t.Errorf("Expected 2 action movies, got %d: %v", len(ids), ids)
	}

	ids, err = bleve.SearchPath("/data/music", 10)
	if err != nil {
		t.Fatalf("SearchPath for music failed: %v", err)
	}
	if len(ids) != 1 {
		t.Errorf("Expected 1 music file, got %d: %v", len(ids), ids)
	}
}

func TestBleveSearchCaseSensitivity(t *testing.T) {
	fixture := testutils.Setup(t)
	defer fixture.Cleanup()

	dbPath := fixture.DBPath

	if err := bleve.InitIndex(dbPath); err != nil {
		t.Fatalf("InitIndex failed: %v", err)
	}
	defer bleve.CloseIndex()

	// Index a document
	doc := &bleve.MediaDocument{
		ID:            "doc-1",
		Path:          "/test/sample-video.mp4",
		PathTokenized: "/test/sample-video",
		Title:         "Sample Video Title",
		Description:   "Sample description text",
	}
	if err := bleve.IndexDocument(doc); err != nil {
		t.Fatalf("IndexDocument failed: %v", err)
	}

	// Test case insensitive search (standard analyzer should handle this)
	tests := []struct {
		term string
	}{
		{"sample"},
		{"SAMPLE"},
		{"Sample"},
		{"SaMpLe"},
	}

	for _, tt := range tests {
		t.Run(tt.term, func(t *testing.T) {
			ids, err := BleveSearch([]string{tt.term}, 10)
			if err != nil {
				t.Fatalf("BleveSearch for '%s' failed: %v", tt.term, err)
			}
			if len(ids) < 1 {
				t.Errorf("Expected results for case-insensitive search '%s', got %d", tt.term, len(ids))
			}
		})
	}
}

func TestBleveSearchSpecialCharacters(t *testing.T) {
	fixture := testutils.Setup(t)
	defer fixture.Cleanup()

	dbPath := fixture.DBPath

	if err := bleve.InitIndex(dbPath); err != nil {
		t.Fatalf("InitIndex failed: %v", err)
	}
	defer bleve.CloseIndex()

	// Index documents with special characters in paths and titles
	docs := []*bleve.MediaDocument{
		{
			ID:            "doc-1",
			Path:          "/media/videos/sample-video-1999-1080p.mp4",
			PathTokenized: "/media/videos/sample-video-1999-1080p",
			Title:         "Sample Video (1999)",
			Description:   "Video with special chars: @#$%",
			Type:          "video",
		},
		{
			ID:            "doc-2",
			Path:          "/media/videos/another-sample-video.mp4",
			PathTokenized: "/media/videos/another-sample-video",
			Title:         "Another Sample Video",
			Description:   "Classic sample video",
			Type:          "video",
		},
	}

	for _, doc := range docs {
		if err := bleve.IndexDocument(doc); err != nil {
			t.Fatalf("IndexDocument failed: %v", err)
		}
	}

	// Test search with numbers
	ids, err := BleveSearch([]string{"1999"}, 10)
	if err != nil {
		t.Fatalf("BleveSearch for year failed: %v", err)
	}
	if len(ids) < 1 {
		t.Errorf("Expected results for '1999', got %d", len(ids))
	}

	// Test search with partial title
	ids, err = BleveSearch([]string{"sample"}, 10)
	if err != nil {
		t.Fatalf("BleveSearch for 'sample' failed: %v", err)
	}
	if len(ids) < 1 {
		t.Errorf("Expected results for 'sample', got %d", len(ids))
	}
}
