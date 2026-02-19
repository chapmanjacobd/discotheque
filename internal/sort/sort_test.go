package sort

import (
	"testing"

	"github.com/chapmanjacobd/discotheque/internal/models"
)

func int64Ptr(i int64) *int64 { return &i }

func TestApply_BySize(t *testing.T) {
	media := []models.Media{
		{Path: "/c", Size: int64Ptr(3000)},
		{Path: "/a", Size: int64Ptr(1000)},
		{Path: "/b", Size: int64Ptr(2000)},
	}

	Apply(media, BySize, false, false)

	expected := []int64{1000, 2000, 3000}
	for i, m := range media {
		if *m.Size != expected[i] {
			t.Errorf("Index %d: expected size %d, got %d", i, expected[i], *m.Size)
		}
	}
}

func TestApply_BySizeReverse(t *testing.T) {
	media := []models.Media{
		{Path: "/c", Size: int64Ptr(3000)},
		{Path: "/a", Size: int64Ptr(1000)},
		{Path: "/b", Size: int64Ptr(2000)},
	}

	Apply(media, BySize, true, false)

	expected := []int64{3000, 2000, 1000}
	for i, m := range media {
		if *m.Size != expected[i] {
			t.Errorf("Index %d: expected size %d, got %d", i, expected[i], *m.Size)
		}
	}
}

func TestApply_NaturalSort(t *testing.T) {
	media := []models.Media{
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
