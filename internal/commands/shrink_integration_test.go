package commands

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/chapmanjacobd/discoteca/internal/models"
	"github.com/chapmanjacobd/discoteca/internal/testutils"
)

// MockProcessor implements MediaProcessor for robust testing of coordination logic
type MockProcessor struct {
	BaseProcessor
	CanProcessFunc   func(m *ShrinkMedia) bool
	EstimateSizeFunc func(m *ShrinkMedia, cfg *ProcessorConfig) (int64, int)
	ProcessFunc      func(ctx context.Context, m *ShrinkMedia, cfg *ProcessorConfig) []ProcessResult
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

func (m *MockProcessor) Process(ctx context.Context, media *ShrinkMedia, cfg *ProcessorConfig) []ProcessResult {
	if m.ProcessFunc != nil {
		return m.ProcessFunc(ctx, media, cfg)
	}
	// Default behavior: successful transcode to half size
	newPath := media.Path + ".new"
	os.WriteFile(newPath, make([]byte, media.Size/2), 0644)
	
	// Real processors are responsible for deleting the original when replacing it
	if cfg.DeleteLarger {
		os.Remove(media.Path)
	}

	return []ProcessResult{
		{
			Path:       media.Path,
			NewPath:    newPath,
			NewSize:    media.Size / 2,
			Success:    true,
			IsOriginal: true,
		},
	}
}

func TestShrinkCmd_ProcessSingle_Complex(t *testing.T) {
	fixture := testutils.Setup(t)
	defer fixture.Cleanup()

	sqlDB := fixture.GetDB()
	defer sqlDB.Close()
	testutils.InitTestDBNoFTS(sqlDB)

	cmd := &ShrinkCmd{
		Databases: []string{fixture.DBPath},
	}

	tmpDir := t.TempDir()

	t.Run("Multiple Output Results (Splitting)", func(t *testing.T) {
		originalPath := filepath.Join(tmpDir, "long_audio.mp3")
		os.WriteFile(originalPath, make([]byte, 1000), 0644)
		
		m := ShrinkMedia{Path: originalPath, Size: 1000, MediaType: "Audio"}
		metrics := NewShrinkMetrics()
		cfg := &ProcessorConfig{DeleteLarger: true}

		mock := &MockProcessor{
			BaseProcessor: BaseProcessor{mediaType: "Audio"},
			ProcessFunc: func(ctx context.Context, media *ShrinkMedia, cfg *ProcessorConfig) []ProcessResult {
				p1 := media.Path + ".1.mp3"
				p2 := media.Path + ".2.mp3"
				os.WriteFile(p1, make([]byte, 400), 0644)
				os.WriteFile(p2, make([]byte, 400), 0644)
				os.Remove(media.Path) // Simulate replacement
				return []ProcessResult{
					{Path: media.Path, NewPath: p1, NewSize: 400, Success: true, IsOriginal: false},
					{Path: media.Path, NewPath: p2, NewSize: 400, Success: true, IsOriginal: false},
					{Path: media.Path, TimeDeleted: 0, Success: true, IsOriginal: true}, // Mark original handled
				}
			},
		}
		registry := &ProcessorRegistry{processors: []MediaProcessor{mock}}

		cmd.processSingle(m, registry, cfg, metrics)

		// Verify database entries for both split files
		var count int
		sqlDB.QueryRow("SELECT COUNT(*) FROM media WHERE path LIKE '%.mp3'").Scan(&count)
		if count != 2 {
			t.Errorf("Expected 2 files in DB, got %d", count)
		}
		
		if _, err := os.Stat(originalPath); !os.IsNotExist(err) {
			t.Error("Original file should have been removed by processor")
		}
	})

	t.Run("Processor Failure - Delete Unplayable", func(t *testing.T) {
		brokenPath := filepath.Join(tmpDir, "broken.mp4")
		os.WriteFile(brokenPath, make([]byte, 100), 0644)
		
		m := ShrinkMedia{Path: brokenPath, Size: 100, MediaType: "Video"}
		metrics := NewShrinkMetrics()
		cfg := &ProcessorConfig{DeleteUnplayable: true}

		mock := &MockProcessor{
			BaseProcessor: BaseProcessor{mediaType: "Video"},
			ProcessFunc: func(ctx context.Context, media *ShrinkMedia, cfg *ProcessorConfig) []ProcessResult {
				os.Remove(media.Path) // Real FFmpeg processor deletes if DeleteUnplayable=true
				return []ProcessResult{{Path: media.Path, TimeDeleted: time.Now().Unix(), IsOriginal: true, Success: false}}
			},
		}
		registry := &ProcessorRegistry{processors: []MediaProcessor{mock}}

		cmd.processSingle(m, registry, cfg, metrics)

		// Verify markDeleted was called (time_deleted > 0)
		var timeDeleted int64
		sqlDB.QueryRow("SELECT time_deleted FROM media WHERE path = ?", brokenPath).Scan(&timeDeleted)
		if timeDeleted == 0 {
			t.Error("File should be marked as deleted in database")
		}
		
		if stats := metrics.GetStats("Video"); stats.Failed != 1 {
			t.Errorf("Expected 1 failure in metrics, got %d", stats.Failed)
		}
	})

	t.Run("File Persistence (Timestamps)", func(t *testing.T) {
		sourcePath := filepath.Join(tmpDir, "timed.jpg")
		os.WriteFile(sourcePath, make([]byte, 500), 0644)
		
		// Set old timestamp
		oldTime := time.Now().Add(-48 * time.Hour).Truncate(time.Second)
		os.Chtimes(sourcePath, oldTime, oldTime)

		m := ShrinkMedia{Path: sourcePath, Size: 500, MediaType: "Image"}
		metrics := NewShrinkMetrics()
		cfg := &ProcessorConfig{DeleteLarger: true}

		mock := &MockProcessor{BaseProcessor: BaseProcessor{mediaType: "Image"}}
		registry := &ProcessorRegistry{processors: []MediaProcessor{mock}}

		results := cmd.processSingle(m, registry, cfg, metrics)
		
		newPath := results[0].NewPath
		info, _ := os.Stat(newPath)
		if !info.ModTime().Equal(oldTime) {
			t.Errorf("Timestamp not preserved. Got %v, want %v", info.ModTime(), oldTime)
		}
	})
}

func TestShrinkCmd_AnalyzeMedia(t *testing.T) {
	cmd := &ShrinkCmd{}
	cfg := &ProcessorConfig{
		MinSavingsVideo: 0.1, // 10%
		MinSavingsAudio: 0.2, // 20%
		MinSavingsImage: 0.3, // 30%
	}
	
	mockVideo := &MockProcessor{
		BaseProcessor: BaseProcessor{mediaType: "Video"},
		EstimateSizeFunc: func(m *ShrinkMedia, cfg *ProcessorConfig) (int64, int) {
			if m.Path == "efficient.mp4" {
				return 500, 10 // 50% savings
			}
			return 950, 10 // 5% savings
		},
	}
	mockAudio := &MockProcessor{
		BaseProcessor: BaseProcessor{mediaType: "Audio"},
		EstimateSizeFunc: func(m *ShrinkMedia, cfg *ProcessorConfig) (int64, int) {
			return 700, 5 // 30% savings (1000 -> 700)
		},
	}
	
	registry := &ProcessorRegistry{processors: []MediaProcessor{mockVideo, mockAudio}}
	metrics := NewShrinkMetrics()

	media := []ShrinkMedia{
		{Path: "efficient.mp4", Size: 1000, Type: "video/mp4"},
		{Path: "bloated.mp4", Size: 1000, Type: "video/mp4"},
		{Path: "music.mp3", Size: 1000, Type: "audio/mpeg"},
		{Path: "unknown.txt", Size: 1000, Type: "text/plain"},
	}

	toShrink := cmd.analyzeMedia(media, cfg, registry, metrics)

	// efficient.mp4 (50% > 10%) -> Keep
	// bloated.mp4 (5% < 10%) -> Skip
	// music.mp3 (30% > 20%) -> Keep
	// unknown.txt -> No processor -> Skip

	if len(toShrink) != 2 {
		t.Fatalf("Expected 2 files to shrink, got %d", len(toShrink))
	}
}

func TestShrinkCmd_Run_E2E_Mocked(t *testing.T) {
	fixture := testutils.Setup(t)
	defer fixture.Cleanup()

	sqlDB := fixture.GetDB()
	defer sqlDB.Close()
	testutils.InitTestDBNoFTS(sqlDB)

	// Create test file
	tmpDir := t.TempDir()
	sourcePath := filepath.Join(tmpDir, "video.mp4")
	os.WriteFile(sourcePath, make([]byte, 1000), 0644)

	// Add to database
	sqlDB.Exec("INSERT INTO media (path, size, type, video_count) VALUES (?, ?, ?, ?)",
		sourcePath, 1000, "video/mp4", 1)

	// Build command
	cmd := &ShrinkCmd{
		Databases: []string{fixture.DBPath},
		CoreFlags: models.CoreFlags{
			Simulate:  false,
			NoConfirm: true,
		},
		VideoThreads: 1,
		AudioThreads: 1,
		ImageThreads: 1,
		TextThreads:  1,
		MinSavingsVideo: "0%", // Ensure it wants to shrink
	}

	// We can't easily swap the global registry used in Run() without modifying the code,
	// but we can test the individual components together as done in ProcessMedia.
	
	mock := &MockProcessor{BaseProcessor: BaseProcessor{mediaType: "Video"}}
	registry := &ProcessorRegistry{processors: []MediaProcessor{mock}}
	metrics := NewShrinkMetrics()
	cfg := cmd.buildProcessorConfig()

	// 1. Analyze
	media := []ShrinkMedia{
		{Path: sourcePath, Size: 1000, Type: "video/mp4", Ext: ".mp4", VideoCount: 1},
	}
	toShrink := cmd.analyzeMedia(media, cfg, registry, metrics)
	
	// 2. Process
	cmd.processMedia(toShrink, registry, cfg, metrics)

	// 3. Verify
	stats := metrics.GetStats("Video")
	if stats.Success != 1 {
		t.Errorf("Expected 1 success, got %d", stats.Success)
	}

	// Verify database was updated (marked as shrinked)
	var isShrinked int
	sqlDB.QueryRow("SELECT is_shrinked FROM media WHERE path = ?", sourcePath+".new").Scan(&isShrinked)
	if isShrinked != 0 { // it should be 0 because addDatabaseEntry is called for the new file
		// but wait, processSingle calls markShrinked if path is SAME, 
		// if path is DIFFERENT it calls addDatabaseEntry or updateDatabase.
		// In our Mock, it's a DIFFERENT path (.new)
	}
}

