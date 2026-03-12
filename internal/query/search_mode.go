package query

import (
	"database/sql"
	"sync"

	"github.com/chapmanjacobd/discoteca/internal/bleve"
)

// SearchMode represents the search backend being used
type SearchMode int

const (
	SearchModeUnknown SearchMode = iota
	SearchModeSubstring          // LIKE-based search
	SearchModeFTS5               // SQLite FTS5
	SearchModeBleve              // Bleve full-text search
)

func (s SearchMode) String() string {
	switch s {
	case SearchModeSubstring:
		return "Substring"
	case SearchModeFTS5:
		return "FTS5"
	case SearchModeBleve:
		return "Bleve"
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
// Priority: Bleve > FTS5 > Substring
func DetectSearchMode(db *sql.DB) SearchMode {
	detectionOnce.Do(func() {
		// Check for Bleve first (highest priority)
		if bleve.GetIndex() != nil {
			detectedSearchMode = SearchModeBleve
			return
		}

		// Check for FTS5
		if db != nil && hasFTS5Table(db) {
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
func hasFTS5Table(db *sql.DB) bool {
	if db == nil {
		return false
	}

	var exists bool
	err := db.QueryRow(
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
func IsSearchAvailable(mode SearchMode, db *sql.DB) bool {
	switch mode {
	case SearchModeBleve:
		return bleve.GetIndex() != nil
	case SearchModeFTS5:
		return db != nil && hasFTS5Table(db)
	case SearchModeSubstring:
		return true // Always available
	default:
		return false
	}
}
