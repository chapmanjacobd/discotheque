package commands

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/chapmanjacobd/discoteca/internal/testutils"
)

func TestLoadMediaFromDB(t *testing.T) {
	fixture := testutils.Setup(t)
	defer fixture.Cleanup()

	sqlDB := fixture.GetDB()
	defer sqlDB.Close()

	// Initialize the DB schema (without FTS5 to avoid issues if not compiled with it)
	if err := testutils.InitTestDBNoFTS(sqlDB); err != nil {
		t.Fatal(err)
	}

	// Insert test data
	_, err := sqlDB.Exec(`
		INSERT INTO media (path, size, time_deleted, is_shrinked, type)
		VALUES 
			('active.mp4', 1000, 0, 0, 'video/mp4'),
			('deleted.mp4', 1000, 123456, 0, 'video/mp4'),
			('shrinked.mp4', 1000, 0, 1, 'video/mp4'),
			('empty.mp4', 0, 0, 0, 'video/mp4')
	`)
	if err != nil {
		t.Fatal(err)
	}

	c := &ShrinkCmd{ForceReshrink: false}

	t.Run("Default (no ForceReshrink)", func(t *testing.T) {
		media, err := c.loadMediaFromDB(sqlDB)
		if err != nil {
			t.Fatal(err)
		}
		if len(media) != 1 {
			t.Errorf("Expected 1 file, got %d", len(media))
		}
		if media[0].Path != "active.mp4" {
			t.Errorf("Expected active.mp4, got %s", media[0].Path)
		}
	})

	t.Run("ForceReshrink", func(t *testing.T) {
		c.ForceReshrink = true
		media, err := c.loadMediaFromDB(sqlDB)
		if err != nil {
			t.Fatal(err)
		}
		// Should include both active and already shrinked
		if len(media) != 2 {
			t.Errorf("Expected 2 files, got %d", len(media))
		}
	})
}

func TestSortByEfficiency(t *testing.T) {
	c := &ShrinkCmd{}
	media := []ShrinkMedia{
		{Path: "low", Savings: 100, ProcessingTime: 100},  // Ratio: 1
		{Path: "high", Savings: 1000, ProcessingTime: 10}, // Ratio: 100
		{Path: "mid", Savings: 500, ProcessingTime: 50},   // Ratio: 10
	}

	c.sortByEfficiency(media)

	if media[0].Path != "high" {
		t.Errorf("Expected first element to be 'high', got %s", media[0].Path)
	}
	if media[1].Path != "mid" {
		t.Errorf("Expected second element to be 'mid', got %s", media[1].Path)
	}
	if media[2].Path != "low" {
		t.Errorf("Expected third element to be 'low', got %s", media[2].Path)
	}
}

func TestFilterByContinueFrom(t *testing.T) {
	c := &ShrinkCmd{ContinueFrom: "file2.mp4"}
	media := []ShrinkMedia{
		{Path: "file1.mp4"},
		{Path: "file2.mp4"},
		{Path: "file3.mp4"},
	}

	filtered := c.applyContinueFrom(media)

	if len(filtered) != 2 {
		t.Fatalf("Expected 2 files, got %d", len(filtered))
	}
	if filtered[0].Path != "file2.mp4" {
		t.Errorf("Expected first file to be file2.mp4, got %s", filtered[0].Path)
	}
	if filtered[1].Path != "file3.mp4" {
		t.Errorf("Expected second file to be file3.mp4, got %s", filtered[1].Path)
	}
}

func TestFilterByTools(t *testing.T) {
	c := &ShrinkCmd{}
	media := []ShrinkMedia{
		{Path: "video.mp4", Type: "video/mp4", Ext: ".mp4", VideoCount: 1},
		{Path: "image.jpg", Type: "image/jpeg", Ext: ".jpg"},
		{Path: "text.pdf", Type: "application/pdf", Ext: ".pdf"},
		{Path: "archive.zip", Type: "application/zip", Ext: ".zip"},
	}

	// Case 1: All tools missing
	toolsNone := InstalledTools{}
	filtered := c.filterByTools(media, toolsNone)
	if len(filtered) != 0 {
		t.Errorf("Expected 0 files when no tools are installed, got %d", len(filtered))
	}

	// Case 2: Only FFmpeg
	toolsFFmpeg := InstalledTools{FFmpeg: true}
	filtered = c.filterByTools(media, toolsFFmpeg)
	if len(filtered) != 1 || filtered[0].Path != "video.mp4" {
		t.Errorf("Expected only video.mp4 when only FFmpeg is installed")
	}

	// Case 3: Only ImageMagick
	toolsMagick := InstalledTools{ImageMagick: true}
	filtered = c.filterByTools(media, toolsMagick)
	if len(filtered) != 1 || filtered[0].Path != "image.jpg" {
		t.Errorf("Expected only image.jpg when only ImageMagick is installed")
	}
}

func TestApplyTimestamps(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "shrink_test_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	filePath := filepath.Join(tmpDir, "file.txt")
	if err := os.WriteFile(filePath, []byte("test"), 0o644); err != nil {
		t.Fatal(err)
	}

	atime := time.Now().Add(-10 * time.Hour).Truncate(time.Second)
	mtime := time.Now().Add(-5 * time.Hour).Truncate(time.Second)

	applyTimestamps(filePath, atime, mtime)

	info, err := os.Stat(filePath)
	if err != nil {
		t.Fatal(err)
	}

	// Some filesystems might have lower precision, but we truncated to second
	if !info.ModTime().Equal(mtime) {
		t.Errorf("Expected mtime %v, got %v", mtime, info.ModTime())
	}
}

func TestShouldShrink(t *testing.T) {
	cfg := &ProcessorConfig{
		MinSavingsVideo: 0.05, // 5%
		MinSavingsAudio: 0.10, // 10%
		MinSavingsImage: 0.15, // 15%
	}

	tests := []struct {
		name       string
		mediaType  string
		size       int64
		futureSize int64
		want       bool
	}{
		{
			name:       "Video - exactly 5% savings - should NOT shrink",
			mediaType:  "Video",
			size:       105,
			futureSize: 100,
			want:       false,
		},
		{
			name:       "Video - more than 5% savings - should shrink",
			mediaType:  "Video",
			size:       106,
			futureSize: 100,
			want:       true,
		},
		{
			name:       "Audio - exactly 10% savings - should NOT shrink",
			mediaType:  "Audio",
			size:       110,
			futureSize: 100,
			want:       false,
		},
		{
			name:       "Image - more than 15% savings - should shrink",
			mediaType:  "Image",
			size:       116,
			futureSize: 100,
			want:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &ShrinkMedia{MediaType: tt.mediaType, Size: tt.size}
			if got := ShouldShrink(m, tt.futureSize, cfg); got != tt.want {
				t.Errorf("ShouldShrink() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestImageMagickErrorCategorization(t *testing.T) {
	unsupportedLog := []string{"magick: no decode delegate for this image format"}
	fileErrorLog := []string{"magick: unable to open image: No such file"}
	envErrorLog := []string{"magick: cache resources exhausted"}

	if !isImageMagickUnsupportedError(unsupportedLog) {
		t.Error("Expected unsupported error")
	}
	if !isImageMagickFileError(fileErrorLog) {
		t.Error("Expected file error")
	}
	if !isImageMagickEnvironmentError(envErrorLog) {
		t.Error("Expected environment error")
	}
}

func TestIsAnimationFromProbe(t *testing.T) {
	p := &FFmpegProcessor{}

	// Case 1: Audio stream present -> Animated
	probeWithAudio := &FFProbeResult{
		AudioStreams: []FFProbeStream{{Index: 1}},
	}
	isAnimated := p.isAnimationFromProbe(probeWithAudio)
	if isAnimated == nil || !*isAnimated {
		t.Error("Expected animated because of audio streams")
	}

	// Case 2: No audio, multiple video frames -> Animated
	probeWithFrames := &FFProbeResult{
		VideoStreams: []FFProbeStream{{Index: 0, NbFrames: "10"}},
	}
	isAnimated = p.isAnimationFromProbe(probeWithFrames)
	if isAnimated == nil || !*isAnimated {
		t.Error("Expected animated because of multiple video frames")
	}

	// Case 3: No audio, single video frame -> Static
	probeStatic := &FFProbeResult{
		VideoStreams: []FFProbeStream{{Index: 0, NbFrames: "1"}},
	}
	isAnimated = p.isAnimationFromProbe(probeStatic)
	if isAnimated == nil || *isAnimated {
		t.Error("Expected static because of single video frame")
	}
}

func TestBuildScaleFilter(t *testing.T) {
	p := &FFmpegProcessor{
		config: &ProcessorConfig{
			MaxVideoWidth:   1000,
			MaxVideoHeight:  1000,
			MaxWidthBuffer:  0.0,
			MaxHeightBuffer: 0.0,
		},
	}

	// Case 1: Normal scaling (width > MaxVideoWidth)
	filters := p.buildScaleFilter("", 2000, 1000)
	if len(filters) == 0 || filters[0] != "scale=1000:-2" {
		t.Errorf("Expected scale=1000:-2, got %v", filters)
	}

	// Case 2: Normal scaling (height > MaxVideoHeight)
	filters = p.buildScaleFilter("", 1000, 2000)
	if len(filters) == 0 || filters[0] != "scale=-2:1000" {
		t.Errorf("Expected scale=-2:1000, got %v", filters)
	}

	// Case 3: No scaling needed (within limits)
	filters = p.buildScaleFilter("", 500, 500)
	if len(filters) == 0 || !strings.Contains(filters[0], "pad=") {
		t.Errorf("Expected pad= filter when no scaling needed, got %v", filters)
	}

	// Case 4: SBS scaling
	filters = p.buildScaleFilter("sbs", 3000, 1000) // 1500 per eye > 1000
	if len(filters) == 0 || filters[0] != "scale=2000:-2" {
		t.Errorf("Expected scale=2000:-2 for SBS, got %v", filters)
	}
}

func TestArchiveProcessorCanProcess(t *testing.T) {
	p := NewArchiveProcessor()

	tests := []struct {
		name  string
		ext   string
		mtype string
		want  bool
	}{
		{"Zip by extension", ".zip", "application/zip", true},
		{"Rar by extension", ".rar", "application/x-rar", true},
		{"7z by extension", ".7z", "application/x-7z-compressed", true},
		{"Random file", ".txt", "text/plain", false},
		{"Archive type", ".bin", "archive/custom", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &ShrinkMedia{Ext: tt.ext, Type: tt.mtype}
			if got := p.CanProcess(m); got != tt.want {
				t.Errorf("CanProcess() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTextProcessorCanProcess(t *testing.T) {
	p := NewTextProcessor()

	tests := []struct {
		name string
		ext  string
		want bool
	}{
		{"PDF", ".pdf", true},
		{"EPUB", ".epub", true},
		{"MOBI", ".mobi", true},
		{"Image", ".jpg", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &ShrinkMedia{Ext: tt.ext}
			if got := p.CanProcess(m); got != tt.want {
				t.Errorf("CanProcess() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBuildFFmpegArgs(t *testing.T) {
	p := &FFmpegProcessor{
		config: &ProcessorConfig{
			TargetVideoBitrate: 800000,
			TargetAudioBitrate: 128000,
			Preset:             "7",
			CRF:                "40",
		},
	}

	probe := &FFProbeResult{}
	videoStream := &FFProbeStream{CodecType: "video", Width: 1920, Height: 1080}
	audioStream := &FFProbeStream{CodecType: "audio"}

	t.Run("Default (Video + Audio)", func(t *testing.T) {
		args := p.buildFFmpegArgs("in.mp4", "out.mkv", probe, videoStream, audioStream, nil, nil)
		argStr := strings.Join(args, " ")
		if !strings.Contains(argStr, "-c:v libsvtav1") {
			t.Error("Expected libsvtav1 video codec")
		}
		if !strings.Contains(argStr, "-c:a libopus") {
			t.Error("Expected libopus audio codec")
		}
	})

	t.Run("AudioOnly", func(t *testing.T) {
		p.config.AudioOnly = true
		args := p.buildFFmpegArgs("in.mp4", "out.mka", probe, videoStream, audioStream, nil, nil)
		argStr := strings.Join(args, " ")
		if strings.Contains(argStr, "-c:v libsvtav1") {
			t.Error("Did NOT expect libsvtav1 video codec in AudioOnly mode")
		}
		p.config.AudioOnly = false // reset
	})

	t.Run("VideoOnly", func(t *testing.T) {
		p.config.VideoOnly = true
		// Note: Current code doesn't discard audio when VideoOnly is true
		// but it does discard video when AudioOnly is true.
		args := p.buildFFmpegArgs("in.mp4", "out.mkv", probe, videoStream, audioStream, nil, nil)
		argStr := strings.Join(args, " ")
		if !strings.Contains(argStr, "-c:v libsvtav1") {
			t.Error("Expected libsvtav1 video codec")
		}
		p.config.VideoOnly = false // reset
	})
}
