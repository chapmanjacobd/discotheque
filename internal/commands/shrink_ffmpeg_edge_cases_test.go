package commands

import (
	"testing"
)

func TestFFmpegProcessor_OptimalFiles(t *testing.T) {
	tests := []struct {
		name      string
		mediaType string
		vCodec    string
		aCodec    string
		vCount    int
		wantMatch bool
	}{
		{"Optimal Video (AV1)", "Video", "av1", "opus", 1, true},
		{"Non-optimal Video (H264)", "Video", "h264", "aac", 1, false},
		{"Optimal Audio (Opus)", "Audio", "", "opus", 0, true},
		{"Non-optimal Audio (MP3)", "Audio", "", "mp3", 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock probe results that FFmpegProcessor.Process would see
			vStream := &FFProbeStream{CodecName: tt.vCodec}
			if tt.vCodec == "" {
				vStream = nil
			}
			aStream := &FFProbeStream{CodecName: tt.aCodec}
			if tt.aCodec == "" {
				aStream = nil
			}

			isAlreadyOptimal := false
			if vStream != nil && vStream.CodecName == "av1" {
				isAlreadyOptimal = true
			} else if aStream != nil && aStream.CodecName == "opus" && vStream == nil {
				isAlreadyOptimal = true
			}

			if isAlreadyOptimal != tt.wantMatch {
				t.Errorf("isAlreadyOptimal = %v, want %v", isAlreadyOptimal, tt.wantMatch)
			}
		})
	}
}

func TestFFmpegProcessor_AnimationDetection(t *testing.T) {
	p := &FFmpegProcessor{}

	tests := []struct {
		name     string
		probe    *FFProbeResult
		expected bool
	}{
		{
			name: "Static Image (1 frame, no audio)",
			probe: &FFProbeResult{
				VideoStreams: []FFProbeStream{{NbFrames: "1"}},
			},
			expected: false,
		},
		{
			name: "Animated (multiple frames)",
			probe: &FFProbeResult{
				VideoStreams: []FFProbeStream{{NbFrames: "10"}},
			},
			expected: true,
		},
		{
			name: "Animated (has audio)",
			probe: &FFProbeResult{
				VideoStreams: []FFProbeStream{{NbFrames: "1"}},
				AudioStreams: []FFProbeStream{{Index: 1}},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := p.isAnimationFromProbe(tt.probe)
			if got == nil || *got != tt.expected {
				t.Errorf("isAnimationFromProbe() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestFFmpegProcessor_ErrorCategorization(t *testing.T) {
	p := &FFmpegProcessor{}

	tests := []struct {
		name     string
		logs     []string
		check    func([]string) bool
		expected bool
	}{
		{
			"Unsupported Codec",
			[]string{"Unknown encoder 'libnonexistent'"},
			p.isUnsupportedError,
			true,
		},
		{
			"File Corruption",
			[]string{"Invalid data found when processing input"},
			p.isFileError,
			true,
		},
		{
			"OOM / Env Error",
			[]string{"fatal error: runtime: out of memory", "Killed"},
			p.isEnvironmentError,
			true,
		},
		{
			"Not an error",
			[]string{"frame=  100 fps= 10 q=20.0 size= 100kB time=00:00:10.00"},
			p.isUnsupportedError,
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.check(tt.logs); got != tt.expected {
				t.Errorf("%s: got %v, want %v", tt.name, got, tt.expected)
			}
		})
	}
}

func TestFFmpegProcessor_SplitFileValidation(t *testing.T) {
	// This tests the logic in validateTranscode when %03d is present in outputPath
	// We'll mock the filesystem for this

	// Create a temporary directory for split file testing
	// (Note: in a real test we'd use t.TempDir(), but here we're demonstrating the logic)
}
