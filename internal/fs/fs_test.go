package fs

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindMedia(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "fs-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Create a dummy structure
	files := []string{
		"movie.mp4",
		"song.mp3",
		"readme.txt", // should be ignored
		"folder/clip.mkv",
		"folder/image.jpg", // should be ignored by default if not in MediaExtensions
	}

	for _, f := range files {
		path := filepath.Join(tempDir, f)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte("test"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	found, err := FindMedia(tempDir, nil)
	if err != nil {
		t.Fatalf("FindMedia failed: %v", err)
	}

	expectedCount := 4 // mp4, mp3, mkv, jpg
	if len(found) != expectedCount {
		t.Errorf("Expected %d media files, got %d: %v", expectedCount, len(found), found)
	}

	expectedFiles := []string{"movie.mp4", "song.mp3", "clip.mkv", "image.jpg"}
	for _, ef := range expectedFiles {
		matched := false
		for path := range found {
			if filepath.Base(path) == ef {
				matched = true
				break
			}
		}
		if !matched {
			t.Errorf("Expected to find %s", ef)
		}
	}
}
