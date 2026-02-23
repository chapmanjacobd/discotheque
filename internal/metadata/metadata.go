package metadata

import (
	"context"
	"database/sql"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/chapmanjacobd/discotheque/internal/db"
	"github.com/chapmanjacobd/discotheque/internal/utils"
)

type FFProbeOutput struct {
	Streams  []Stream  `json:"streams"`
	Format   Format    `json:"format"`
	Chapters []Chapter `json:"chapters"`
}

type Chapter struct {
	ID        int               `json:"id"`
	StartTime string            `json:"start_time"`
	EndTime   string            `json:"end_time"`
	Tags      map[string]string `json:"tags"`
}

type MediaMetadata struct {
	Media    db.UpsertMediaParams
	Captions []db.InsertCaptionParams
}

type Stream struct {
	CodecType    string            `json:"codec_type"`
	CodecName    string            `json:"codec_name"`
	Width        int               `json:"width"`
	Height       int               `json:"height"`
	AvgFrameRate string            `json:"avg_frame_rate"`
	RFrameRate   string            `json:"r_frame_rate"`
	Duration     string            `json:"duration"`
	Tags         map[string]string `json:"tags"`
	Disposition  map[string]int    `json:"disposition"`
}

type Format struct {
	Filename string            `json:"filename"`
	Duration string            `json:"duration"`
	Size     string            `json:"size"`
	BitRate  string            `json:"bit_rate"`
	Tags     map[string]string `json:"tags"`
}

func Extract(ctx context.Context, path string, scanSubtitles bool) (*MediaMetadata, error) {
	stat, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	// Detect mimetype first
	mimeStr := utils.DetectMimeType(path)

	// Advanced Type Detection
	mediaType := ""
	if strings.HasPrefix(mimeStr, "image/") {
		mediaType = "image"
	} else if strings.HasPrefix(mimeStr, "text/") || mimeStr == "application/pdf" || mimeStr == "application/epub+zip" {
		mediaType = "text"
	} else if mimeStr != "" {
		// Fallback to coarse mimetype category
		parts := strings.Split(mimeStr, "/")
		mediaType = parts[0]
	}

	cmd := exec.CommandContext(ctx, "ffprobe",
		"-hide_banner",
		"-show_format",
		"-show_streams",
		"-show_chapters",
		"-of", "json",
		path,
	)

	output, err := cmd.Output()
	if err != nil {
		// If ffprobe fails, we still return basic file info with detected mimetype
		return basicInfo(path, stat, mediaType), nil
	}

	var data FFProbeOutput
	if err := json.Unmarshal(output, &data); err != nil {
		return basicInfo(path, stat, mediaType), nil
	}

	params := db.UpsertMediaParams{
		Path:         path,
		Size:         utils.ToNullInt64(stat.Size()),
		TimeCreated:  utils.ToNullInt64(stat.ModTime().Unix()),
		TimeModified: utils.ToNullInt64(stat.ModTime().Unix()),
	}

	// Format info
	var duration int64
	if d, err := strconv.ParseFloat(data.Format.Duration, 64); err == nil {
		duration = int64(d)
		params.Duration = utils.ToNullInt64(duration)
	}

	if data.Format.Tags != nil {
		params.Title = utils.ToNullString(data.Format.Tags["title"])
		params.Artist = utils.ToNullString(data.Format.Tags["artist"])
		params.Album = utils.ToNullString(data.Format.Tags["album"])
		params.Genre = utils.ToNullString(data.Format.Tags["genre"])
		params.Mood = utils.ToNullString(data.Format.Tags["mood"])
		if bpm, err := strconv.ParseInt(data.Format.Tags["bpm"], 10, 64); err == nil {
			params.Bpm = utils.ToNullInt64(bpm)
		}
		params.Key = utils.ToNullString(data.Format.Tags["key"])
		params.Decade = utils.ToNullString(data.Format.Tags["decade"])
		params.Categories = utils.ToNullString(data.Format.Tags["categories"])
		params.City = utils.ToNullString(data.Format.Tags["city"])
		params.Country = utils.ToNullString(data.Format.Tags["country"])
		params.Description = utils.ToNullString(data.Format.Tags["comment"])
		params.Language = utils.ToNullString(data.Format.Tags["language"])

		if ts := utils.SpecificDate(
			data.Format.Tags["originalyear"],
			data.Format.Tags["TDOR"],
			data.Format.Tags["TORY"],
			data.Format.Tags["date"],
			data.Format.Tags["TDRC"],
			data.Format.Tags["TDRL"],
			data.Format.Tags["year"],
		); ts != nil {
			params.TimeCreated = utils.ToNullInt64(*ts)
		}

		params.Uploader = utils.ToNullString(data.Format.Tags["uploader"])
		params.Webpath = utils.ToNullString(data.Format.Tags["purl"]) // FFmpeg sometimes puts URL in purl tag
		if params.Webpath.String == "" {
			params.Webpath = utils.ToNullString(data.Format.Tags["comment"]) // common for youtube-dl
		}

		if v, err := strconv.ParseInt(data.Format.Tags["view_count"], 10, 64); err == nil {
			params.ViewCount = utils.ToNullInt64(v)
		}
		if v, err := strconv.ParseInt(data.Format.Tags["comment_count"], 10, 64); err == nil {
			params.NumComments = utils.ToNullInt64(v)
		}
		if v, err := strconv.ParseInt(data.Format.Tags["like_count"], 10, 64); err == nil {
			params.FavoriteCount = utils.ToNullInt64(v)
		}
	}

	// Streams info
	var vCodecs, aCodecs, sCodecs []string
	var vCount, aCount, sCount int64

	for _, s := range data.Streams {
		switch s.CodecType {
		case "video":
			// Check if it's album art or image-based "video" codec
			if s.Disposition["attached_pic"] == 1 || s.CodecName == "mjpeg" || s.CodecName == "png" {
				continue
			}
			vCount++
			vCodecs = append(vCodecs, s.CodecName)
			if params.Width.Int64 == 0 {
				params.Width = utils.ToNullInt64(int64(s.Width))
				params.Height = utils.ToNullInt64(int64(s.Height))
				params.Fps = utils.ToNullFloat64(parseFPS(s.AvgFrameRate))
			}
		case "audio":
			aCount++
			aCodecs = append(aCodecs, s.CodecName)
		case "subtitle":
			sCount++
			sCodecs = append(sCodecs, s.CodecName)
		}
	}

	params.VideoCodecs = utils.ToNullString(utils.Combine(vCodecs))
	params.AudioCodecs = utils.ToNullString(utils.Combine(aCodecs))

	// External Subtitles
	if scanSubtitles {
		externalSubs := utils.GetExternalSubtitles(path)
		for _, sub := range externalSubs {
			sCount++
			sCodecs = append(sCodecs, strings.TrimPrefix(filepath.Ext(sub), "."))
		}
	}

	params.SubtitleCodecs = utils.ToNullString(utils.Combine(sCodecs))
	params.VideoCount = utils.ToNullInt64(vCount)
	params.AudioCount = utils.ToNullInt64(aCount)
	params.SubtitleCount = utils.ToNullInt64(sCount)

	// Refine Type Detection with stream info
	if vCount > 0 {
		mediaType = "video"
		// If it's a single video track with no duration and no audio, it's actually an image
		if vCount == 1 && aCount == 0 && duration == 0 {
			mediaType = "image"
		}
	} else if aCount > 0 {
		mediaType = "audio"
		// Distinguish audiobooks: duration > 1 hour OR "audiobook" in path
		lowerPath := strings.ToLower(path)
		if duration > 3600 || strings.Contains(lowerPath, "audiobook") {
			mediaType = "audiobook"
		}
	}

	params.Type = utils.ToNullString(mediaType)

	if mediaType == "text" && params.Duration.Int64 == 0 {
		// Python: cast(length(tags) / 4.2 / 220 * 60 as INT) + 10 duration
		allTags := ""
		if data.Format.Tags != nil {
			for _, v := range data.Format.Tags {
				allTags += v
			}
		}
		if allTags != "" {
			d := int64(float64(len(allTags))/4.2/220*60) + 10
			params.Duration = utils.ToNullInt64(d)
		}
	}

	metadata := &MediaMetadata{
		Media: params,
	}

	// Extract chapters
	for _, ch := range data.Chapters {
		title := ch.Tags["title"]
		if title == "" {
			continue
		}
		startTime, _ := strconv.ParseFloat(ch.StartTime, 64)
		metadata.Captions = append(metadata.Captions, db.InsertCaptionParams{
			MediaPath: path,
			Time:      sql.NullFloat64{Float64: startTime, Valid: true},
			Text:      sql.NullString{String: title, Valid: true},
		})
	}

	return metadata, nil
}

func basicInfo(path string, stat os.FileInfo, mimeStr string) *MediaMetadata {
	return &MediaMetadata{
		Media: db.UpsertMediaParams{
			Path:         path,
			Size:         utils.ToNullInt64(stat.Size()),
			TimeCreated:  utils.ToNullInt64(stat.ModTime().Unix()),
			TimeModified: utils.ToNullInt64(stat.ModTime().Unix()),
			Type:         utils.ToNullString(mimeStr),
		},
	}
}

func parseFPS(s string) float64 {
	parts := strings.Split(s, "/")
	if len(parts) != 2 {
		return 0
	}
	num, _ := strconv.ParseFloat(parts[0], 64)
	den, _ := strconv.ParseFloat(parts[1], 64)
	if den == 0 {
		return 0
	}
	return num / den
}
