package commands

import (
	"database/sql"
	"testing"

	"github.com/chapmanjacobd/discotheque/internal/testutils"
)

func TestSearchCaptionsCmd_Run(t *testing.T) {
	fixture := testutils.Setup(t)
	defer fixture.Cleanup()

	dbPath := fixture.DBPath
	sqlDB, _ := sql.Open("sqlite3", dbPath)
	InitDB(sqlDB)

	// Skip if FTS5 is not available
	var name string
	err := sqlDB.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='captions_fts'").Scan(&name)
	if err != nil {
		t.Skip("FTS5 not available, skipping search captions tests")
	}

	sqlDB.Exec("INSERT INTO media (path, title) VALUES (?, ?)", "/path/video1.mp4", "Video 1")
	sqlDB.Exec("INSERT INTO captions (media_path, time, text) VALUES (?, ?, ?)", "/path/video1.mp4", 10.0, "hello world")
	sqlDB.Exec("INSERT INTO captions (media_path, time, text) VALUES (?, ?, ?)", "/path/video1.mp4", 12.0, "this is overlapping")
	sqlDB.Exec("INSERT INTO captions (media_path, time, text) VALUES (?, ?, ?)", "/path/video1.mp4", 40.0, "different world")
	sqlDB.Close()

	t.Run("BasicSearch", func(t *testing.T) {
		cmd := &SearchCaptionsCmd{
			Database: dbPath,
			Search:   []string{"world"},
		}

		if err := cmd.Run(nil); err != nil {
			t.Fatalf("SearchCaptionsCmd failed: %v", err)
		}
	})

	t.Run("OverlapMerging", func(t *testing.T) {
		cmd := &SearchCaptionsCmd{
			Database: dbPath,
			Search:   []string{"world", "overlapping"},
			Overlap:  10,
		}

		if err := cmd.Run(nil); err != nil {
			t.Fatalf("SearchCaptionsCmd failed: %v", err)
		}
	})
}
