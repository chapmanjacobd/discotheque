package testutils

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

type TestFixture struct {
	T       *testing.T
	DBPath  string
	TempDir string
}

func Setup(t *testing.T) *TestFixture {
	tempDir, err := os.MkdirTemp("", "disco-test-*")
	if err != nil {
		t.Fatal(err)
	}

	// Get absolute path for tempDir to avoid confusion
	absTempDir, err := filepath.Abs(tempDir)
	if err != nil {
		t.Fatal(err)
	}

	dbPath := filepath.Join(absTempDir, "test.db")

	return &TestFixture{
		T:       t,
		DBPath:  dbPath,
		TempDir: absTempDir,
	}
}

func (f *TestFixture) Cleanup() {
	os.RemoveAll(f.TempDir)
}

func (f *TestFixture) CreateDummyFile(name string) string {
	path := filepath.Join(f.TempDir, name)
	f.writeFile(path, []byte("dummy data"))
	return path
}

func (f *TestFixture) CreateFileTree(tree map[string]any) {
	f.createTree(f.TempDir, tree)
}

func (f *TestFixture) createTree(parent string, tree map[string]any) {
	for name, content := range tree {
		path := filepath.Join(parent, name)
		if subTree, ok := content.(map[string]any); ok {
			if err := os.MkdirAll(path, 0755); err != nil {
				f.T.Fatal(err)
			}
			f.createTree(path, subTree)
		} else if strContent, ok := content.(string); ok {
			f.writeFile(path, []byte(strContent))
		} else {
			f.writeFile(path, []byte("dummy"))
		}
	}
}

func (f *TestFixture) writeFile(path string, data []byte) {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		f.T.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		f.T.Fatal(err)
	}
}

func (f *TestFixture) GetDB() *sql.DB {
	db, err := sql.Open("sqlite3", f.DBPath)
	if err != nil {
		f.T.Fatal(err)
	}
	return db
}
