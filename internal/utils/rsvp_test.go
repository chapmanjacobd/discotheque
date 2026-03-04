package utils

import (
	"os"
	"strings"
	"testing"
)

func TestGenerateRSVPAss(t *testing.T) {
	text := "Hello world this is RSVP"
	wpm := 60 // 1 word per second
	ass, duration := GenerateRSVPAss(text, wpm)

	if duration != 5.0 {
		t.Errorf("expected duration 5.0, got %f", duration)
	}

	if !strings.Contains(ass, "Dialogue: 0,0:00:00.00,0:00:01.00,Default,,0,0,0,,Hello") {
		t.Errorf("ASS content missing first word or timing incorrect")
	}
	if !strings.Contains(ass, "Dialogue: 0,0:00:04.00,0:00:05.00,Default,,0,0,0,,RSVP") {
		t.Errorf("ASS content missing last word or timing incorrect")
	}
}

func TestExtractText(t *testing.T) {
	// Test plain text
	tmpFile, _ := os.CreateTemp("", "test*.txt")
	defer os.Remove(tmpFile.Name())
	content := "Test content"
	os.WriteFile(tmpFile.Name(), []byte(content), 0644)

	text, err := ExtractText(tmpFile.Name())
	if err != nil {
		t.Fatalf("ExtractText failed: %v", err)
	}
	if strings.TrimSpace(text) != content {
		t.Errorf("expected %q, got %q", content, text)
	}
}
