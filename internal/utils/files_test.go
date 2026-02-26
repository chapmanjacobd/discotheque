package utils

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/chapmanjacobd/discotheque/internal/models"
)

func TestSampleHashFile(t *testing.T) {
	f, err := os.CreateTemp("", "hash-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())

	content := "hello world"
	if _, err := f.WriteString(content); err != nil {
		t.Fatal(err)
	}
	f.Close()

	hash1, err := SampleHashFile(f.Name(), 1, 0, 0)
	if err != nil {
		t.Fatalf("SampleHashFile failed: %v", err)
	}
	hash2, err := SampleHashFile(f.Name(), 1, 0, 0)
	if err != nil {
		t.Fatalf("SampleHashFile failed: %v", err)
	}
	if hash1 != hash2 {
		t.Errorf("Hashes for same file differ: %s, %s", hash1, hash2)
	}

	if _, err := SampleHashFile("/non/existent", 1, 0, 0); err == nil {
		t.Error("Expected error for non-existent file")
	}
}

func TestFullHashFile(t *testing.T) {
	f, err := os.CreateTemp("", "full-hash-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())

	content := "hello world"
	if _, err := f.WriteString(content); err != nil {
		t.Fatal(err)
	}
	f.Close()

	hash, err := FullHashFile(f.Name())
	if err != nil {
		t.Fatalf("FullHashFile failed: %v", err)
	}
	if hash == "" {
		t.Error("Empty hash")
	}
}

func TestSimulationFunctions(t *testing.T) {
	flags := models.GlobalFlags{
		CoreFlags: models.CoreFlags{Simulate: true},
	}
	if err := Rename(flags, "src", "dst"); err != nil {
		t.Errorf("Rename failed in simulation: %v", err)
	}
	if err := Unlink(flags, "path"); err != nil {
		t.Errorf("Unlink failed in simulation: %v", err)
	}
	if err := Rmtree(flags, "path"); err != nil {
		t.Errorf("Rmtree failed in simulation: %v", err)
	}
}

func TestAltName(t *testing.T) {
	f, err := os.CreateTemp("", "alt-test.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	f.Close()

	alt := AltName(f.Name())
	if alt == f.Name() {
		t.Errorf("AltName should be different for existing file: %s", alt)
	}

	nonExistent := filepath.Join(os.TempDir(), "non-existent-file-xyz.txt")
	if got := AltName(nonExistent); got != nonExistent {
		t.Errorf("AltName mismatch for non-existent file: %s", got)
	}
}

func TestGetExternalSubtitles(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "subs-test")
	defer os.RemoveAll(tmpDir)

	video := filepath.Join(tmpDir, "movie.mp4")
	os.WriteFile(video, nil, 0o644)
	os.WriteFile(filepath.Join(tmpDir, "movie.srt"), nil, 0o644)
	os.WriteFile(filepath.Join(tmpDir, "movie.en.srt"), nil, 0o644)

	subs := GetExternalSubtitles(video)
	if len(subs) != 2 {
		t.Errorf("Expected 2 subtitles, got %d: %v", len(subs), subs)
	}
}

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
	expected := "/home/user/vids"
	if got != expected {
		t.Errorf("CommonPathFull expected %q, got %q", expected, got)
	}
}
