package commands

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/chapmanjacobd/discotheque/internal/testutils"
)

func TestServeCmd_Handlers(t *testing.T) {
	fixture := testutils.Setup(t)
	defer fixture.Cleanup()

	db := fixture.GetDB()
	InitDB(db)

	// Insert test data
	_, err := db.Exec(`
		INSERT INTO media (path, type, categories, score, time_deleted)
		VALUES 
		('video1.mp4', 'video/mp4', ';comedy;sports;', 5.0, 0),
		('video2.mp4', 'video/mp4', ';comedy;', 4.0, 0),
		('audio1.mp3', 'audio/mpeg', ';music;', 0.0, 0),
		('audio2.mp3', 'audio/mpeg', ';music;', 3.0, 0)
	`)
	if err != nil {
		t.Fatal(err)
	}
	db.Close()

	cmd := &ServeCmd{
		Databases: []string{fixture.DBPath},
	}
	handler := cmd.Mux()

	t.Run("HandleCategories", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/categories", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected 200, got %d", w.Code)
		}

		var resp []struct {
			Category string `json:"category"`
			Count    int    `json:"count"`
		}
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatal(err)
		}

		foundComedy := false
		for _, cat := range resp {
			if cat.Category == "comedy" {
				foundComedy = true
				if cat.Count != 2 {
					t.Errorf("Expected 2 comedy files, got %d", cat.Count)
				}
			}
		}
		if !foundComedy {
			t.Error("Category 'comedy' not found in response")
		}
	})

	t.Run("HandleRatings", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/ratings", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected 200, got %d", w.Code)
		}

		var resp []struct {
			Rating int `json:"rating"`
			Count  int `json:"count"`
		}
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatal(err)
		}

		// Ratings 5, 4, 3, 0
		if len(resp) != 4 {
			t.Errorf("Expected 4 ratings, got %d", len(resp))
		}
	})

	t.Run("HandleQuery", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/query?category=comedy", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected 200, got %d", w.Code)
		}

		var resp []any
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatal(err)
		}

		if len(resp) != 2 {
			t.Errorf("Expected 2 results for category=comedy, got %d", len(resp))
		}
	})

	t.Run("HandleRate", func(t *testing.T) {
		// Test updating rating
		payload := `{"path": "video1.mp4", "score": 2}`
		req := httptest.NewRequest(http.MethodPost, "/api/rate", strings.NewReader(payload))
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected 200, got %d", w.Code)
		}

		// Verify in DB
		db := fixture.GetDB()
		defer db.Close()
		var score float64
		err := db.QueryRow("SELECT score FROM media WHERE path = 'video1.mp4'").Scan(&score)
		if err != nil {
			t.Fatal(err)
		}
		if score != 2.0 {
			t.Errorf("Expected score 2.0, got %f", score)
		}
	})

	t.Run("HandleDelete", func(t *testing.T) {
		payload := `{"path": "video2.mp4", "restore": false}`
		req := httptest.NewRequest(http.MethodPost, "/api/delete", strings.NewReader(payload))
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected 200, got %d", w.Code)
		}

		// Verify in DB
		db := fixture.GetDB()
		defer db.Close()
		var timeDeleted int64
		err := db.QueryRow("SELECT time_deleted FROM media WHERE path = 'video2.mp4'").Scan(&timeDeleted)
		if err != nil {
			t.Fatal(err)
		}
		if timeDeleted == 0 {
			t.Error("Expected time_deleted to be non-zero")
		}
	})

	t.Run("HandleProgress", func(t *testing.T) {
		payload := `{"path": "audio1.mp3", "playhead": 120, "duration": 300}`
		req := httptest.NewRequest(http.MethodPost, "/api/progress", strings.NewReader(payload))
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected 200, got %d", w.Code)
		}

		// Verify in DB
		db := fixture.GetDB()
		defer db.Close()
		var playhead int64
		err := db.QueryRow("SELECT playhead FROM media WHERE path = 'audio1.mp3'").Scan(&playhead)
		if err != nil {
			t.Fatal(err)
		}
		if playhead != 120 {
			t.Errorf("Expected playhead 120, got %d", playhead)
		}
	})

	t.Run("HandleRaw", func(t *testing.T) {
		// Create a real dummy file
		dummyPath := fixture.CreateDummyFile("real_video.mp4")

		// Add it to DB
		db := fixture.GetDB()
		db.Exec("INSERT INTO media (path, type, time_deleted) VALUES (?, 'video/mp4', 0)", dummyPath)
		db.Close()

		req := httptest.NewRequest(http.MethodGet, "/api/raw?path="+dummyPath, nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected 200, got %d", w.Code)
		}
	})

	t.Run("HandleEvents", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/events", nil)
		// We can't easily test SSE with httptest.Recorder if it doesn't flush or if it blocks.
		// But we can check if it returns the right content type.
		w := httptest.NewRecorder()

		// Use a channel to timeout if it blocks
		done := make(chan bool)
		go func() {
			handler.ServeHTTP(w, req)
			done <- true
		}()

		// Give it a tiny bit of time then "cancel" or just check what we have
		select {
		case <-done:
		case <-time.After(10 * time.Millisecond):
			// SSE handler might block forever, which is fine for this test if we just want to check headers
		}

		if w.Header().Get("Content-Type") != "text/event-stream" {
			t.Errorf("Expected text/event-stream, got %s", w.Header().Get("Content-Type"))
		}
	})

	t.Run("HandleSubtitles", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/subtitles?path=video1.mp4", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		// Returns 404 because no subtitle file exists on disk
		if w.Code != http.StatusNotFound {
			t.Errorf("Expected 404, got %d", w.Code)
		}
	})

	t.Run("HandleThumbnail", func(t *testing.T) {
		dummyPath := fixture.CreateDummyFile("thumb_test.mp4")
		db := fixture.GetDB()
		db.Exec("INSERT INTO media (path, type, time_deleted) VALUES (?, 'video/mp4', 0)", dummyPath)
		db.Close()

		req := httptest.NewRequest(http.MethodGet, "/api/thumbnail?path="+dummyPath, nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		// It will fail because it's a dummy file, but it should be 200/500 depending on where it fails
		// If ffmpeg fails, it returns 500
		if w.Code == http.StatusNotFound {
			t.Errorf("Expected 200 or 500, got 404")
		}
	})

	t.Run("HandleOPDS", func(t *testing.T) {
		db := fixture.GetDB()
		db.Exec("INSERT INTO media (path, type, time_deleted) VALUES (?, 'application/pdf', 0)", "book.pdf")
		db.Close()

		req := httptest.NewRequest(http.MethodGet, "/opds", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected 200, got %d", w.Code)
		}
		if !strings.Contains(w.Body.String(), "<feed") {
			t.Errorf("Expected OPDS feed, got %s", w.Body.String())
		}
	})
}
