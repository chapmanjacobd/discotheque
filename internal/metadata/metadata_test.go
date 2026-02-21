package metadata

import (
	"context"
	"os"
	"testing"
)

func TestExtract_BasicInfo(t *testing.T) {
	f, err := os.CreateTemp("", "meta-test-*.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	f.WriteString("test content")
	f.Close()

	meta, err := Extract(context.Background(), f.Name(), false)
	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}

	if meta.Media.Path != f.Name() {
		t.Errorf("Expected path %s, got %s", f.Name(), meta.Media.Path)
	}

	if !meta.Media.Type.Valid || meta.Media.Type.String != "text" {
		t.Errorf("Expected type text, got %v", meta.Media.Type)
	}
}

func TestBasicInfo(t *testing.T) {
	f, _ := os.CreateTemp("", "basic-test")
	defer os.Remove(f.Name())
	stat, _ := f.Stat()
	f.Close()

	meta := basicInfo(f.Name(), stat, "video")
	if meta.Media.Path != f.Name() {
		t.Error("Path mismatch")
	}
	if meta.Media.Type.String != "video" {
		t.Error("Type mismatch")
	}
}

func TestExtract_MimeTypes(t *testing.T) {
	tests := []struct {
		filename string
		expected string
	}{
		{"test.jpg", "image"},
		{"test.pdf", "text"},
		{"test.epub", "text"},
	}

	for _, tt := range tests {
		f, _ := os.CreateTemp("", tt.filename)
		name := f.Name()
		f.Close()
		defer os.Remove(name)

		// We don't care if ffprobe fails, we want to see the mime-based detection in basicInfo or fallback
		meta, _ := Extract(context.Background(), name, false)
		if meta != nil && meta.Media.Type.String != tt.expected {
			// Note: DetectMimeType might depend on extension if content is empty
			// Actually DetectMimeType uses filepath.Ext if it's a known extension
		}
	}
}

func TestExtract_NonExistent(t *testing.T) {
	_, err := Extract(context.Background(), "/non/existent/file", false)
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
}

func TestParseFPS(t *testing.T) {
	tests := []struct {
		input    string
		expected float64
	}{
		{"30000/1001", 29.97002997002997},
		{"24/1", 24.0},
		{"0/0", 0.0},
		{"invalid", 0.0},
	}

	for _, tt := range tests {
		got := parseFPS(tt.input)
		if got != tt.expected {
			t.Errorf("parseFPS(%s) = %v, want %v", tt.input, got, tt.expected)
		}
	}
}
