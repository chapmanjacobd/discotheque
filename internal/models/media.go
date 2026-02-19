package models

import (
	"database/sql"
	"path/filepath"

	"github.com/chapmanjacobd/discotheque/internal/db"
)

type Media struct {
	Path            string   `json:"path"`
	Title           *string  `json:"title,omitempty"`
	Duration        *int64   `json:"duration,omitempty"`
	Size            *int64   `json:"size,omitempty"`
	TimeCreated     *int64   `json:"time_created,omitempty"`
	TimeModified    *int64   `json:"time_modified,omitempty"`
	TimeDeleted     *int64   `json:"time_deleted,omitempty"`
	TimeFirstPlayed *int64   `json:"time_first_played,omitempty"`
	TimeLastPlayed  *int64   `json:"time_last_played,omitempty"`
	PlayCount       *int64   `json:"play_count,omitempty"`
	Playhead        *int64   `json:"playhead,omitempty"`
	Type            *string  `json:"type,omitempty"`
	Width           *int64   `json:"width,omitempty"`
	Height          *int64   `json:"height,omitempty"`
	Fps             *float64 `json:"fps,omitempty"`
	VideoCodecs     *string  `json:"video_codecs,omitempty"`
	AudioCodecs     *string  `json:"audio_codecs,omitempty"`
	SubtitleCodecs  *string  `json:"subtitle_codecs,omitempty"`
	VideoCount      *int64   `json:"video_count,omitempty"`
	AudioCount      *int64   `json:"audio_count,omitempty"`
	SubtitleCount   *int64   `json:"subtitle_count,omitempty"`
	Album           *string  `json:"album,omitempty"`
	Artist          *string  `json:"artist,omitempty"`
	Genre           *string  `json:"genre,omitempty"`
	Description     *string  `json:"description,omitempty"`
	Language        *string  `json:"language,omitempty"`
}

func (m *Media) Parent() string {
	return filepath.Dir(m.Path)
}

// Helper functions for mapping from sql.Null types

func FromDB(m db.Media) Media {
	return Media{
		Path:            m.Path,
		Title:           nullStringPtr(m.Title),
		Duration:        nullInt64Ptr(m.Duration),
		Size:            nullInt64Ptr(m.Size),
		TimeCreated:     nullInt64Ptr(m.TimeCreated),
		TimeModified:    nullInt64Ptr(m.TimeModified),
		TimeDeleted:     nullInt64Ptr(m.TimeDeleted),
		TimeFirstPlayed: nullInt64Ptr(m.TimeFirstPlayed),
		TimeLastPlayed:  nullInt64Ptr(m.TimeLastPlayed),
		PlayCount:       nullInt64Ptr(m.PlayCount),
		Playhead:        nullInt64Ptr(m.Playhead),
		Type:            nullStringPtr(m.Type),
		Width:           nullInt64Ptr(m.Width),
		Height:          nullInt64Ptr(m.Height),
		Fps:             nullFloat64Ptr(m.Fps),
		VideoCodecs:     nullStringPtr(m.VideoCodecs),
		AudioCodecs:     nullStringPtr(m.AudioCodecs),
		SubtitleCodecs:  nullStringPtr(m.SubtitleCodecs),
		VideoCount:      nullInt64Ptr(m.VideoCount),
		AudioCount:      nullInt64Ptr(m.AudioCount),
		SubtitleCount:   nullInt64Ptr(m.SubtitleCount),
		Album:           nullStringPtr(m.Album),
		Artist:          nullStringPtr(m.Artist),
		Genre:           nullStringPtr(m.Genre),
		Description:     nullStringPtr(m.Description),
		Language:        nullStringPtr(m.Language),
	}
}

func nullStringPtr(ns sql.NullString) *string {
	if !ns.Valid {
		return nil
	}
	return &ns.String
}

func nullInt64Ptr(ni sql.NullInt64) *int64 {
	if !ni.Valid {
		return nil
	}
	return &ni.Int64
}

func nullFloat64Ptr(nf sql.NullFloat64) *float64 {
	if !nf.Valid {
		return nil
	}
	return &nf.Float64
}
