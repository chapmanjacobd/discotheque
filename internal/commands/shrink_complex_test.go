package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/chapmanjacobd/discoteca/internal/db"
	"github.com/chapmanjacobd/discoteca/internal/testutils"
)

func TestTextProcessor_UpdateImageReferences(t *testing.T) {
	tmpDir := t.TempDir()
	htmlPath := filepath.Join(tmpDir, "index.html")
	content := `<html><body><img src="cover.jpg"> <a href="link.png"></a> <p>No change .txt</p></body></html>`
	os.WriteFile(htmlPath, []byte(content), 0o644)

	p := &TextProcessor{}
	p.updateReferencesInFile(htmlPath)

	newContent, _ := os.ReadFile(htmlPath)
	got := string(newContent)

	if !strings.Contains(got, "cover.avif") {
		t.Error("Expected cover.jpg to be replaced with cover.avif")
	}
	if !strings.Contains(got, "link.avif") {
		t.Error("Expected link.png to be replaced with link.avif")
	}
	if !strings.Contains(got, "No change .txt") {
		t.Error("Expected .txt to remain unchanged")
	}
}

func TestArchiveProcessor_EstimationLogic(t *testing.T) {
	// This test mocks the logic inside EstimateSizeForArchive by testing the core decision loop
	// if it were refactored into a testable pure function.
	// Since it's currently coupled with lsar, we'll document the expected behavior.
}

func TestShrinkCmd_MultipleOutputsFromZip(t *testing.T) {
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
	zipPath, _ := filepath.Abs(filepath.Join(tmpDir, "album.zip"))
	os.WriteFile(zipPath, make([]byte, 1000), 0o644)

	// Add to database so markDeleted has something to update
	_, err = sqlDB.Exec("INSERT INTO media (path, size, type) VALUES (?, ?, ?)", zipPath, 1000, "application/zip")
	if err != nil {
		t.Fatalf("Failed to insert zip into DB: %v", err)
	}

	m := ShrinkMedia{Path: zipPath, Size: 1000, MediaType: "Archived"}
	metrics := NewShrinkMetrics()

	// Mock processor that extracts 3 MP3s and deletes the ZIP
	mock := &MockProcessor{
		BaseProcessor: BaseProcessor{mediaType: "Archived"},
		ProcessFunc: func(ctx context.Context, media *ShrinkMedia, cfg *ProcessorConfig) ProcessResult {
			p1, _ := filepath.Abs(filepath.Join(tmpDir, "track1.mp3"))
			p2, _ := filepath.Abs(filepath.Join(tmpDir, "track2.mp3"))
			p3, _ := filepath.Abs(filepath.Join(tmpDir, "track3.mp3"))
			os.WriteFile(p1, make([]byte, 300), 0o644)
			os.WriteFile(p2, make([]byte, 300), 0o644)
			os.WriteFile(p3, make([]byte, 300), 0o644)

			return ProcessResult{
				SourcePath: media.Path,
				Outputs: []ProcessOutputFile{
					{Path: p1, Size: 300},
					{Path: p2, Size: 300},
					{Path: p3, Size: 300},
				},
				ToDelete: []string{media.Path},
				ToMove:   []string{p1, p2, p3},
				Success:  true,
			}
		},
	}
	registry := &ProcessorRegistry{processors: []MediaProcessor{mock}}

	cmd.processSingle(m, registry, &ProcessorConfig{}, metrics)

	// Verify all 3 files in DB
	var count int
	sqlDB.QueryRow("SELECT COUNT(*) FROM media WHERE path LIKE '%.mp3'").Scan(&count)
	if count != 3 {
		t.Errorf("Expected 3 MP3s in DB, got %d", count)
	}

	// Verify ZIP is gone from DB
	var timeDeleted int64
	err = sqlDB.QueryRow("SELECT time_deleted FROM media WHERE path = ?", zipPath).Scan(&timeDeleted)
	if err != nil {
		t.Errorf("Failed to query ZIP status in DB: %v", err)
	}
	if timeDeleted == 0 {
		t.Errorf("Original ZIP should be marked as deleted in DB (time_deleted was 0, path: %s)", zipPath)

		// Debug: list all files in DB
		rows, _ := sqlDB.Query("SELECT path, time_deleted FROM media")
		t.Log("Current database state:")
		for rows.Next() {
			var p string
			var td int64
			rows.Scan(&p, &td)
			t.Logf("  Path: %s, Deleted: %d", p, td)
		}
		rows.Close()
	}

	if _, err := os.Stat(zipPath); !os.IsNotExist(err) {
		t.Error("Original ZIP should be deleted from filesystem")
	}
}

func TestShrinkMetrics_ConcurrencyStress(t *testing.T) {
	m := NewShrinkMetrics()
	var wg sync.WaitGroup
	numGoroutines := 20
	iterations := 100

	for i := range numGoroutines {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			mediaType := "Video"
			if id%2 == 0 {
				mediaType = "Audio"
			}

			for j := range iterations {
				m.RecordStarted(mediaType, fmt.Sprintf("file-%d-%d", id, j))
				if j%2 == 0 {
					m.RecordSuccess(mediaType, 1000, 500, 1, 10)
				} else {
					m.RecordFailure(mediaType)
				}
			}
		}(i)
	}

	wg.Wait()

	statsVideo := m.GetStats("Video")
	expectedTotal := (numGoroutines / 2) * iterations
	if statsVideo.Total != expectedTotal {
		t.Errorf("Video Total mismatch: got %d, want %d", statsVideo.Total, expectedTotal)
	}

	expectedSuccess := expectedTotal / 2
	if statsVideo.Success != expectedSuccess {
		t.Errorf("Video Success mismatch: got %d, want %d", statsVideo.Success, expectedSuccess)
	}
}

func TestShrinkCmd_MoveBrokenMissingFile(t *testing.T) {
	fixture := testutils.Setup(t)
	defer fixture.Cleanup()

	tmpDir := t.TempDir()
	brokenPath := filepath.Join(tmpDir, "already_gone.mp4")
	// Note: file does NOT exist

	cmd := &ShrinkCmd{
		Databases:  []string{fixture.DBPath},
		MoveBroken: t.TempDir(),
	}

	m := ShrinkMedia{Path: brokenPath, Size: 100, MediaType: "Video"}
	metrics := NewShrinkMetrics()

	// Mock processor that says "delete it" but it's already gone
	mock := &MockProcessor{
		BaseProcessor: BaseProcessor{mediaType: "Video"},
		ProcessFunc: func(ctx context.Context, media *ShrinkMedia, cfg *ProcessorConfig) ProcessResult {
			return ProcessResult{
				SourcePath: media.Path,
				ToDelete:   []string{media.Path},
				Success:    false,
				Error:      fmt.Errorf("failed"),
			}
		},
	}
	registry := &ProcessorRegistry{processors: []MediaProcessor{mock}}

	// This should not panic or error out even though brokenPath is missing
	// because processSingle checks for existence before calling os.Remove
	// and handles the MoveBroken case gracefully.
	cmd.processSingle(m, registry, &ProcessorConfig{}, metrics)

	if stats := metrics.GetStats("Video"); stats.Failed != 1 {
		t.Errorf("Expected failure to be recorded, got %d", stats.Failed)
	}
}
