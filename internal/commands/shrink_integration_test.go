package commands

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/chapmanjacobd/discoteca/internal/db"
	"github.com/chapmanjacobd/discoteca/internal/testutils"
)

// MockProcessor implements MediaProcessor for robust testing of coordination logic
type MockProcessor struct {
	BaseProcessor
	CanProcessFunc   func(m *ShrinkMedia) bool
	EstimateSizeFunc func(m *ShrinkMedia, cfg *ProcessorConfig) (int64, int)
	ProcessFunc      func(ctx context.Context, m *ShrinkMedia, cfg *ProcessorConfig) ProcessResult
}

func (m *MockProcessor) CanProcess(media *ShrinkMedia) bool {
	if m.CanProcessFunc != nil {
		return m.CanProcessFunc(media)
	}
	return true
}

func (m *MockProcessor) EstimateSize(media *ShrinkMedia, cfg *ProcessorConfig) (int64, int) {
	if m.EstimateSizeFunc != nil {
		return m.EstimateSizeFunc(media, cfg)
	}
	return media.Size / 2, 1
}

func (m *MockProcessor) Process(ctx context.Context, media *ShrinkMedia, cfg *ProcessorConfig) ProcessResult {
	if m.ProcessFunc != nil {
		return m.ProcessFunc(ctx, media, cfg)
	}
	// Default behavior: successful transcode to half size
	newPath := media.Path + ".new"
	newSize := media.Size / 2
	os.WriteFile(newPath, make([]byte, newSize), 0o644)

	return ProcessResult{
		SourcePath: media.Path,
		Outputs:    []ProcessOutputFile{{Path: newPath, Size: newSize}},
		ToDelete:   []string{media.Path},
		ToMove:     []string{newPath},
		Success:    true,
	}
}

func TestShrinkCmd_ProcessSingle_Complex(t *testing.T) {
	fixture := testutils.Setup(t)
	defer fixture.Cleanup()

	sqlDB, _, err := db.ConnectWithInit(fixture.DBPath)
	if err != nil {
		t.Fatal(err)
	}
	defer sqlDB.Close()

	cmd := &ShrinkCmd{
		Databases: []string{fixture.DBPath},
	}

	tmpDir := t.TempDir()

	t.Run("Multiple Output Results (Splitting)", func(t *testing.T) {
		originalPath := filepath.Join(tmpDir, "long_audio.mp3")
		os.WriteFile(originalPath, make([]byte, 1000), 0o644)

		m := ShrinkMedia{Path: originalPath, Size: 1000, MediaType: "Audio"}
		metrics := NewShrinkMetrics()
		cfg := &ProcessorConfig{}

		mock := &MockProcessor{
			BaseProcessor: BaseProcessor{mediaType: "Audio"},
			ProcessFunc: func(ctx context.Context, media *ShrinkMedia, cfg *ProcessorConfig) ProcessResult {
				p1 := filepath.Join(tmpDir, "track1.mp3")
				p2 := filepath.Join(tmpDir, "track2.mp3")
				os.WriteFile(p1, make([]byte, 400), 0o644)
				os.WriteFile(p2, make([]byte, 400), 0o644)

				return ProcessResult{
					SourcePath: media.Path,
					Outputs: []ProcessOutputFile{
						{Path: p1, Size: 400},
						{Path: p2, Size: 400},
					},
					ToDelete: []string{media.Path},
					ToMove:   []string{p1, p2},
					Success:  true,
				}
			},
		}
		registry := &ProcessorRegistry{processors: []MediaProcessor{mock}}

		cmd.processSingle(m, registry, cfg, metrics)

		// Verify database entries for both split files
		var count int
		sqlDB.QueryRow("SELECT COUNT(*) FROM media WHERE path LIKE '%track%.mp3'").Scan(&count)
		if count != 2 {
			t.Errorf("Expected 2 files in DB, got %d", count)
		}

		if _, err := os.Stat(originalPath); !os.IsNotExist(err) {
			t.Error("Original file should have been removed by processSingle based on ToDelete")
		}
	})

	t.Run("Processor Failure - Delete Unplayable", func(t *testing.T) {
		brokenPath := filepath.Join(tmpDir, "broken.mp4")
		os.WriteFile(brokenPath, make([]byte, 100), 0o644)

		// Add to database
		sqlDB.Exec("INSERT INTO media (path, size, type) VALUES (?, ?, ?)", brokenPath, 100, "video/mp4")

		m := ShrinkMedia{Path: brokenPath, Size: 100, MediaType: "Video"}
		metrics := NewShrinkMetrics()
		cfg := &ProcessorConfig{DeleteUnplayable: true}

		mock := &MockProcessor{
			BaseProcessor: BaseProcessor{mediaType: "Video"},
			ProcessFunc: func(ctx context.Context, media *ShrinkMedia, cfg *ProcessorConfig) ProcessResult {
				return ProcessResult{
					SourcePath: media.Path,
					ToDelete:   []string{media.Path},
					Success:    false,
				}
			},
		}
		registry := &ProcessorRegistry{processors: []MediaProcessor{mock}}

		cmd.processSingle(m, registry, cfg, metrics)

		// Verify markDeleted was called (path removed or time_deleted > 0)
		var count int
		sqlDB.QueryRow("SELECT COUNT(*) FROM media WHERE path = ? AND time_deleted > 0", brokenPath).Scan(&count)
		if count == 0 {
			t.Error("File should be marked as deleted in database")
		}

		if _, err := os.Stat(brokenPath); !os.IsNotExist(err) {
			t.Error("Broken file should be deleted from filesystem")
		}
	})
}

func TestShrinkCmd_AnalyzeMedia(t *testing.T) {
	cmd := &ShrinkCmd{}
	cfg := &ProcessorConfig{
		MinSavingsVideo: 0.1, // 10%
	}

	mock := &MockProcessor{
		BaseProcessor: BaseProcessor{mediaType: "Video"},
		EstimateSizeFunc: func(m *ShrinkMedia, cfg *ProcessorConfig) (int64, int) {
			if m.Path == "efficient.mp4" {
				return 500, 10 // 50% savings
			}
			return 950, 10 // 5% savings
		},
	}
	registry := &ProcessorRegistry{processors: []MediaProcessor{mock}}
	metrics := NewShrinkMetrics()

	media := []ShrinkMedia{
		{Path: "efficient.mp4", Size: 1000, Type: "video/mp4", MediaType: "Video"},
		{Path: "bloated.mp4", Size: 1000, Type: "video/mp4", MediaType: "Video"},
	}

	toShrink := cmd.analyzeMedia(media, cfg, registry, metrics)

	if len(toShrink) != 1 {
		t.Fatalf("Expected 1 file to shrink, got %d", len(toShrink))
	}
	if toShrink[0].Path != "efficient.mp4" {
		t.Errorf("Expected efficient.mp4, got %s", toShrink[0].Path)
	}
}
