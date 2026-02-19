package utils

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFilterDeleted(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "filter-deleted-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	f1 := filepath.Join(tempDir, "f1.txt")
	if err := os.WriteFile(f1, []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}

	f2 := filepath.Join(tempDir, "f2.txt")
	// f2 does not exist

	paths := []string{f1, f2}
	filtered := FilterDeleted(paths)

	if len(filtered) != 1 {
		t.Errorf("Expected 1 path, got %d", len(filtered))
	}
	if filtered[0] != f1 {
		t.Errorf("Expected path %s, got %s", f1, filtered[0])
	}
}

func TestGetFileStats(t *testing.T) {
	f, err := os.CreateTemp("", "stats-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())

	content := "hello"
	if _, err := f.WriteString(content); err != nil {
		t.Fatal(err)
	}
	f.Close()

	stats, err := GetFileStats(f.Name())
	if err != nil {
		t.Fatalf("GetFileStats failed: %v", err)
	}

	if stats.Size != int64(len(content)) {
		t.Errorf("Expected size %d, got %d", len(content), stats.Size)
	}
	if stats.TimeModified == 0 {
		t.Error("Expected non-zero TimeModified")
	}
}

func TestIsFileOpen(t *testing.T) {
	f, err := os.CreateTemp("", "is-open-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())

	// Test closed file
	f.Close()
	if IsFileOpen(f.Name()) {
		t.Errorf("Expected file %s to be closed", f.Name())
	}

	// Test open file
	f2, err := os.Open(f.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer f2.Close()

	// Note: on Linux, IsFileOpen checks /proc, so it should find its own open FD
	if !IsFileOpen(f.Name()) {
		t.Errorf("Expected file %s to be open", f.Name())
	}
}

func TestDetectMimeType(t *testing.T) {
	f, err := os.CreateTemp("", "mime-test.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())

	if _, err := f.WriteString("hello"); err != nil {
		t.Fatal(err)
	}
	f.Close()

	mime := DetectMimeType(f.Name())
	if mime != "text/plain; charset=utf-8" {
		t.Errorf("Expected text/plain; charset=utf-8, got %s", mime)
	}
}

func TestCommonPath(t *testing.T) {
	paths := []string{
		"/home/user/vids/v1.mp4",
		"/home/user/vids/v2.mp4",
		"/home/user/music/a1.mp3",
	}
	expected := "/home/user"
	got := CommonPath(paths)
	if got != expected {
		t.Errorf("CommonPath expected %q, got %q", expected, got)
	}
}

func TestCommonPathFull(t *testing.T) {
	paths := []string{
		"/home/user/vids/action_movie_part1.mp4",
		"/home/user/vids/action_movie_part2.mp4",
		"/home/user/vids/action_movie_part3.mp4",
	}
	got := CommonPathFull(paths)
	if !strings.Contains(got, "action") || !strings.Contains(got, "movie") {
		t.Errorf("CommonPathFull failed to find common words, got %q", got)
	}
}
