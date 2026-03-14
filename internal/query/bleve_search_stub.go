//go:build !bleve

package query

// BleveSearch executes a Bleve search and returns matching paths
// Stub implementation when bleve is not enabled
func BleveSearch(searchTerms []string, limit int) ([]string, error) {
	return nil, nil
}

// BleveSearchPaginated executes a Bleve search with pagination support
// Stub implementation when bleve is not enabled
func BleveSearchPaginated(searchTerms []string, limit, offset int, searchAfter []string) ([]string, uint64, []string, error) {
	return nil, 0, nil, nil
}

// HasBleveIndex checks if a Bleve index is available
// Stub implementation when bleve is not enabled
func HasBleveIndex() bool {
	return false
}
