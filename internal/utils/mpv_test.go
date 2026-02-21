package utils

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/chapmanjacobd/discotheque/internal/models"
)

func TestPathToMpvWatchLaterMD5(t *testing.T) {
	path := "/home/xk/github/xk/lb/tests/data/test.mp4"
	expected := "E1E0D0E3F0D2CB748303FDA43224B7E7"

	actual := PathToMpvWatchLaterMD5(path)
	if actual != expected {
		t.Errorf("PathToMpvWatchLaterMD5(%s) = %s; want %s", path, actual, expected)
	}
}

func TestGetPlayhead(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "mpv_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	flags := models.GlobalFlags{
		WatchLaterDir: tmpDir,
	}
	path := "/home/runner/work/library/library/tests/data/test.mp4"
	md5Hash := PathToMpvWatchLaterMD5(path)
	metadataPath := filepath.Join(tmpDir, md5Hash)

	// Use MPV time
	startTime := time.Now().Add(-2 * time.Second)
	os.WriteFile(metadataPath, []byte("start=5.000000\n"), 0o644)
	if ph := GetPlayhead(flags, path, startTime, 0, 0); ph != 5 {
		t.Errorf("GetPlayhead (mpv time) = %d; want 5", ph)
	}

	// Check invalid MPV time (beyond duration)
	os.WriteFile(metadataPath, []byte("start=13.000000\n"), 0o644)
	// GetPlayhead currently returns mpvPlayhead if found.
	// The Python code:
	//   if mpv_playhead: return mpv_playhead
	// Wait, I should re-read Python's get_playhead logic.

	/*
	   for playhead in [mpv_playhead or 0, python_playhead]:
	       if playhead > 0 and (media_duration is None or media_duration >= playhead):
	           return playhead
	*/

	// So if mpv_playhead is 13 and media_duration is 12, it skips 13 and tries python_playhead (2).

	if ph := GetPlayhead(flags, path, startTime, 0, 12); ph != 2 {
		t.Errorf("GetPlayhead (invalid mpv time) = %d; want 2", ph)
	}

	// Use session time only if MPV does not exist
	os.Remove(metadataPath)
	if ph := GetPlayhead(flags, path, startTime, 0, 0); ph != 2 {
		t.Errorf("GetPlayhead (session time) = %d; want 2", ph)
	}

	// Append existing time
	startTime = time.Now().Add(-3 * time.Second)
	if ph := GetPlayhead(flags, path, startTime, 4, 12); ph != 7 {
		t.Errorf("GetPlayhead (existing time) = %d; want 7", ph)
	}
}

func TestMpvArgsToMap(t *testing.T) {
	args := []string{"--volume=50,mute=yes", "--speed=1.5"}
	expected := map[string]string{
		"volume": "50",
		"mute":   "yes",
		"speed":  "1.5",
	}

	actual := MpvArgsToMap(args)
	if len(actual) != len(expected) {
		t.Errorf("MpvArgsToMap len = %d; want %d", len(actual), len(expected))
	}
	for k, v := range expected {
		if actual[k] != v {
			t.Errorf("MpvArgsToMap[%s] = %s; want %s", k, actual[k], v)
		}
	}
}
