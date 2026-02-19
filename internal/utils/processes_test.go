package utils

import (
	"testing"
)

func TestAdjustDuration(t *testing.T) {
	s10 := 10
	s50 := 50
	tests := []struct {
		duration int
		start    *int
		end      *int
		expected int
	}{
		{100, nil, nil, 100},
		{100, &s10, nil, 90},
		{100, &s10, &s50, 40},
		{100, nil, &s50, 50},
	}

	for _, tt := range tests {
		got := AdjustDuration(tt.duration, tt.start, tt.end)
		if got != tt.expected {
			t.Errorf("AdjustDuration(%d, %v, %v) = %d, want %d", tt.duration, tt.start, tt.end, got, tt.expected)
		}
	}
}

func TestSizeTimeout(t *testing.T) {
	if SizeTimeout("10MB", 5*1024*1024) {
		t.Error("SizeTimeout(10MB, 5MB) should be false")
	}
	if !SizeTimeout("10MB", 11*1024*1024) {
		t.Error("SizeTimeout(10MB, 11MB) should be true")
	}
}
