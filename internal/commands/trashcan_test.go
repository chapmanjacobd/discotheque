package commands

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/chapmanjacobd/discotheque/internal/testutils"
)

func TestServeCmd_TrashcanDisabled(t *testing.T) {
	fixture := testutils.Setup(t)
	defer fixture.Cleanup()

	cmd := &ServeCmd{
		Databases: []string{fixture.DBPath},
		Trashcan:  false,
	}
	handler := cmd.Mux()

	// 1. Verify /api/databases returns trashcan: false
	req := httptest.NewRequest(http.MethodGet, "/api/databases", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var resp struct {
		Trashcan bool `json:"trashcan"`
	}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}
	if resp.Trashcan {
		t.Errorf("Expected trashcan to be false in API response")
	}

	// 2. Verify /api/trash returns 404
	req = httptest.NewRequest(http.MethodGet, "/api/trash", nil)
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("Expected 404 for /api/trash when disabled, got %d", w.Code)
	}

	// 3. Verify /api/empty-bin returns 404
	req = httptest.NewRequest(http.MethodPost, "/api/empty-bin", nil)
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("Expected 404 for /api/empty-bin when disabled, got %d", w.Code)
	}
}

func TestServeCmd_TrashcanEnabled(t *testing.T) {
	fixture := testutils.Setup(t)
	defer fixture.Cleanup()

	// Initialize the DB table
	db := fixture.GetDB()
	InitDB(db)
	db.Close()

	cmd := &ServeCmd{
		Databases: []string{fixture.DBPath},
		Trashcan:  true,
	}
	handler := cmd.Mux()

	// 1. Verify /api/databases returns trashcan: true
	req := httptest.NewRequest(http.MethodGet, "/api/databases", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	var resp struct {
		Trashcan bool `json:"trashcan"`
	}
	json.NewDecoder(w.Body).Decode(&resp)
	if !resp.Trashcan {
		t.Errorf("Expected trashcan to be true in API response")
	}

	// 2. Verify /api/trash is registered (should return 200, even if empty)
	req = httptest.NewRequest(http.MethodGet, "/api/trash", nil)
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("Expected 200 for /api/trash when enabled, got %d", w.Code)
	}

	// 3. Verify /api/empty-bin is registered
	req = httptest.NewRequest(http.MethodPost, "/api/empty-bin", nil)
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	// Might return error if no body, but should NOT be 404
	if w.Code == http.StatusNotFound {
		t.Errorf("Expected /api/empty-bin to be registered")
	}
}
