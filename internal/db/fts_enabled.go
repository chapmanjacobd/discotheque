//go:build fts5

package db

const FtsEnabled = true

// _fts5BuildTagRequired is defined only when fts5 tag is used.
// This satisfies the reference in fts_required.go and prevents compile errors.
var _fts5BuildTagRequired struct{}

// Ensure the variable is used to avoid staticcheck warnings
// This is a compile-time check only, no runtime cost
var _ = _fts5BuildTagRequired
