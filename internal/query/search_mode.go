package query

import (
	"context"
	"database/sql"
	"sync"

	database "github.com/chapmanjacobd/discoteca/internal/db"
)

// SearchMode represents the search backend being used
type SearchMode int

const (
	SearchModeUnknown   SearchMode = iota
	SearchModeSubstring            // LIKE-based search
	SearchModeFTS5                 // SQLite FTS5
)

func (s SearchMode) String() string {
	switch s {
	case SearchModeSubstring:
		return "Substring"
	case SearchModeFTS5:
		return "FTS5"
	default:
		return "Unknown"
	}
}

var (
	detectedSearchMode SearchMode
	detectionOnce      sync.Once
	detectionMutex     sync.RWMutex
)

// DetectSearchMode detects the best available search backend
// Priority: FTS5 > Substring
func DetectSearchMode(ctx context.Context, db *sql.DB) SearchMode {
	detectionOnce.Do(func() {
		// Check for FTS5
		if database.FtsEnabled && db != nil && hasFTS5Table(ctx, db) {
			detectedSearchMode = SearchModeFTS5
			return
		}

		// Default to substring search
		detectedSearchMode = SearchModeSubstring
	})

	detectionMutex.RLock()
	defer detectionMutex.RUnlock()
	return detectedSearchMode
}

// hasFTS5Table checks if the media_fts table exists in the database
func hasFTS5Table(ctx context.Context, db *sql.DB) bool {
	if db == nil {
		return false
	}

	var exists bool
	err := db.QueryRowContext(
		ctx,
		"SELECT EXISTS(SELECT 1 FROM sqlite_master WHERE type='table' AND name='media_fts')",
	).Scan(&exists)
	return err == nil && exists
}

// GetSearchMode returns the currently detected search mode
func GetSearchMode() SearchMode {
	detectionMutex.RLock()
	defer detectionMutex.RUnlock()
	return detectedSearchMode
}

// ResetSearchModeDetection resets the detection cache (useful for testing)
func ResetSearchModeDetection() {
	detectionMutex.Lock()
	defer detectionMutex.Unlock()
	detectedSearchMode = SearchModeUnknown
	detectionOnce = sync.Once{}
}

// IsSearchAvailable checks if a specific search mode is available
func IsSearchAvailable(ctx context.Context, mode SearchMode, db *sql.DB) bool {
	switch mode {
	case SearchModeFTS5:
		return db != nil && hasFTS5Table(ctx, db)
	case SearchModeSubstring:
		return true // Always available
	default:
		return false
	}
}
