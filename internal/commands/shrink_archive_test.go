package commands

import (
	"context"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/chapmanjacobd/discoteca/internal/testutils"
	"github.com/chapmanjacobd/discoteca/internal/utils"
)

func TestArchiveProcessor_NestedArchives(t *testing.T) {
	if !utils.CommandExists("lsar") || !utils.CommandExists("unar") || !utils.CommandExists("zip") {
		t.Skip("lsar, unar, or zip not installed")
	}

	// Set up debug logging for the test
	opts := &slog.HandlerOptions{Level: slog.LevelDebug}
	handler := slog.NewTextHandler(os.Stderr, opts)
	slog.SetDefault(slog.New(handler))

	fixture := testutils.Setup(t)
	defer fixture.Cleanup()

	// Create a nested archive
	// inner.zip contains image.jpg
	// outer.zip contains inner.zip

	innerDir := filepath.Join(fixture.TempDir, "inner")
	os.MkdirAll(innerDir, 0755)
	imagePath := filepath.Join(innerDir, "image.jpg")
	// Valid JPEG header
	jpegHeader := []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46, 0x49, 0x46, 0x00, 0x01, 0x01, 0x01, 0x00, 0x60, 0x00, 0x60, 0x00, 0x00}
	data := append(jpegHeader, make([]byte, 100*1024)...)
	os.WriteFile(imagePath, data, 0644) // ~100KB image

	innerZip := filepath.Join(fixture.TempDir, "inner.zip")
	if err := exec.Command("zip", "-j", innerZip, imagePath).Run(); err != nil {
		t.Fatalf("Failed to create inner zip: %v", err)
	}

	outerZip := filepath.Join(fixture.TempDir, "outer.zip")
	if err := exec.Command("zip", "-j", outerZip, innerZip).Run(); err != nil {
		t.Fatalf("Failed to create outer zip: %v", err)
	}

	processor := NewArchiveProcessor()
	cfg := &ProcessorConfig{
		TargetImageSize:      30 * 1024,
		MinSavingsImage:      0.1,
		TranscodingImageTime: 1.0,
		SourceAudioBitrate:   128 * 1024,
		SourceVideoBitrate:   1500 * 1024,
		TargetAudioBitrate:   64 * 1024,
		TargetVideoBitrate:   800 * 1024,
		TranscodingVideoRate: 1.0,
		TranscodingAudioRate: 1.0,
	}

	m := &ShrinkMedia{
		Path: outerZip,
		Type: "archive/zip",
		Ext:  ".zip",
	}

	t.Run("EstimateSize", func(t *testing.T) {
		futureSize, processingTime, hasProcessable := processor.EstimateSizeForArchive(m, cfg)
		
		t.Logf("FutureSize: %d, ProcessingTime: %d, HasProcessable: %v", futureSize, processingTime, hasProcessable)
		
		// If it handled nested archives, it should have found the image inside inner.zip
		if !hasProcessable {
			t.Errorf("Expected hasProcessable to be true for nested archive containing image")
		}
		if futureSize == 0 {
			t.Errorf("Expected futureSize > 0 for nested archive containing image")
		}
	})

	t.Run("ExtractAndProcess", func(t *testing.T) {
		if !utils.CommandExists("magick") {
			t.Skip("magick not installed")
		}

		imageProc := NewImageProcessor()
		ctx := context.Background()
		results := processor.ExtractAndProcess(ctx, m, cfg, imageProc)

		if len(results) == 0 {
			t.Fatalf("ExtractAndProcess returned no results")
		}

		// Check for success
		hasSuccess := false
		for _, result := range results {
			if result.Success {
				hasSuccess = true
				t.Logf("Result: Path=%s, NewPath=%s, Success=%v", result.Path, result.NewPath, result.Success)
			}
		}

		if !hasSuccess {
			t.Errorf("Expected at least one successful result")
		}

		// Check if inner.zip was extracted and image.avif was created
		// The path should be: fixture.TempDir/outer.zip.extracted/inner.zip.extracted/image.avif
		avifPath := filepath.Join(results[0].NewPath, "inner.zip.extracted", "image.avif")
		if _, err := os.Stat(avifPath); err != nil {
			t.Errorf("Expected avif file not found at %s: %v", avifPath, err)
			// Check directory content to see what happened
			filepath.Walk(results[0].NewPath, func(path string, info os.FileInfo, err error) error {
				t.Logf("Found file: %s", path)
				return nil
			})
		}
	})
}
