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
