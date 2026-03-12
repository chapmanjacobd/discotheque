package query

import (
	"github.com/chapmanjacobd/discoteca/internal/bleve"
)

// BleveSearch executes a Bleve search and returns matching paths
func BleveSearch(searchTerms []string, limit int) ([]string, error) {
	// Check if bleve index is available
	if bleve.GetIndex() == nil {
		return nil, nil
	}

	// Join search terms for Bleve query
	query := ""
	for i, term := range searchTerms {
		if i > 0 {
			query += " "
		}
		query += term
	}

	ids, _, err := bleve.Search(query, limit)
	return ids, err
}

// HasBleveIndex checks if a Bleve index is available
func HasBleveIndex() bool {
	idx := bleve.GetIndex()
	return idx != nil
}
