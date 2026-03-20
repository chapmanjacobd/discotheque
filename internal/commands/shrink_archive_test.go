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
	os.MkdirAll(innerDir, 0o755)
	imagePath := filepath.Join(innerDir, "image.png")

	// Create a real PNG using magick
	if err := exec.Command("magick", "-size", "100x100", "xc:white", imagePath).Run(); err != nil {
		t.Fatalf("Failed to create test image with magick: %v", err)
	}

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
		// Mock lsar behavior for the test if it fails to detect media types from extensionless files
		// but here we gave it image.png, so it should be fine.
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
		result := processor.ExtractAndProcess(ctx, m, cfg, imageProc)

		if !result.Success {
			t.Fatalf("ExtractAndProcess failed: %v", result.Error)
		}

		// Check for success
		t.Logf("Result: SourcePath=%s, Outputs=%d, Success=%v", result.SourcePath, len(result.Outputs), result.Success)

		// Check if inner.zip was extracted and image.avif was created
		// The path should be: fixture.TempDir/outer.zip.extracted/inner.zip.extracted/image.avif
		if len(result.Outputs) == 0 {
			t.Fatalf("Expected at least one output")
		}
		avifPath := filepath.Join(result.Outputs[0].Path, "inner.zip.extracted", "image.avif")
		if _, err := os.Stat(avifPath); err != nil {
			t.Errorf("Expected avif file not found at %s: %v", avifPath, err)
			// Check directory content to see what happened
			filepath.Walk(result.Outputs[0].Path, func(path string, info os.FileInfo, err error) error {
				t.Logf("Found file: %s", path)
				return nil
			})
		}
	})
}
