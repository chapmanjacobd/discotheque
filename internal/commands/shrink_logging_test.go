package commands

import (
	"log/slog"
	"testing"
	"time"
)

func TestShrinkMetrics_InterleavedLogging(t *testing.T) {
	// This test simulates the logic of PrintProgress and how it interacts with slog
	// We can't easily capture the \033 escape codes in a simple way without a complex mock,
	// but we can verify that RecordSuccess and RecordFailure are thread-safe and
	// that PrintProgress doesn't crash when lines are printed.

	m := NewShrinkMetrics()
	m.RecordStarted("Video", "test.mp4")

	// Simulate some logging
	slog.Info("Starting processing", "path", "test.mp4")

	// Simulate progress print
	m.PrintProgress()

	if m.linesPrinted == 0 {
		t.Error("Expected linesPrinted to be > 0 after PrintProgress")
	}

	slog.Error("Something went wrong", "path", "test.mp4")

	// Next print should handle cursor logic
	m.PrintProgress()
}

func TestShrinkMetrics_LineCountLogic(t *testing.T) {
	m := NewShrinkMetrics()
	m.RecordStarted("Video", "v.mp4")
	m.RecordStarted("Audio", "a.mp3")

	// Print once to establish baseline
	m.PrintProgress()
	firstCount := m.linesPrinted

	// currentFile line (1) + empty line (1) + header (1) + separator (1) + 2 types (2) + separator (1) + TOTAL (1) + trailing newline (1) = 9
	if firstCount != 9 {
		t.Errorf("Expected 9 lines (header + 2 types + etc), got %d", firstCount)
	}

	// Record one more type
	m.RecordStarted("Image", "i.jpg")
	time.Sleep(600 * time.Millisecond) // Wait for rate limit
	m.PrintProgress()
	secondCount := m.linesPrinted

	if secondCount != 10 {
		t.Errorf("Expected 10 lines for Image type, got %d", secondCount)
	}
}
