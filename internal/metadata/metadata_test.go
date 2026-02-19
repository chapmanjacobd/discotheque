package metadata

import (
	"context"
	"os"
	"reflect"
	"testing"

	"github.com/chapmanjacobd/discotheque/internal/utils"
)

func TestExtract(t *testing.T) {
	f, err := os.CreateTemp("", "metadata-test-*.mp4")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	f.WriteString("dummy data")
	f.Close()

	ctx := context.Background()
	res, err := Extract(ctx, f.Name(), false)
	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}

	if res.Media.Path != f.Name() {
		t.Errorf("Expected path %s, got %s", f.Name(), res.Media.Path)
	}
	if res.Media.Size.Int64 != 10 {
		t.Errorf("Expected size 10, got %d", res.Media.Size.Int64)
	}
}

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
		{[]string{"mp3", "", "mp3"}, []string{"mp3", ""}},
		{[]string{}, nil},
	}

	for _, tt := range tests {
		result := utils.Unique(tt.input)
		if !reflect.DeepEqual(result, tt.expected) {
			t.Errorf("utils.Unique(%v) = %v, want %v", tt.input, result, tt.expected)
		}
	}
}

func TestToNullHelpers(t *testing.T) {
	if got := utils.ToNullInt64(123); !got.Valid || got.Int64 != 123 {
		t.Errorf("utils.ToNullInt64(123) = %v", got)
	}
	if got := utils.ToNullInt64(0); got.Valid {
		t.Errorf("utils.ToNullInt64(0) should be invalid")
	}

	if got := utils.ToNullString("abc"); !got.Valid || got.String != "abc" {
		t.Errorf("utils.ToNullString(abc) = %v", got)
	}
	if got := utils.ToNullString(""); got.Valid {
		t.Errorf("utils.ToNullString(\"\") should be invalid")
	}

	if got := utils.ToNullFloat64(1.5); !got.Valid || got.Float64 != 1.5 {
		t.Errorf("utils.ToNullFloat64(1.5) = %v", got)
	}
	if got := utils.ToNullFloat64(0.0); got.Valid {
		t.Errorf("utils.ToNullFloat64(0.0) should be invalid")
	}
}
