//go:build bleve

package query

import (
	"strings"

	"github.com/chapmanjacobd/discoteca/internal/bleve"
)

// BleveSearch executes a Bleve search and returns matching paths
func BleveSearch(searchTerms []string, limit int) ([]string, error) {
	// Check if bleve index is available
	if bleve.GetIndex() == nil {
		return nil, nil
	}

	// Join search terms for Bleve query
	var query strings.Builder
	for i, term := range searchTerms {
		if i > 0 {
			query.WriteString(" ")
		}
		query.WriteString(term)
	}

	ids, _, err := bleve.Search(query.String(), limit)
	return ids, err
}

// BleveSearchPaginated executes a Bleve search with pagination support
// Returns: matching IDs, total count, search_after keys for next page
// Supports both offset-based (From/Size) and cursor-based (SearchAfter) pagination
func BleveSearchPaginated(searchTerms []string, limit, offset int, searchAfter []string) ([]string, uint64, []string, error) {
	// Check if bleve index is available
	if bleve.GetIndex() == nil {
		return nil, 0, nil, nil
	}

	// Join search terms for Bleve query
	var query strings.Builder
	for i, term := range searchTerms {
		if i > 0 {
			query.WriteString(" ")
		}
		query.WriteString(term)
	}

	return bleve.SearchWithPagination(query.String(), limit, offset, searchAfter)
}

// HasBleveIndex checks if a Bleve index is available
func HasBleveIndex() bool {
	idx := bleve.GetIndex()
	return idx != nil
}
