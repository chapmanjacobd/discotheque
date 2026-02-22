package utils

import (
	"log/slog"
	"path/filepath"
	"strings"

	"github.com/chapmanjacobd/discotheque/internal/models"
)

type TranscodeStrategy struct {
	NeedsTranscode bool
	VideoCopy      bool
	AudioCopy      bool
	TargetMime     string
}

func GetTranscodeStrategy(m models.Media) TranscodeStrategy {
	ext := strings.ToLower(filepath.Ext(m.Path))

	// If it's a known non-media format, don't even try
	if ext == ".sqlite" || ext == ".db" || ext == ".txt" {
		return TranscodeStrategy{NeedsTranscode: false}
	}

	isSupportedVideoCodec := func(codec string) bool {
		codec = strings.ToLower(codec)
		return strings.Contains(codec, "h264") || strings.Contains(codec, "avc1") || strings.Contains(codec, "vp8") || strings.Contains(codec, "vp9") || strings.Contains(codec, "av1")
	}

	isSupportedAudioCodec := func(codec string) bool {
		if codec == "" {
			return false
		}
		codec = strings.ToLower(codec)
		// If it contains any incompatible codec, return false
		incompatible := []string{"eac3", "ac3", "dts", "truehd", "mlp"}
		for _, inc := range incompatible {
			if strings.Contains(codec, inc) {
				return false
			}
		}

		// It must contain at least one supported codec
		supported := []string{"aac", "mp3", "opus", "vorbis", "flac", "pcm", "wav"}
		for _, sup := range supported {
			if strings.Contains(codec, sup) {
				return true
			}
		}
		return false
	}

	vCodecs := ""
	if m.VideoCodecs != nil {
		vCodecs = *m.VideoCodecs
	}
	aCodecs := ""
	if m.AudioCodecs != nil {
		aCodecs = *m.AudioCodecs
	}

	mime := ""
	if m.Type != nil && *m.Type != "" {
		mime = *m.Type
	} else {
		mime = DetectMimeType(m.Path)
	}

	if strings.HasPrefix(mime, "image") {
		return TranscodeStrategy{NeedsTranscode: false}
	}

	var strategy TranscodeStrategy
	if strings.HasPrefix(mime, "video") {
		vNeeds := !isSupportedVideoCodec(vCodecs)
		aNeeds := !isSupportedAudioCodec(aCodecs)

		// Prefer WebM for VP9/VP8/AV1/Opus/Vorbis
		preferWebm := strings.Contains(strings.ToLower(vCodecs), "vp9") || strings.Contains(strings.ToLower(vCodecs), "vp8") || strings.Contains(strings.ToLower(vCodecs), "av1") ||
			strings.Contains(strings.ToLower(aCodecs), "opus") || strings.Contains(strings.ToLower(aCodecs), "vorbis")

		targetMime := "video/mp4"
		if preferWebm {
			targetMime = "video/webm"
		}

		// Check if container already matches the target mime type
		containerMatches := false
		if targetMime == "video/mp4" {
			// Most browsers support H264/AAC in MKV or MOV as well, but we'll be slightly conservative
			if ext == ".mp4" || ext == ".m4v" || ext == ".mov" || ext == ".mkv" {
				containerMatches = true
			}
		} else if targetMime == "video/webm" {
			if ext == ".webm" || ext == ".mkv" {
				containerMatches = true
			}
		}

		if vNeeds || aNeeds || !containerMatches {
			strategy = TranscodeStrategy{
				NeedsTranscode: true,
				VideoCopy:      !vNeeds,
				AudioCopy:      !aNeeds,
				TargetMime:     targetMime,
			}
		} else {
			strategy = TranscodeStrategy{NeedsTranscode: false}
		}
	} else if strings.HasPrefix(mime, "audio") {
		if !isSupportedAudioCodec(aCodecs) || (ext != ".mp3" && ext != ".m4a" && ext != ".ogg" && ext != ".flac" && ext != ".wav" && ext != ".opus") {
			strategy = TranscodeStrategy{
				NeedsTranscode: true,
				AudioCopy:      isSupportedAudioCodec(aCodecs),
				TargetMime:     "audio/mpeg",
			}
		} else {
			strategy = TranscodeStrategy{NeedsTranscode: false}
		}
	}

	if strategy.NeedsTranscode {
		slog.Debug("Transcode Strategy", "path", m.Path, "video_copy", strategy.VideoCopy, "audio_copy", strategy.AudioCopy, "target", strategy.TargetMime, "vcodecs", vCodecs, "acodecs", aCodecs)
	}

	return strategy
}
