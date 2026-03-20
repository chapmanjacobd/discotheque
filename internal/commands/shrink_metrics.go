package commands

import (
	"fmt"
	"log/slog"
	"maps"
	"strings"
	"sync"
	"time"

	"github.com/chapmanjacobd/discoteca/internal/utils"
)

// MediaTypeStats tracks processing statistics for a specific media type
type MediaTypeStats struct {
	Total          int
	Processed      int
	Success        int
	Failed         int
	Skipped        int
	CompressedSize int64
	TotalSize      int64
	FutureSize     int64
	TotalTime      int   // processing time in seconds
	TotalDuration  int64 // total media duration in seconds (for speed ratio)
	CompletedAt    time.Time
}

// SuccessRate returns the success rate as a percentage
func (s *MediaTypeStats) SuccessRate() float64 {
	if s.Processed == 0 {
		return 0
	}
	return float64(s.Success) / float64(s.Processed) * 100
}

// SpaceSaved returns bytes saved
func (s *MediaTypeStats) SpaceSaved() int64 {
	return s.TotalSize - s.FutureSize
}

// AvgProcessingTime returns average processing time per file
func (s *MediaTypeStats) AvgProcessingTime() int {
	if s.Processed == 0 {
		return 0
	}
	return s.TotalTime / s.Processed
}

// SpeedRatio returns the processing speed ratio (e.g., 2.5x realtime)
func (s *MediaTypeStats) SpeedRatio() float64 {
	if s.TotalTime == 0 || s.TotalDuration == 0 {
		return 0
	}
	return float64(s.TotalDuration) / float64(s.TotalTime)
}

// ShrinkMetrics aggregates statistics across all media types
type ShrinkMetrics struct {
	mu            sync.RWMutex
	started       time.Time
	completed     time.Time
	types         map[string]*MediaTypeStats
	currentFile   string
	lastPrintTime time.Time
	linesPrinted  int // Track how many lines we printed for cursor repositioning
}

// NewShrinkMetrics creates a new metrics tracker
func NewShrinkMetrics() *ShrinkMetrics {
	return &ShrinkMetrics{
		started: time.Now(),
		types:   make(map[string]*MediaTypeStats),
	}
}

// RecordStarted records that a media item is being processed
func (m *ShrinkMetrics) RecordStarted(mediaType string, path string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	stats := m.getOrCreateType(mediaType)
	stats.Total++
	m.currentFile = path
}

// RecordSuccess records a successful processing
func (m *ShrinkMetrics) RecordSuccess(mediaType string, size, futureSize int64, processingTime int, duration int64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	stats := m.getOrCreateType(mediaType)
	stats.Processed++
	stats.Success++
	stats.TotalSize += size
	stats.FutureSize += futureSize
	stats.TotalTime += processingTime
	stats.TotalDuration += duration
	stats.CompletedAt = time.Now()
}

// RecordFailure records a failed processing
func (m *ShrinkMetrics) RecordFailure(mediaType string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	stats := m.getOrCreateType(mediaType)
	stats.Processed++
	stats.Failed++
	stats.CompletedAt = time.Now()
}

// RecordSkipped records a skipped media item
func (m *ShrinkMetrics) RecordSkipped(mediaType string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	stats := m.getOrCreateType(mediaType)
	stats.Skipped++
}

// getOrCreateType gets or creates stats for a media type
func (m *ShrinkMetrics) getOrCreateType(mediaType string) *MediaTypeStats {
	if stats, ok := m.types[mediaType]; ok {
		return stats
	}
	stats := &MediaTypeStats{}
	m.types[mediaType] = stats
	return stats
}

// PrintProgress prints the current progress with summary table
// Errors are printed normally via slog and will temporarily overwrite progress
// Progress is reprinted on next update cycle
func (m *ShrinkMetrics) PrintProgress() {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Rate limit to avoid excessive updates (max 2 per second)
	now := time.Now()
	if now.Sub(m.lastPrintTime) < 500*time.Millisecond {
		return
	}
	m.lastPrintTime = now

	// Calculate totals
	var totalSuccess, totalFailed, totalSkipped, totalQueued int
	var totalSavings int64
	var totalTime, totalDuration int

	for _, stats := range m.types {
		totalSuccess += stats.Success
		totalFailed += stats.Failed
		totalSkipped += stats.Skipped
		totalSavings += stats.SpaceSaved()
		totalTime += stats.TotalTime
		totalDuration += int(stats.TotalDuration)
		totalQueued += stats.Total - stats.Processed - stats.Skipped
	}

	// Build the progress output
	var sb strings.Builder

	// Current file path (middle-truncated to full terminal width)
	displayPath := m.currentFile
	displayPath = utils.TruncateMiddle(displayPath, utils.GetTerminalWidth())
	sb.WriteString(displayPath)
	sb.WriteString("\n\n")

	// Print summary table header
	sb.WriteString(fmt.Sprintf("%-8s %6s %6s %6s %10s %10s %8s\n",
		"Queue", "OK", "Fail", "Skip", "Saved", "Time", "Speed"))
	sb.WriteString(strings.Repeat("-", 70))
	sb.WriteString("\n")

	for mediaType, stats := range m.types {
		speed := ""
		if stats.SpeedRatio() > 0 {
			speed = fmt.Sprintf("%.1fx", stats.SpeedRatio())
		}
		timeStr := utils.FormatDuration(stats.TotalTime)
		queued := stats.Total - stats.Processed - stats.Skipped
		sb.WriteString(fmt.Sprintf("%-8s %6d %6d %6d %10s %10s %8s\n",
			mediaType,
			queued,
			stats.Success,
			stats.Failed,
			utils.FormatSize(stats.SpaceSaved()),
			timeStr,
			speed))
	}

	sb.WriteString(strings.Repeat("-", 70))
	sb.WriteString("\n")

	// Print totals
	overallSpeed := ""
	if totalTime > 0 && totalDuration > 0 {
		overallSpeed = fmt.Sprintf("%.1fx", float64(totalDuration)/float64(totalTime))
	}
	sb.WriteString(fmt.Sprintf("%-8s %6d %6d %6d %10s %10s %8s\n",
		"TOTAL",
		totalQueued,
		totalSuccess,
		totalFailed,
		utils.FormatSize(totalSavings),
		utils.FormatDuration(totalTime),
		overallSpeed))

	output := sb.String()
	lineCount := strings.Count(output, "\n") + 1

	// Move cursor to home, print progress, then clear leftover lines
	// Errors from slog will overwrite this, but we reprint on next cycle
	fmt.Print("\033[H") // Move to home
	fmt.Print(output)   // Print progress
	// Clear remaining lines from old progress (in case new progress is shorter)
	for i := lineCount; i < m.linesPrinted; i++ {
		fmt.Print("\033[K\n") // Clear line and move down
	}
	// Move cursor back to home for next error to appear below progress
	fmt.Print("\033[H")

	// Track lines printed
	m.linesPrinted = lineCount
}

// LogSummary logs the final metrics summary
func (m *ShrinkMetrics) LogSummary() {
	m.mu.RLock()
	defer m.mu.RUnlock()

	m.completed = time.Now()
	duration := m.completed.Sub(m.started)

	// Calculate totals
	var totalProcessed, totalSuccess, totalFailed int
	var totalSize, totalFuture, totalSavings int64
	var totalTime, totalDuration int

	for _, stats := range m.types {
		totalProcessed += stats.Processed
		totalSuccess += stats.Success
		totalFailed += stats.Failed
		totalSize += stats.TotalSize
		totalFuture += stats.FutureSize
		totalSavings += stats.SpaceSaved()
		totalTime += stats.TotalTime
		totalDuration += int(stats.TotalDuration)
	}

	// Log summary
	slog.Info("Processing complete",
		"duration", duration.String(),
		"processed", totalProcessed,
		"success", totalSuccess,
		"failed", totalFailed,
		"savings", utils.FormatSize(totalSavings))

	// Log per-type breakdown
	for mediaType, stats := range m.types {
		speed := ""
		if stats.SpeedRatio() > 0 {
			speed = fmt.Sprintf("%.1fx", stats.SpeedRatio())
		}
		slog.Info("Media type summary",
			"type", mediaType,
			"processed", stats.Processed,
			"success", stats.Success,
			"failed", stats.Failed,
			"savings", utils.FormatSize(stats.SpaceSaved()),
			"time", utils.FormatDuration(stats.TotalTime),
			"speed", speed)
	}
}

// GetStats returns stats for a specific media type
func (m *ShrinkMetrics) GetStats(mediaType string) *MediaTypeStats {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.types[mediaType]
}

// GetAllStats returns all stats (read-only copy)
func (m *ShrinkMetrics) GetAllStats() map[string]*MediaTypeStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	copy := make(map[string]*MediaTypeStats, len(m.types))
	maps.Copy(copy, m.types)
	return copy
}

// SetCurrentFile sets the currently processing file
func (m *ShrinkMetrics) SetCurrentFile(path string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.currentFile = path
}
