package main

import (
	"testing"
)

func TestHumanToBytes(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"100", 100},
		{"1KB", 1024},
		{"1MB", 1024 * 1024},
		{"1GB", 1024 * 1024 * 1024},
		{"1.5MB", 1572864},
		{" 100 MB ", 100 * 1024 * 1024},
	}

	for _, tt := range tests {
		result, err := humanToBytes(tt.input)
		if err != nil {
			t.Errorf("humanToBytes(%q) error: %v", tt.input, err)
			continue
		}
		if result != tt.expected {
			t.Errorf("humanToBytes(%q) = %d, want %d", tt.input, result, tt.expected)
		}
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		input    int
		expected string
	}{
		{0, "-"},
		{45, "0:45"},
		{60, "1:00"},
		{3600, "1:00:00"},
		{3661, "1:01:01"},
	}

	for _, tt := range tests {
		result := formatDuration(tt.input)
		if result != tt.expected {
			t.Errorf("formatDuration(%d) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestFormatSize(t *testing.T) {
	tests := []struct {
		input    int64
		expected string
	}{
		{0, "-"},
		{500, "500 B"},
		{1024, "1.0 KB"},
		{1024 * 1024, "1.0 MB"},
		{1024 * 1024 * 1024, "1.0 GB"},
	}

	for _, tt := range tests {
		result := formatSize(tt.input)
		if result != tt.expected {
			t.Errorf("formatSize(%d) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}
