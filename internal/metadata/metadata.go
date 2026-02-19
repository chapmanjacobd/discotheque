package metadata

import (
	"context"
	"database/sql"
	"encoding/json"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/chapmanjacobd/discotheque/internal/db"
)

type FFProbeOutput struct {
	Streams []Stream `json:"streams"`
	Format  Format   `json:"format"`
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

func Extract(ctx context.Context, path string) (*db.UpsertMediaParams, error) {
	stat, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	cmd := exec.CommandContext(ctx, "ffprobe",
		"-hide_banner",
		"-show_format",
		"-show_streams",
		"-of", "json",
		path,
	)

	output, err := cmd.Output()
	if err != nil {
		// If ffprobe fails, we still return basic file info
		return basicInfo(path, stat), nil
	}

	var data FFProbeOutput
	if err := json.Unmarshal(output, &data); err != nil {
		return basicInfo(path, stat), nil
	}

	params := &db.UpsertMediaParams{
		Path:         path,
		Size:         toNullInt64(stat.Size()),
		TimeCreated:  toNullInt64(stat.ModTime().Unix()), // stat.ModTime is safer across OSes than ctime
		TimeModified: toNullInt64(stat.ModTime().Unix()),
	}

	// Format info
	if d, err := strconv.ParseFloat(data.Format.Duration, 64); err == nil {
		params.Duration = toNullInt64(int64(d))
	}

	if data.Format.Tags != nil {
		params.Title = toNullString(data.Format.Tags["title"])
		params.Artist = toNullString(data.Format.Tags["artist"])
		params.Album = toNullString(data.Format.Tags["album"])
		params.Genre = toNullString(data.Format.Tags["genre"])
		params.Description = toNullString(data.Format.Tags["comment"])
		params.Language = toNullString(data.Format.Tags["language"])
	}

	// Streams info
	var vCodecs, aCodecs, sCodecs []string
	var vCount, aCount, sCount int64

	for _, s := range data.Streams {
		switch s.CodecType {
		case "video":
			// Check if it's album art
			if s.Disposition["attached_pic"] == 1 {
				continue
			}
			vCount++
			vCodecs = append(vCodecs, s.CodecName)
			if params.Width.Int64 == 0 {
				params.Width = toNullInt64(int64(s.Width))
				params.Height = toNullInt64(int64(s.Height))
				params.Fps = toNullFloat64(parseFPS(s.AvgFrameRate))
			}
		case "audio":
			aCount++
			aCodecs = append(aCodecs, s.CodecName)
		case "subtitle":
			sCount++
			sCodecs = append(sCodecs, s.CodecName)
		}
	}

	params.VideoCodecs = toNullString(strings.Join(unique(vCodecs), ","))
	params.AudioCodecs = toNullString(strings.Join(unique(aCodecs), ","))
	params.SubtitleCodecs = toNullString(strings.Join(unique(sCodecs), ","))
	params.VideoCount = toNullInt64(vCount)
	params.AudioCount = toNullInt64(aCount)
	params.SubtitleCount = toNullInt64(sCount)

	// Determine type
	if vCount > 0 {
		params.Type = toNullString("video")
	} else if aCount > 0 {
		params.Type = toNullString("audio")
	}

	return params, nil
}

func basicInfo(path string, stat os.FileInfo) *db.UpsertMediaParams {
	return &db.UpsertMediaParams{
		Path:         path,
		Size:         toNullInt64(stat.Size()),
		TimeCreated:  toNullInt64(stat.ModTime().Unix()),
		TimeModified: toNullInt64(stat.ModTime().Unix()),
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

func unique(ss []string) []string {
	m := make(map[string]bool)
	var res []string
	for _, s := range ss {
		if s != "" && !m[s] {
			m[s] = true
			res = append(res, s)
		}
	}
	return res
}

// Helpers for sql.Null* types
func toNullInt64(i int64) sql.NullInt64 {
	return sql.NullInt64{Int64: i, Valid: i != 0}
}

func toNullString(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}

func toNullFloat64(f float64) sql.NullFloat64 {
	return sql.NullFloat64{Float64: f, Valid: f != 0}
}
