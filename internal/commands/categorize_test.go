package commands

import (
	"database/sql"
	"testing"

	"github.com/chapmanjacobd/discotheque/internal/testutils"
)

func TestCategorizeCmd_Run(t *testing.T) {
	fixture := testutils.Setup(t)
	defer fixture.Cleanup()

	// Create some files with keywords in path
	f1 := fixture.CreateDummyFile("sports football.mp4")
	f2 := fixture.CreateDummyFile("fitness workout.mp4")
	f3 := fixture.CreateDummyFile("unknown movie.mp4")

	addCmd := &AddCmd{
		Args: []string{fixture.DBPath, f1, f2, f3},
	}
	addCmd.AfterApply()
	addCmd.Run(nil)

	t.Run("ApplyCategories", func(t *testing.T) {
		cmd := &CategorizeCmd{
			Databases: []string{fixture.DBPath},
		}
		if err := cmd.Run(nil); err != nil {
			t.Fatalf("CategorizeCmd failed: %v", err)
		}

		// Verify categories
		dbConn := fixture.GetDB()
		defer dbConn.Close()

		var cat1, cat2, cat3 sql.NullString
		dbConn.QueryRow("SELECT categories FROM media WHERE path = ?", f1).Scan(&cat1)
		dbConn.QueryRow("SELECT categories FROM media WHERE path = ?", f2).Scan(&cat2)
		dbConn.QueryRow("SELECT categories FROM media WHERE path = ?", f3).Scan(&cat3)

		if cat1.String != ";sports;" {
			t.Errorf("Expected ;sports; for f1, got %s", cat1.String)
		}
		if cat2.String != ";fitness;" {
			t.Errorf("Expected ;fitness; for f2, got %s", cat2.String)
		}
		if cat3.Valid && cat3.String != "" {
			t.Errorf("Expected no categories for f3, got %s", cat3.String)
		}
	})

	t.Run("MineCategories", func(t *testing.T) {
		cmd := &CategorizeCmd{
			Databases: []string{fixture.DBPath},
			Other:     true,
		}
		if err := cmd.Run(nil); err != nil {
			t.Fatalf("CategorizeCmd (Other) failed: %v", err)
		}
		// mineCategories just prints to stdout, so we just check it doesn't crash here
		// unless we want to capture stdout and verify
	})
}
