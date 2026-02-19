package models

import (
	"database/sql"
	"testing"

	"github.com/chapmanjacobd/discotheque/internal/db"
)

func TestFromDB(t *testing.T) {
	dbMedia := db.Media{
		Path:     "/test/movie.mp4",
		Title:    sql.NullString{String: "Test Movie", Valid: true},
		Size:     sql.NullInt64{Int64: 1024, Valid: true},
		Duration: sql.NullInt64{Valid: false},
	}

	media := FromDB(dbMedia)

	if media.Path != "/test/movie.mp4" {
		t.Errorf("Expected path /test/movie.mp4, got %s", media.Path)
	}

	if media.Title == nil || *media.Title != "Test Movie" {
		t.Errorf("Expected title Test Movie, got %v", media.Title)
	}

	if media.Size == nil || *media.Size != 1024 {
		t.Errorf("Expected size 1024, got %v", media.Size)
	}

	if media.Duration != nil {
		t.Errorf("Expected duration nil, got %v", media.Duration)
	}

	// Test nullFloat64Ptr
	dbMedia.Fps = sql.NullFloat64{Float64: 24.0, Valid: true}
	media = FromDB(dbMedia)
	if media.Fps == nil || *media.Fps != 24.0 {
		t.Errorf("Expected Fps 24.0, got %v", media.Fps)
	}
}

func TestMedia_Parent(t *testing.T) {
	m := Media{Path: "/dir/sub/file.mp4"}
	if got := m.Parent(); got != "/dir/sub" {
		t.Errorf("Parent() = %v, want /dir/sub", got)
	}
}
