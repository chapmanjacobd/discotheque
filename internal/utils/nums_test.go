package utils

import (
	"fmt"
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
		result, err := HumanToBytes(tt.input)
		if err != nil {
			t.Errorf("HumanToBytes(%q) error: %v", tt.input, err)
			continue
		}
		if result != tt.expected {
			t.Errorf("HumanToBytes(%q) = %d, want %d", tt.input, result, tt.expected)
		}
	}
}

func TestHumanToSeconds(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"1 hour", 3600},
		{"30 min", 1800},
		{"45s", 45},
		{"100", 100},
		{"1 day", 86400},
		{"1 week", 604800},
	}
	for _, tt := range tests {
		result, err := HumanToSeconds(tt.input)
		if err != nil {
			t.Errorf("HumanToSeconds(%q) error: %v", tt.input, err)
			continue
		}
		if result != tt.expected {
			t.Errorf("HumanToSeconds(%q) = %d, want %d", tt.input, result, tt.expected)
		}
	}
}

func TestParseRange(t *testing.T) {
	mockHumanToX := func(s string) (int64, error) {
		if s == "100" {
			return 100, nil
		}
		return 0, fmt.Errorf("invalid")
	}

	tests := []struct {
		input string
		check func(Range) bool
	}{
		{">100", func(r Range) bool { return r.Min != nil && *r.Min == 101 }},
		{"+100", func(r Range) bool { return r.Min != nil && *r.Min == 100 }},
		{"<100", func(r Range) bool { return r.Max != nil && *r.Max == 99 }},
		{"-100", func(r Range) bool { return r.Max != nil && *r.Max == 100 }},
		{"100%10", func(r Range) bool { return r.Min != nil && *r.Min == 90 && r.Max != nil && *r.Max == 110 }},
	}

	for _, tt := range tests {
		r, err := ParseRange(tt.input, mockHumanToX)
		if err != nil {
			t.Errorf("ParseRange(%q) error: %v", tt.input, err)
			continue
		}
		if !tt.check(r) {
			t.Errorf("ParseRange(%q) failed check: %+v", tt.input, r)
		}
	}
}

func TestPercent(t *testing.T) {
	if got := Percent(50, 200); got != 25.0 {
		t.Errorf("Percent(50, 200) = %v, want 25.0", got)
	}
}

func TestFloatFromPercent(t *testing.T) {
	tests := []struct {
		input    string
		expected float64
	}{
		{"50%", 0.5},
		{"0.5", 0.5},
	}
	for _, tt := range tests {
		got, _ := FloatFromPercent(tt.input)
		if got != tt.expected {
			t.Errorf("FloatFromPercent(%q) = %v, want %v", tt.input, got, tt.expected)
		}
	}
}
