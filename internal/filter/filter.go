package filter

import (
	"path/filepath"
	"regexp"
	"strings"

	"github.com/chapmanjacobd/discotheque/internal/db"
)

type Criteria struct {
	Include      []string
	Exclude      []string
	PathContains []string
	MinSize      int64
	MaxSize      int64
	MinDuration  int64
	MaxDuration  int64
	Regex        *regexp.Regexp
	Exists       bool
}

func Apply(media []db.Media, criteria Criteria) []db.Media {
	var filtered []db.Media

	for _, m := range media {
		if !matches(m, criteria) {
			continue
		}
		filtered = append(filtered, m)
	}

	return filtered
}

func matches(m db.Media, c Criteria) bool {
	// Existence check
	if c.Exists && !fileExists(m.Path) {
		return false
	}

	// Include patterns
	if len(c.Include) > 0 && !matchesAny(m.Path, c.Include) {
		return false
	}

	// Exclude patterns
	if len(c.Exclude) > 0 && matchesAny(m.Path, c.Exclude) {
		return false
	}

	// Path contains
	for _, contain := range c.PathContains {
		if !strings.Contains(m.Path, contain) {
			return false
		}
	}

	// Size filters
	if c.MinSize > 0 {
		if !m.Size.Valid || m.Size.Int64 < c.MinSize {
			return false
		}
	}

	if c.MaxSize > 0 {
		if !m.Size.Valid || m.Size.Int64 > c.MaxSize {
			return false
		}
	}

	// Duration filters
	if c.MinDuration > 0 {
		if !m.Duration.Valid || m.Duration.Int64 < c.MinDuration {
			return false
		}
	}

	if c.MaxDuration > 0 {
		if !m.Duration.Valid || m.Duration.Int64 > c.MaxDuration {
			return false
		}
	}

	// Regex filter
	if c.Regex != nil && !c.Regex.MatchString(m.Path) {
		return false
	}

	return true
}

func matchesAny(path string, patterns []string) bool {
	for _, pattern := range patterns {
		if matched, _ := filepath.Match(pattern, path); matched {
			return true
		}
		if strings.Contains(path, pattern) {
			return true
		}
	}
	return false
}

func fileExists(path string) bool {
	// Implementation from previous version
	return true // placeholder
}
