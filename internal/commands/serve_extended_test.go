package commands

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/chapmanjacobd/discotheque/internal/testutils"
)

func TestServeCmd_ExtendedHandlers(t *testing.T) {
	fixture := testutils.Setup(t)
	defer fixture.Cleanup()

	db := fixture.GetDB()
	InitDB(db)

	// Insert test data
	_, err := db.Exec(`
		INSERT INTO media (path, type, genre, time_deleted)
		VALUES 
		('video1.mp4', 'video/mp4', 'Action', 0),
		('video2.mp4', 'video/mp4', 'Comedy', 0),
		('audio1.mp3', 'audio/mpeg', 'Rock', 0),
		('audio2.mp3', 'audio/mpeg', 'Jazz', 0),
		('deleted.mp4', 'video/mp4', 'Horror', 123456789)
	`)
	if err != nil {
		t.Fatal(err)
	}
	db.Close()

	cmd := &ServeCmd{
		Databases: []string{fixture.DBPath},
		Trashcan:  true,
	}
	cmd.HideDeleted = true
	handler := cmd.Mux()

	t.Run("HandleDatabases", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/databases", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected 200, got %d", w.Code)
		}

		var resp struct {
			Databases []string `json:"databases"`
			Trashcan  bool     `json:"trashcan"`
		}
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatal(err)
		}
		if len(resp.Databases) != 1 || resp.Databases[0] != fixture.DBPath {
			t.Errorf("Unexpected databases in response: %v", resp.Databases)
		}
		if !resp.Trashcan {
			t.Error("Expected Trashcan to be true")
		}
	})

	t.Run("HandleGenres", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/genres", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected 200, got %d", w.Code)
		}

		var resp []struct {
			Genre string `json:"genre"`
			Count int    `json:"count"`
		}
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatal(err)
		}

		if len(resp) != 4 {
			t.Errorf("Expected 4 genres, got %d", len(resp))
		}
	})

	t.Run("HandleQuery_Offset", func(t *testing.T) {
		// Get all 4 active items, sorted by path
		// video1.mp4, video2.mp4, audio1.mp3, audio2.mp3
		// Natural sort: audio1.mp3, audio2.mp3, video1.mp4, video2.mp4

		req := httptest.NewRequest(http.MethodGet, "/api/query?limit=2&offset=2&sort=path", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected 200, got %d", w.Code)
		}

		var resp []struct {
			Path string `json:"path"`
		}
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatal(err)
		}

		if len(resp) != 2 {
			t.Errorf("Expected 2 results, got %d", len(resp))
		}
		// Based on path sort: audio1.mp3 (0), audio2.mp3 (1), video1.mp4 (2), video2.mp4 (3)
		if resp[0].Path != "video1.mp4" {
			t.Errorf("Expected video1.mp4 at offset 2, got %s", resp[0].Path)
		}
	})

	t.Run("HandlePlaylists", func(t *testing.T) {
		// POST create playlist
		payload := `{"title": "My Playlist"}`
		req := httptest.NewRequest(http.MethodPost, "/api/playlists", strings.NewReader(payload))
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected 200, got %d", w.Code)
		}

		var createResp struct {
			ID int64  `json:"id"`
			DB string `json:"db"`
		}
		if err := json.NewDecoder(w.Body).Decode(&createResp); err != nil {
			t.Fatal(err)
		}

		// GET playlists
		req = httptest.NewRequest(http.MethodGet, "/api/playlists", nil)
		w = httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		var listResp []struct {
			ID    int64  `json:"id"`
			Title string `json:"title"`
		}
		if err := json.NewDecoder(w.Body).Decode(&listResp); err != nil {
			t.Fatal(err)
		}
		if len(listResp) != 1 || listResp[0].Title != "My Playlist" {
			t.Errorf("Unexpected playlist list: %v", listResp)
		}

		// POST add item to playlist
		addItemPayload := fmt.Sprintf(`{"playlist_id": %d, "db": "%s", "media_path": "video1.mp4"}`, createResp.ID, createResp.DB)
		req = httptest.NewRequest(http.MethodPost, "/api/playlists/items", strings.NewReader(addItemPayload))
		w = httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("Expected 200 adding item, got %d", w.Code)
		}

		// GET playlist items
		req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/playlists/items?id=%d&db=%s", createResp.ID, createResp.DB), nil)
		w = httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		var itemsResp []any
		if err := json.NewDecoder(w.Body).Decode(&itemsResp); err != nil {
			t.Fatal(err)
		}
		if len(itemsResp) != 1 {
			t.Errorf("Expected 1 item in playlist, got %d", len(itemsResp))
		}

		// DELETE playlist item
		deleteItemPayload := fmt.Sprintf(`{"playlist_id": %d, "db": "%s", "media_path": "video1.mp4"}`, createResp.ID, createResp.DB)
		req = httptest.NewRequest(http.MethodDelete, "/api/playlists/items", strings.NewReader(deleteItemPayload))
		w = httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("Expected 200 deleting item, got %d", w.Code)
		}

		// DELETE playlist
		req = httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/playlists?id=%d&db=%s", createResp.ID, createResp.DB), nil)
		w = httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("Expected 200 deleting playlist, got %d", w.Code)
		}
	})

	t.Run("HandleTrash", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/trash", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected 200, got %d", w.Code)
		}

		var resp []any
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatal(err)
		}
		if len(resp) != 1 {
			t.Errorf("Expected 1 deleted item, got %d", len(resp))
		}
	})

	t.Run("HandleHLSPlaylist", func(t *testing.T) {
		db := fixture.GetDB()
		db.Exec("INSERT INTO media (path, type, duration, time_deleted) VALUES (?, 'video/mp4', 120, 0)", "hls_video.mp4")
		db.Close()

		req := httptest.NewRequest(http.MethodGet, "/api/hls/playlist?path=hls_video.mp4", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected 200, got %d", w.Code)
		}
		if !strings.Contains(w.Body.String(), "#EXTM3U") {
			t.Error("Response body did not contain #EXTM3U")
		}
	})

	t.Run("HandleEmptyBin", func(t *testing.T) {
		// Create a file, mark it deleted, then empty bin
		dummyPath := fixture.CreateDummyFile("to_be_permanently_deleted.mp4")
		db := fixture.GetDB()
		db.Exec("INSERT INTO media (path, type, time_deleted) VALUES (?, 'video/mp4', 12345)", dummyPath)
		db.Close()

		req := httptest.NewRequest(http.MethodPost, "/api/empty-bin", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected 200, got %d", w.Code)
		}

		// Verify file is gone from disk
		if _, err := os.Stat(dummyPath); !os.IsNotExist(err) {
			t.Error("Expected file to be deleted from disk")
		}

		// Verify row is gone from DB
		db = fixture.GetDB()
		defer db.Close()
		var count int
		db.QueryRow("SELECT count(*) FROM media WHERE path = ?", dummyPath).Scan(&count)
		if count != 0 {
			t.Error("Expected row to be deleted from DB")
		}
	})
}
