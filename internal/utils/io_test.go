package utils

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestFileExists(t *testing.T) {
	f, _ := os.CreateTemp("", "exists-test")
	defer os.Remove(f.Name())
	f.Close()

	if !FileExists(f.Name()) {
		t.Errorf("FileExists(%s) should be true", f.Name())
	}
	if FileExists("/non/existent/path") {
		t.Error("FileExists should be false for non-existent path")
	}
}

func TestDirExists(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "dir-exists-test")
	defer os.RemoveAll(tmpDir)

	if !DirExists(tmpDir) {
		t.Errorf("DirExists(%s) should be true", tmpDir)
	}

	f, _ := os.CreateTemp(tmpDir, "file")
	defer os.Remove(f.Name())
	f.Close()

	if DirExists(f.Name()) {
		t.Errorf("DirExists(%s) should be false for file", f.Name())
	}
}

func TestGetDefaultBrowser(t *testing.T) {
	got := GetDefaultBrowser()
	if got == "" {
		t.Error("GetDefaultBrowser returned empty string")
	}
}

func TestIsSQLite(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "sqlite-test")
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "test.db")
	os.WriteFile(dbPath, []byte("SQLite format 3\x00extra data"), 0o644)

	if !IsSQLite(dbPath) {
		t.Error("IsSQLite should be true for valid header")
	}

	notDbPath := filepath.Join(tmpDir, "not.db")
	os.WriteFile(notDbPath, []byte("Not a sqlite file"), 0o644)
	if IsSQLite(notDbPath) {
		t.Error("IsSQLite should be false for invalid header")
	}

	if IsSQLite("/non/existent") {
		t.Error("IsSQLite should be false for non-existent file")
	}
}

func TestReadLines(t *testing.T) {
	input := `line1
  line2  

line3
`
	r := strings.NewReader(input)
	got := ReadLines(r)
	expected := []string{"line1", "line2", "line3"}
	if !reflect.DeepEqual(got, expected) {
		t.Errorf("ReadLines failed: got %v, want %v", got, expected)
	}
}

func TestExpandStdin(t *testing.T) {
	// This is hard to test because it uses os.Stdin
	// We can skip it or mock os.Stdin
}
