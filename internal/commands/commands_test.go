package commands

import (
	"database/sql"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/chapmanjacobd/discotheque/internal/models"
)

func TestFilesInfoCmd_Run(t *testing.T) {
	f, err := os.CreateTemp("", "test-*.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	f.WriteString("hello world")
	f.Close()

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cmd := &FilesInfoCmd{
		GlobalFlags: models.GlobalFlags{JSON: true},
		Paths:       []string{f.Name()},
	}
	if err := cmd.Run(nil); err != nil {
		t.Fatalf("FilesInfoCmd failed: %v", err)
	}

	w.Close()
	os.Stdout = oldStdout

	var buf strings.Builder
	io.Copy(&buf, r)
	output := buf.String()

	if !strings.Contains(output, "text/plain") {
		t.Errorf("Expected output to contain text/plain, got %s", output)
	}
}

func TestSearchDBCmd_Run(t *testing.T) {
	f, err := os.CreateTemp("", "sdb-test-*.db")
	if err != nil {
		t.Fatal(err)
	}
	dbPath := f.Name()
	f.Close()
	defer os.Remove(dbPath)

	dbConn, _ := sql.Open("sqlite3", dbPath)
	dbConn.Exec("CREATE TABLE test (name TEXT, val TEXT)")
	dbConn.Exec("INSERT INTO test VALUES ('apple', 'fruit'), ('carrot', 'vegetable')")
	dbConn.Close()

	t.Run("FuzzyTableMatching", func(t *testing.T) {
		cmd := &SearchDBCmd{
			GlobalFlags: models.GlobalFlags{JSON: true},
			Database:    dbPath,
			Table:       "tes", // fuzzy match for 'test'
			Search:      []string{"apple"},
		}
		if err := cmd.Run(nil); err != nil {
			t.Fatalf("SearchDBCmd failed: %v", err)
		}
	})

	t.Run("DeleteRows", func(t *testing.T) {
		cmd := &SearchDBCmd{
			GlobalFlags: models.GlobalFlags{DeleteRows: true},
			Database:    dbPath,
			Table:       "test",
			Search:      []string{"carrot"},
		}
		if err := cmd.Run(nil); err != nil {
			t.Fatalf("SearchDBCmd failed: %v", err)
		}

		// Verify deletion
		dbConn, _ = sql.Open("sqlite3", dbPath)
		defer dbConn.Close()
		var count int
		dbConn.QueryRow("SELECT COUNT(*) FROM test").Scan(&count)
		if count != 1 {
			t.Errorf("Expected 1 row left, got %d", count)
		}
	})
}

func TestMergeDBsCmd_Run(t *testing.T) {
	// 1. Setup Source 1
	src1, _ := os.CreateTemp("", "merge-src1-*.db")
	src1Path := src1.Name()
	src1.Close()
	defer os.Remove(src1Path)
	db1, _ := sql.Open("sqlite3", src1Path)
	db1.Exec("CREATE TABLE media (path TEXT PRIMARY KEY, title TEXT)")
	db1.Exec("INSERT INTO media VALUES ('/path1', 'Title 1')")
	db1.Close()

	// 2. Setup Source 2
	src2, _ := os.CreateTemp("", "merge-src2-*.db")
	src2Path := src2.Name()
	src2.Close()
	defer os.Remove(src2Path)
	db2, _ := sql.Open("sqlite3", src2Path)
	db2.Exec("CREATE TABLE media (path TEXT PRIMARY KEY, title TEXT)")
	db2.Exec("INSERT INTO media VALUES ('/path2', 'Title 2')")
	db2.Close()

	// 3. Setup Target
	target, _ := os.CreateTemp("", "merge-target-*.db")
	targetPath := target.Name()
	target.Close()
	defer os.Remove(targetPath)

	// 4. Run Merge
	cmd := &MergeDBsCmd{
		TargetDB:  targetPath,
		SourceDBs: []string{src1Path, src2Path},
	}
	if err := cmd.Run(nil); err != nil {
		t.Fatalf("MergeDBsCmd failed: %v", err)
	}

	// 5. Verify
	targetDB, _ := sql.Open("sqlite3", targetPath)
	defer targetDB.Close()
	var count int
	targetDB.QueryRow("SELECT COUNT(*) FROM media").Scan(&count)
	if count != 2 {
		t.Errorf("Expected 2 rows in merged database, got %d", count)
	}
}

func TestRegexSortCmd_Run(t *testing.T) {
	input := "red apple\nbroccoli\nyellow\ngreen\norange apple\nred apple\n"

	t.Run("DefaultSort", func(t *testing.T) {
		var out strings.Builder
		cmd := &RegexSortCmd{
			Reader: strings.NewReader(input),
			Writer: &out,
		}
		if err := cmd.Run(nil); err != nil {
			t.Fatalf("RegexSortCmd failed: %v", err)
		}
		output := out.String()
		if !strings.Contains(output, "broccoli") {
			t.Errorf("Output missing expected content: %s", output)
		}
	})

	t.Run("LineSortDup", func(t *testing.T) {
		var out strings.Builder
		cmd := &RegexSortCmd{
			GlobalFlags: models.GlobalFlags{
				LineSorts: []string{"dup", "natural"},
			},
			Reader: strings.NewReader(input),
			Writer: &out,
		}
		if err := cmd.Run(nil); err != nil {
			t.Fatalf("RegexSortCmd failed: %v", err)
		}
		output := out.String()
		if !strings.Contains(output, "red apple") {
			t.Errorf("Output missing expected content: %s", output)
		}
	})
}

func TestDeleteMediaItem(t *testing.T) {
	f, err := os.CreateTemp("", "delete-test")
	if err != nil {
		t.Fatal(err)
	}
	f.Close()
	
	m := models.MediaWithDB{
		Media: models.Media{Path: f.Name()},
	}
	
	if err := DeleteMediaItem(m); err != nil {
		t.Fatalf("DeleteMediaItem failed: %v", err)
	}
	
	if _, err := os.Stat(f.Name()); !os.IsNotExist(err) {
		t.Error("File still exists after DeleteMediaItem")
	}
}
