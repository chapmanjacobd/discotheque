package metadata

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func createMock(t *testing.T, tmpDir, name, content string) string {
	fullName := name
	if runtime.GOOS == "windows" {
		fullName += ".bat"
	}
	path := filepath.Join(tmpDir, fullName)

	actualContent := content
	if runtime.GOOS == "windows" {
		if name == "ffmpeg" {
			actualContent = `@echo off
for %%a in (%*) do (
    if "%%a"=="20.00" exit /b 1
)
exit /b 0`
		} else if name == "ffprobe" {
			// On Windows, escaping JSON for echo in a .bat is painful.
			// We'll use a slightly more robust approach by escaping quotes.
			escaped := strings.ReplaceAll(content, "\"", "^\"")
			actualContent = "@echo off\necho " + escaped
		}
	} else {
		actualContent = "#!/bin/sh\n" + content
	}

	if err := os.WriteFile(path, []byte(actualContent), 0o755); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestDecodeQuickScan_MockFFmpeg(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "mock-ffmpeg-path")
	defer os.RemoveAll(tmpDir)

	createMock(t, tmpDir, "ffmpeg", `
for arg in "$@"; do
    if [ "$arg" = "20.00" ]; then
        exit 1
    fi
done
exit 0
`)

	oldPath := os.Getenv("PATH")
	os.Setenv("PATH", tmpDir+string(os.PathListSeparator)+oldPath)
	defer os.Setenv("PATH", oldPath)

	scans := []float64{10.0, 20.0, 30.0, 40.0}
	corruption := DecodeQuickScan(context.Background(), "dummy.mp4", scans, 1.0)

	// One scan (20.0) should fail out of four
	if corruption != 0.25 {
		t.Errorf("Expected corruption 0.25, got %f", corruption)
	}
}

func TestDecodeFullScan_MockFFProbe(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "mock-ffprobe-path")
	defer os.RemoveAll(tmpDir)

	createMock(t, tmpDir, "ffprobe", `echo '{
  "streams": [
    {
      "r_frame_rate": "30/1",
      "nb_read_frames": "3000"
    }
  ],
  "format": {
    "duration": "100.0"
  }
}'`)

	oldPath := os.Getenv("PATH")
	os.Setenv("PATH", tmpDir+string(os.PathListSeparator)+oldPath)
	defer os.Setenv("PATH", oldPath)

	// actualDuration = 3000 * 1 / 30 = 100.0
	// metadataDuration = 100.0
	// corruption = 0.0
	corruption, err := DecodeFullScan(context.Background(), "dummy.mp4")
	if err != nil {
		t.Fatalf("DecodeFullScan failed: %v", err)
	}
	if corruption != 0.0 {
		t.Errorf("Expected corruption 0.0, got %f", corruption)
	}
}
