package commands

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/chapmanjacobd/discotheque/internal/testutils"
)

func TestHandleRaw_FileNotFound(t *testing.T) {
	fixture := testutils.Setup(t)
	defer fixture.Cleanup()

	db := fixture.GetDB()
	InitDB(db)

	// Insert a path that does NOT exist on disk
	missingPath := "/tmp/this/file/does/not/exist/at/all/ever.mp4"
	_, err := db.Exec("INSERT INTO media (path, type, time_deleted) VALUES (?, 'video/mp4', 0)", missingPath)
	if err != nil {
		t.Fatal(err)
	}
	db.Close()

	cmd := &ServeCmd{
		Databases: []string{fixture.DBPath},
	}
	handler := cmd.Mux()

	// Request the missing file
	req := httptest.NewRequest(http.MethodGet, "/api/raw?path="+missingPath, nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	// 1. Should return 404
	if w.Code != http.StatusNotFound {
		t.Errorf("Expected 404, got %d", w.Code)
	}

	// 2. Should have marked as deleted in DB
	db = fixture.GetDB()
	defer db.Close()
	var timeDeleted int64
	err = db.QueryRow("SELECT time_deleted FROM media WHERE path = ?", missingPath).Scan(&timeDeleted)
	if err != nil {
		t.Fatal(err)
	}
	if timeDeleted == 0 {
		t.Error("Expected time_deleted to be non-zero after 404")
	}
}
