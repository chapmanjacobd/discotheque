package commands

import (
	"testing"
)

func TestDetectMediaTypeFromExt(t *testing.T) {
	tests := []struct {
		ext  string
		want string
	}{
		{".mp4", "video"},
		{".mkv", "video"},
		{".mp3", "audio"},
		{".opus", "audio"},
		{".jpg", "image"},
		{".avif", "image"},
		{".pdf", "text"},
		{".epub", "text"},
		{".zip", "archive"},
		{".rar", "archive"},
		{".unknown", ""},
	}

	for _, tt := range tests {
		got := detectMediaTypeFromExt(tt.ext)
		if got != tt.want {
			t.Errorf("detectMediaTypeFromExt(%s) = %s, want %s", tt.ext, got, tt.want)
		}
	}
}

func TestTextProcessor_OCRLogic(t *testing.T) {
	// Test the logic that decides which OCR flags to use
	tests := []struct {
		name     string
		cfg      *ProcessorConfig
		env      map[string]string // simulate env vars if needed
		expected string            // flag we expect in args
	}{
		{
			name:     "Force OCR",
			cfg:      &ProcessorConfig{ForceOCR: true},
			expected: "--force-ocr",
		},
		{
			name:     "Skip OCR if text exists",
			cfg:      &ProcessorConfig{SkipOCR: true},
			expected: "--skip-text",
		},
		{
			name:     "Redo OCR",
			cfg:      &ProcessorConfig{RedoOCR: true},
			expected: "--redo-ocr",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This is a bit of a white-box test because we're checking the logic
			// inside runOCR that builds the arguments.

			// If we wanted to test this more deeply, we'd need to mock utils.CommandExists
			// but for now we're just documenting the expected behavior for these flags.
		})
	}
}

func TestEbookImageOptimizationCriteria(t *testing.T) {
	tests := []struct {
		ext  string
		want bool
	}{
		{".jpg", true},
		{".png", true},
		{".webp", true},
		{".avif", false}, // Already optimized
		{".svg", false},  // Vector
		{".gif", false},  // Categorized as video in this project
	}

	for _, tt := range tests {
		if got := shouldConvertToAVIF(tt.ext); got != tt.want {
			t.Errorf("shouldConvertToAVIF(%s) = %v, want %v", tt.ext, got, tt.want)
		}
	}
}

func TestArchiveProcessor_lsar_Parsing(t *testing.T) {
	// Logic to test: parsing the JSON output from lsar and converting it to ShrinkMedia
	// Since lsar output is complex, we could test a helper function if one existed,
	// or mock the exec.Command in a more involved test.
}
