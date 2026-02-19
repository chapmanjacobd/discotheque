package utils

import (
	"testing"
)

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
		result := FormatDuration(tt.input)
		if result != tt.expected {
			t.Errorf("FormatDuration(%d) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestFormatDurationShort(t *testing.T) {
	tests := []struct {
		input    int
		expected string
	}{
		{45, "45 seconds"},
		{60, "1 minute"},
		{66, "1.1 minutes"},
		{3600, "1 hour"},
		{86400, "1 day"},
		{172800, "2 days"},
		{946684800, "30 years and 7 days"},
	}

	for _, tt := range tests {
		result := FormatDurationShort(tt.input)
		if result != tt.expected {
			t.Errorf("FormatDurationShort(%d) = %q, want %q", tt.input, result, tt.expected)
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
		result := FormatSize(tt.input)
		if result != tt.expected {
			t.Errorf("FormatSize(%d) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestFormatPlaybackDuration(t *testing.T) {
	tests := []struct {
		duration     int64
		segmentStart int64
		segmentEnd   int64
		expected     string
	}{
		{0, 0, 0, "Duration: 0:00"},
		{360, 1000, 0, "Duration: 6:00 (16:40 to 22:40)"},
		{3600, 0, 0, "Duration: 1:00:00"},
		{3600, 0, 3000, "Duration: 50:00 (0:00 to 50:00)"},
		{3600, 1800, 0, "Duration: 30:00 (30:00 to 1:00:00)"}, 
		{3600, 1800, 3000, "Duration: 20:00 (30:00 to 50:00)"}, 
		{3600, 3000, 0, "Duration: 10:00 (50:00 to 1:00:00)"},
		{3600, 3000, 2000, "Duration: 16:40 (33:20 to 50:00)"}, // Swap: 3000, 2000 -> 2000, 3000.
	}

	// Actually, let's re-verify the Python expected values.
	// (3600, 1800, 0, "Duration: 30:00 (30:00 to 1:00:00)")
	// 1800 seconds is 30:00.
	// (3600, 1800, 3000, "Duration: 20:00 (30:00 to 50:00)")
	// 1800+3000 = 4800 > 3600. Swap: 1800, 3000 -> 3000, 1800.
	// wait, if start=3000, end=1800.
	// duration = 3000 - 1800 = 1200 (20:00).
	// start_str = 1800 (30:00)? No, start_str = segment_start.
	// If it was swapped, segment_start is 3000? No, 1800, 3000 -> 3000, 1800. segment_start is 3000.
	// 3000 seconds is 50:00.
	// Wait, the Python expected: "Duration: 20:00 (30:00 to 50:00)"
	// 30:00 is 1800. 50:00 is 3000.
	// So it seems it does segment_start, segment_end = sorted(segment_start, segment_end).
	// My implementation: segmentStart, segmentEnd = segmentEnd, segmentStart (swapped).
	// Let's adjust tests to match exactly what Python expected.

	for _, tt := range tests {
		result := FormatPlaybackDuration(tt.duration, tt.segmentStart, tt.segmentEnd)
		if result != tt.expected {
			t.Errorf("FormatPlaybackDuration(%d, %d, %d) = %q, want %q", tt.duration, tt.segmentStart, tt.segmentEnd, result, tt.expected)
		}
	}
}
