package sort

import (
	"testing"

	"github.com/chapmanjacobd/discotheque/internal/db"
)

func TestApply_BySize(t *testing.T) {
	media := []db.Media{
		{Path: "/c", Size: 3000},
		{Path: "/a", Size: 1000},
		{Path: "/b", Size: 2000},
	}

	Apply(media, BySize, false, false)

	expected := []int64{1000, 2000, 3000}
	for i, m := range media {
		if m.Size != expected[i] {
			t.Errorf("Index %d: expected size %d, got %d", i, expected[i], m.Size)
		}
	}
}

func TestApply_BySizeReverse(t *testing.T) {
	media := []db.Media{
		{Path: "/c", Size: 3000},
		{Path: "/a", Size: 1000},
		{Path: "/b", Size: 2000},
	}

	Apply(media, BySize, true, false)

	expected := []int64{3000, 2000, 1000}
	for i, m := range media {
		if m.Size != expected[i] {
			t.Errorf("Index %d: expected size %d, got %d", i, expected[i], m.Size)
		}
	}
}

func TestApply_NaturalSort(t *testing.T) {
	media := []db.Media{
		{Path: "/show/episode10.mp4"},
		{Path: "/show/episode2.mp4"},
		{Path: "/show/episode1.mp4"},
		{Path: "/show/episode20.mp4"},
	}

	Apply(media, ByPath, false, true)

	expected := []string{
		"/show/episode1.mp4",
		"/show/episode2.mp4",
		"/show/episode10.mp4",
		"/show/episode20.mp4",
	}

	for i, m := range media {
		if m.Path != expected[i] {
			t.Errorf("Index %d: expected %s, got %s", i, expected[i], m.Path)
		}
	}
}

func TestNaturalLess(t *testing.T) {
	tests := []struct {
		s1       string
		s2       string
		expected bool
	}{
		{"file1.txt", "file2.txt", true},
		{"file2.txt", "file1.txt", false},
		{"file1.txt", "file10.txt", true},
		{"file10.txt", "file2.txt", false},
		{"Season 1 Episode 1", "Season 1 Episode 10", true},
		{"S01E01", "S01E02", true},
		{"S01E02", "S01E01", false},
		{"S01E09", "S01E10", true},
	}

	for _, tt := range tests {
		result := naturalLess(tt.s1, tt.s2)
		if result != tt.expected {
			t.Errorf("naturalLess(%q, %q) = %v, want %v", tt.s1, tt.s2, result, tt.expected)
		}
	}
}
