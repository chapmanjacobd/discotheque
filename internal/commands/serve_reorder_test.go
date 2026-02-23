package commands

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/chapmanjacobd/discotheque/internal/testutils"
)

func TestServeCmd_PlaylistReorder(t *testing.T) {
	fixture := testutils.Setup(t)
	defer fixture.Cleanup()

	db := fixture.GetDB()
	InitDB(db)

	// Insert test data
	_, err := db.Exec(`
		INSERT INTO media (path, type, time_deleted)
		VALUES 
		('item1.mp4', 'video/mp4', 0),
		('item2.mp4', 'video/mp4', 0),
		('item3.mp4', 'video/mp4', 0);

		INSERT INTO playlists (title, path, time_deleted)
		VALUES ('Test Playlist', 'custom:123', 0);

		INSERT INTO playlist_items (playlist_id, media_path, track_number, time_added)
		VALUES 
		(1, 'item1.mp4', NULL, 100),
		(1, 'item2.mp4', NULL, 200),
		(1, 'item3.mp4', 1, 300);
	`)
	if err != nil {
		t.Fatal(err)
	}
	db.Close()

	cmd := &ServeCmd{
		Databases: []string{fixture.DBPath},
	}
	handler := cmd.Mux()

	// Initial check: item1 (NULL, 100), item2 (NULL, 200), item3 (1, 300)
	// SQLite sorts NULLs first. So order should be item1, item2, item3 (since item1 < item2 by time_added if track_number matches - wait, query orders by track_number, then time_added)
	// Query: ORDER BY pi.track_number, pi.time_added, m.path
	// NULL, 100 -> 1st
	// NULL, 200 -> 2nd
	// 1, 300    -> 3rd

	t.Run("InitialOrder", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/playlists/items?title=Test%20Playlist", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("Expected 200, got %d", w.Code)
		}

		var items []struct {
			Path        string `json:"path"`
			TrackNumber *int64 `json:"track_number"`
		}
		if err := json.NewDecoder(w.Body).Decode(&items); err != nil {
			t.Fatal(err)
		}

		if len(items) != 3 {
			t.Fatalf("Expected 3 items, got %d", len(items))
		}
		if items[0].Path != "item1.mp4" {
			t.Errorf("Expected item1 first, got %s", items[0].Path)
		}
		if items[1].Path != "item2.mp4" {
			t.Errorf("Expected item2 second, got %s", items[1].Path)
		}
		if items[2].Path != "item3.mp4" {
			t.Errorf("Expected item3 third, got %s", items[2].Path)
		}
	})

	// Move item3 (last) to position 0 (first)
	t.Run("MoveLastToFirst", func(t *testing.T) {
		payload := map[string]any{
			"playlist_title": "Test Playlist",
			"media_path":     "item3.mp4",
			"new_index":      0,
		}
		body, _ := json.Marshal(payload)
		req := httptest.NewRequest(http.MethodPost, "/api/playlists/reorder", bytes.NewBuffer(body))
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("Expected 200, got %d", w.Code)
		}

		// Verify new order: item3, item1, item2
		req = httptest.NewRequest(http.MethodGet, "/api/playlists/items?title=Test%20Playlist", nil)
		w = httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		var items []struct {
			Path        string `json:"path"`
			TrackNumber *int64 `json:"track_number"`
		}
		json.NewDecoder(w.Body).Decode(&items)

		if items[0].Path != "item3.mp4" {
			t.Errorf("Expected item3 first, got %s", items[0].Path)
		}
		if items[1].Path != "item1.mp4" {
			t.Errorf("Expected item1 second, got %s", items[1].Path)
		}
		if items[2].Path != "item2.mp4" {
			t.Errorf("Expected item2 third, got %s", items[2].Path)
		}

		// Verify track numbers are normalized (1, 2, 3)
		if *items[0].TrackNumber != 1 {
			t.Errorf("Expected item3 track 1, got %v", items[0].TrackNumber)
		}
		if *items[1].TrackNumber != 2 {
			t.Errorf("Expected item1 track 2, got %v", items[1].TrackNumber)
		}
		if *items[2].TrackNumber != 3 {
			t.Errorf("Expected item2 track 3, got %v", items[2].TrackNumber)
		}
	})

	// Move item1 (now middle) to end (index 2)
	// Current: item3, item1, item2
	// Move item1 -> index 2
	// Expected: item3, item2, item1
	t.Run("MoveMiddleToEnd", func(t *testing.T) {
		payload := map[string]any{
			"playlist_title": "Test Playlist",
			"media_path":     "item1.mp4",
			"new_index":      2,
		}
		body, _ := json.Marshal(payload)
		req := httptest.NewRequest(http.MethodPost, "/api/playlists/reorder", bytes.NewBuffer(body))
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("Expected 200, got %d", w.Code)
		}

		req = httptest.NewRequest(http.MethodGet, "/api/playlists/items?title=Test%20Playlist", nil)
		w = httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		var items []struct {
			Path string `json:"path"`
		}
		json.NewDecoder(w.Body).Decode(&items)

		if items[0].Path != "item3.mp4" {
			t.Errorf("Expected item3 first, got %s", items[0].Path)
		}
		if items[1].Path != "item2.mp4" {
			t.Errorf("Expected item2 second, got %s", items[1].Path)
		}
		if items[2].Path != "item1.mp4" {
			t.Errorf("Expected item1 third, got %s", items[2].Path)
		}
	})

	// Move item1 (now last) to middle (index 1)
	// Current: item3, item2, item1
	// Move item1 -> index 1
	// Expected: item3, item1, item2
	t.Run("MoveEndToMiddle", func(t *testing.T) {
		payload := map[string]any{
			"playlist_title": "Test Playlist",
			"media_path":     "item1.mp4",
			"new_index":      1,
		}
		body, _ := json.Marshal(payload)
		req := httptest.NewRequest(http.MethodPost, "/api/playlists/reorder", bytes.NewBuffer(body))
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("Expected 200, got %d", w.Code)
		}

		req = httptest.NewRequest(http.MethodGet, "/api/playlists/items?title=Test%20Playlist", nil)
		w = httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		var items []struct {
			Path string `json:"path"`
		}
		json.NewDecoder(w.Body).Decode(&items)

		if items[0].Path != "item3.mp4" {
			t.Errorf("Expected item3 first, got %s", items[0].Path)
		}
		if items[1].Path != "item1.mp4" {
			t.Errorf("Expected item1 second, got %s", items[1].Path)
		}
		if items[2].Path != "item2.mp4" {
			t.Errorf("Expected item2 third, got %s", items[2].Path)
		}
	})

	t.Run("MultiDBReorder", func(t *testing.T) {
		// Setup second DB
		db2Path := fixture.TempDir + "/test2.db"
		db2, err := sql.Open("sqlite3", db2Path)
		if err != nil {
			t.Fatal(err)
		}
		InitDB(db2)
		defer db2.Close()

		// Insert data into DB2
		// item4 in DB2, track 3 (global)
		_, err = db2.Exec(`
			INSERT INTO media (path, type, time_deleted)
			VALUES ('item4.mp4', 'video/mp4', 0);

			INSERT INTO playlists (title, path, time_deleted)
			VALUES ('Test Playlist', 'custom:123', 0);

			INSERT INTO playlist_items (playlist_id, media_path, track_number, time_added)
			VALUES (1, 'item4.mp4', 3, 400);
		`)
		if err != nil {
			t.Fatal(err)
		}

		// Reset DB1 state (item1: 0, item3: 1, item2: 2) from previous tests
		// Wait, previous tests modified DB1 state.
		// Current state of DB1 from MoveEndToMiddle:
		// item3 (0), item1 (1), item2 (2).
		// Plus DB2 item4 (3).
		// Global order: item3, item1, item2, item4.

		cmd := &ServeCmd{
			Databases: []string{fixture.DBPath, db2Path},
		}
		handler := cmd.Mux()

		// Move item4 (index 3, from DB2) to index 0 (start)
		// Expected: item4, item3, item1, item2
		payload := map[string]any{
			"playlist_title": "Test Playlist",
			"media_path":     "item4.mp4",
			"new_index":      0,
		}
		body, _ := json.Marshal(payload)
		req := httptest.NewRequest(http.MethodPost, "/api/playlists/reorder", bytes.NewBuffer(body))
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("Expected 200, got %d: %s", w.Code, w.Body.String())
		}

		// Verify order via API
		req = httptest.NewRequest(http.MethodGet, "/api/playlists/items?title=Test%20Playlist", nil)
		w = httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		var items []struct {
			Path        string `json:"path"`
			TrackNumber *int64 `json:"track_number"`
		}
		json.NewDecoder(w.Body).Decode(&items)

		if len(items) != 4 {
			t.Fatalf("Expected 4 items, got %d", len(items))
		}

		// We expect item4 at index 0
		if items[0].Path != "item4.mp4" {
			t.Errorf("Expected item4 first, got %s (track: %v)", items[0].Path, items[0].TrackNumber)
		}
		if items[1].Path != "item3.mp4" {
			t.Errorf("Expected item3 second, got %s (track: %v)", items[1].Path, items[1].TrackNumber)
		}
		if items[2].Path != "item1.mp4" {
			t.Errorf("Expected item1 third, got %s (track: %v)", items[2].Path, items[2].TrackNumber)
		}
		if items[3].Path != "item2.mp4" {
			t.Errorf("Expected item2 fourth, got %s (track: %v)", items[3].Path, items[3].TrackNumber)
		}
	})
}
