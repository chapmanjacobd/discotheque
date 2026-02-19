package metadata

import (
	"reflect"
	"testing"
)

func TestParseFPS(t *testing.T) {
	tests := []struct {
		input    string
		expected float64
	}{
		{"24/1", 24},
		{"30000/1001", 29.97002997002997},
		{"60/1", 60},
		{"0/0", 0},
		{"invalid", 0},
	}

	for _, tt := range tests {
		result := parseFPS(tt.input)
		if result != tt.expected {
			t.Errorf("parseFPS(%q) = %f, want %f", tt.input, result, tt.expected)
		}
	}
}

func TestUnique(t *testing.T) {
	tests := []struct {
		input    []string
		expected []string
	}{
		{[]string{"h264", "h264", "aac"}, []string{"h264", "aac"}},
		{[]string{"mp3", "", "mp3"}, []string{"mp3"}},
		{[]string{}, nil},
	}

	for _, tt := range tests {
		result := unique(tt.input)
		if !reflect.DeepEqual(result, tt.expected) {
			t.Errorf("unique(%v) = %v, want %v", tt.input, result, tt.expected)
		}
	}
}
