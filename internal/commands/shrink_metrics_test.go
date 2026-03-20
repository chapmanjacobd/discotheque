package commands

import (
	"testing"
	"time"
)

func TestMediaTypeStats(t *testing.T) {
	stats := &MediaTypeStats{
		Total:          10,
		Processed:      5,
		Success:        4,
		Failed:         1,
		Skipped:        2,
		TotalSize:      1000,
		FutureSize:     600,
		TotalTime:      100,
		TotalDuration:  200,
	}

	if got := stats.SuccessRate(); got != 80.0 {
		t.Errorf("SuccessRate() = %f, want 80.0", got)
	}

	if got := stats.SpaceSaved(); got != 400 {
		t.Errorf("SpaceSaved() = %d, want 400", got)
	}

	if got := stats.AvgProcessingTime(); got != 20 {
		t.Errorf("AvgProcessingTime() = %d, want 20", got)
	}

	if got := stats.SpeedRatio(); got != 2.0 {
		t.Errorf("SpeedRatio() = %f, want 2.0", got)
	}

	// Test zero processed
	emptyStats := &MediaTypeStats{}
	if got := emptyStats.SuccessRate(); got != 0 {
		t.Errorf("Empty SuccessRate() = %f, want 0", got)
	}
	if got := emptyStats.AvgProcessingTime(); got != 0 {
		t.Errorf("Empty AvgProcessingTime() = %d, want 0", got)
	}
	if got := emptyStats.SpeedRatio(); got != 0 {
		t.Errorf("Empty SpeedRatio() = %f, want 0", got)
	}
}

func TestShrinkMetricsAggregation(t *testing.T) {
	m := NewShrinkMetrics()

	// Record some activity for "Video"
	m.RecordStarted("Video", "file1.mp4")
	m.RecordSuccess("Video", 1000, 500, 10, 20)
	m.RecordStarted("Video", "file2.mp4")
	m.RecordFailure("Video")
	m.RecordSkipped("Video")

	// Record some activity for "Audio"
	m.RecordStarted("Audio", "music.mp3")
	m.RecordSuccess("Audio", 500, 250, 5, 50)

	statsVideo := m.GetStats("Video")
	if statsVideo == nil {
		t.Fatal("Expected Video stats")
	}
	if statsVideo.Total != 2 || statsVideo.Processed != 2 || statsVideo.Success != 1 || statsVideo.Failed != 1 || statsVideo.Skipped != 1 {
		t.Errorf("Video stats mismatch: %+v", statsVideo)
	}

	statsAudio := m.GetStats("Audio")
	if statsAudio == nil {
		t.Fatal("Expected Audio stats")
	}
	if statsAudio.Total != 1 || statsAudio.Processed != 1 || statsAudio.Success != 1 {
		t.Errorf("Audio stats mismatch: %+v", statsAudio)
	}

	allStats := m.GetAllStats()
	if len(allStats) != 2 {
		t.Errorf("Expected 2 types, got %d", len(allStats))
	}
}

func TestShrinkMetricsTiming(t *testing.T) {
	m := NewShrinkMetrics()
	if m.started.IsZero() {
		t.Error("started time should be set")
	}
	
	m.SetCurrentFile("processing.mp4")
	if m.currentFile != "processing.mp4" {
		t.Errorf("Expected currentFile processing.mp4, got %s", m.currentFile)
	}
}
