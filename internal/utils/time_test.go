package utils

import (
	"testing"
	"time"
)

func TestParseDate(t *testing.T) {
	tests := []struct {
		input    string
		expected int64 // Just checking if it returns > 0 for valid dates
	}{
		{"2024-01-01", 1704067200},
		{"01/01/2024", 1704067200},
		{"invalid", 0},
	}
	for _, tt := range tests {
		got := ParseDate(tt.input)
		if tt.expected > 0 && got <= 0 {
			t.Errorf("ParseDate(%q) expected positive timestamp, got %v", tt.input, got)
		}
		if tt.expected == 0 && got != 0 {
			t.Errorf("ParseDate(%q) expected 0, got %v", tt.input, got)
		}
	}
}

func TestSuperParser(t *testing.T) {
	tests := []string{
		"2024-05-20",
		"May 20, 2024",
		"20/05/2024",
		"2024/05/20 15:04:05",
	}
	for _, tt := range tests {
		got := SuperParser(tt)
		if got == nil {
			t.Errorf("SuperParser(%q) expected valid time, got nil", tt)
		}
	}
}

func TestSpecificDate(t *testing.T) {
	// Should pick earliest most-specific past date
	d1 := "2020-01-01" // specific
	d2 := "2019-05-20" // earlier and specific
	d3 := "2025-01-01" // future (assuming now > 2025) - wait, today is 2026!
	// Feb 18, 2026 is today.

	got := SpecificDate(d1, d2, d3)
	if got == nil {
		t.Fatal("SpecificDate returned nil")
	}

	t2 := SuperParser(d2).Unix()
	if *got != t2 {
		t.Errorf("SpecificDate expected %v, got %v", t2, *got)
	}
}

func TestTubeDate(t *testing.T) {
	v := map[string]any{
		"upload_date": "20240520",
		"title":       "test",
	}
	got := TubeDate(v)
	if got == nil {
		t.Fatal("TubeDate returned nil")
	}

	if _, ok := v["upload_date"]; ok {
		t.Error("TubeDate should have removed upload_date from map")
	}
}

func TestParseDateOrRelative(t *testing.T) {
	now := time.Now().Unix()
	tests := []struct {
		input    string
		expected int64
		margin   int64
	}{
		{"2024-01-01", 1704067200, 0},
		{"3 days", now - 3*86400, 5},
		{"-1 week", now - 7*86400, 5},
		{"+1 hour", now + 3600, 5},
	}

	for _, tt := range tests {
		got := ParseDateOrRelative(tt.input)
		if tt.margin == 0 {
			if got != tt.expected {
				t.Errorf("ParseDateOrRelative(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		} else {
			diff := got - tt.expected
			if diff < 0 {
				diff = -diff
			}
			if diff > tt.margin {
				t.Errorf("ParseDateOrRelative(%q) = %v, want near %v (diff %v)", tt.input, got, tt.expected, diff)
			}
		}
	}
}

func TestUtcFromLocalTimestamp(t *testing.T) {
	ts := int64(1716217445) // 2024-05-20 15:04:05 UTC
	got := UtcFromLocalTimestamp(ts)
	if got.Unix() != ts {
		t.Errorf("UtcFromLocalTimestamp expected %v, got %v", ts, got.Unix())
	}
	if got.Location() != time.UTC {
		t.Error("UtcFromLocalTimestamp should return time in UTC")
	}
}
