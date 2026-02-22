package db

import (
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var (
	repairLocks sync.Map
)

func getLock(path string) *sync.Mutex {
	v, _ := repairLocks.LoadOrStore(path, &sync.Mutex{})
	return v.(*sync.Mutex)
}

// IsCorruptionError checks if the error is a database corruption error
func IsCorruptionError(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "database disk image is malformed")
}

// Repair attempts to repair a corrupted SQLite database
func Repair(dbPath string) error {
	mu := getLock(dbPath)
	mu.Lock()
	defer mu.Unlock()

	// Check if it's actually corrupt (maybe fixed by previous repair in race)
	if isHealthy(dbPath) {
		slog.Info("Database appears healthy (race condition handled), skipping repair", "path", dbPath)
		return nil
	}

	if _, err := exec.LookPath("sqlite3"); err != nil {
		return fmt.Errorf("sqlite3 command line tool is required for auto-repair")
	}

	// Backup
	backupPath := fmt.Sprintf("%s.corrupt.%d.bak", dbPath, time.Now().Unix())
	slog.Info("Backing up corrupted database", "src", dbPath, "dst", backupPath)
	
	// Check if file exists before renaming
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		return fmt.Errorf("database file not found: %s", dbPath)
	}

	if err := os.Rename(dbPath, backupPath); err != nil {
		return fmt.Errorf("failed to backup database: %w", err)
	}

	tempPath := dbPath + ".recovering"

	// Attempt .recover
	slog.Info("Attempting recovery using '.recover'...")
	cmdRecover := exec.Command("bash", "-c", fmt.Sprintf("sqlite3 \"%s\" \".recover\" | sqlite3 \"%s\"", backupPath, tempPath))
	out, err := cmdRecover.CombinedOutput()

	// Check if recovery worked by verifying the new file
	if err == nil && isHealthy(tempPath) {
		slog.Info("Recovery successful via .recover")
		if err := os.Rename(tempPath, dbPath); err != nil {
			return fmt.Errorf("failed to restore recovered database: %w", err)
		}
		os.Remove(backupPath)
		return nil
	}
	slog.Warn(".recover failed or produced invalid DB, trying .dump", "error", err, "output", string(out))

	// Cleanup failed attempt
	os.Remove(tempPath)

	// Fallback to .dump
	slog.Info("Attempting recovery using '.dump'...")
	cmdDump := exec.Command("bash", "-c", fmt.Sprintf("sqlite3 \"%s\" \".dump\" | sqlite3 \"%s\"", backupPath, tempPath))
	out, err = cmdDump.CombinedOutput()
	if err != nil {
		os.Remove(tempPath)
		return fmt.Errorf("recovery failed: %v\nOutput: %s", err, string(out))
	}

	if isHealthy(tempPath) {
		slog.Info("Recovery successful via .dump")
		if err := os.Rename(tempPath, dbPath); err != nil {
			return fmt.Errorf("failed to restore recovered database: %w", err)
		}
		os.Remove(backupPath)
		return nil
	}

	os.Remove(tempPath)
	return fmt.Errorf("all recovery attempts produced invalid databases")
}

func isHealthy(dbPath string) bool {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return false
	}
	defer db.Close()

	// Try a simple query to check header
	var s string
	err = db.QueryRow("PRAGMA integrity_check(1)").Scan(&s)
	if err != nil {
		return false
	}
	return s == "ok"
}
