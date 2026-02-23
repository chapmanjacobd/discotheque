package utils

import (
	"bytes"
	"reflect"
	"strings"
	"testing"
)

func TestPrintOverwrite(t *testing.T) {
	origStdout := Stdout
	defer func() { Stdout = origStdout }()

	var buf bytes.Buffer
	Stdout = &buf

	PrintOverwrite("test message")
	got := buf.String()
	if !strings.Contains(got, "test message") {
		t.Errorf("PrintOverwrite failed, got: %q", got)
	}
}

func TestColNaturalDate(t *testing.T) {
	data := []map[string]any{
		{"time": int64(1708682400)}, // some timestamp
		{"time": "invalid"},
		{"other": 1},
	}
	got := ColNaturalDate(data, "time")
	if reflect.TypeOf(got[0]["time"]).Kind() != reflect.String {
		t.Errorf("Expected string for time, got %T", got[0]["time"])
	}
	if got[1]["time"] != nil {
		t.Errorf("Expected nil for invalid time, got %v", got[1]["time"])
	}
}

func TestColFilesize(t *testing.T) {
	data := []map[string]any{
		{"size": int64(1024)},
		{"size": int64(0)},
	}
	got := ColFilesize(data, "size")
	if got[0]["size"] != "1.0 KB" {
		t.Errorf("Expected 1.0 KB, got %v", got[0]["size"])
	}
	if got[1]["size"] != nil {
		t.Errorf("Expected nil for 0 size, got %v", got[1]["size"])
	}
}

func TestColDuration(t *testing.T) {
	data := []map[string]any{
		{"dur": 61},
		{"dur": 0},
	}
	got := ColDuration(data, "dur")
	if got[0]["dur"] != "1:01" {
		t.Errorf("Expected 1:01, got %v", got[0]["dur"])
	}
	if got[1]["dur"] != "" {
		t.Errorf("Expected empty string for 0 duration, got %v", got[1]["dur"])
	}
}
