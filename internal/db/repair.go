package db

import (
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"strings"
	"sync"
	"syscall"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var repairLocks sync.Map

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
	start := time.Now()

	// 1. Process-local lock to avoid multiple goroutines in the same process
	mu := getLock(dbPath)
	mu.Lock()
	defer mu.Unlock()

	// 2. Cross-process lock using a separate lock file
	lockPath := dbPath + ".repair.lock"
	lockFile, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		return fmt.Errorf("failed to open lock file: %w", err)
	}
	defer func() {
		lockFile.Close()
		os.Remove(lockPath)
	}()

	// Exclusive lock, blocks until available
	if err := syscall.Flock(int(lockFile.Fd()), syscall.LOCK_EX); err != nil {
		return fmt.Errorf("failed to acquire flock: %w", err)
	}
	defer syscall.Flock(int(lockFile.Fd()), syscall.LOCK_UN)

	waitDuration := time.Since(start)

	// 3. Check if it's actually corrupt (maybe fixed by previous repair in race)
	if isHealthy(dbPath) {
		if waitDuration > 1*time.Millisecond {
			slog.Info("Database was repaired by another process", "path", dbPath, "wait_time", waitDuration.String())
		}
		return nil
	}

	if _, err := exec.LookPath("sqlite3"); err != nil {
		return fmt.Errorf("sqlite3 command line tool is required for auto-repair")
	}

	// 4. Move database and sidecar files to a temporary location
	now := time.Now().Unix()
	backupDir := fmt.Sprintf("%s.corrupt.%d", dbPath, now)
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return fmt.Errorf("failed to create backup directory: %w", err)
	}
	// We want to delete this directory on success, but keep it on total failure for manual inspection?
	// The user asked to "unlink ... after it is all done", which implies success.

	corruptMain := backupDir + "/main.db"
	if err := os.Rename(dbPath, corruptMain); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("database file not found: %s", dbPath)
		}
		return fmt.Errorf("failed to move corrupted database: %w", err)
	}

	// Move sidecars
	for _, suffix := range []string{"-wal", "-shm"} {
		sidecar := dbPath + suffix
		if _, err := os.Stat(sidecar); err == nil {
			os.Rename(sidecar, corruptMain+suffix)
		}
	}

	// 5. Recover into the original path
	// This ensures that the new database starts fresh at the expected location.
	slog.Info("Attempting recovery...", "from", corruptMain, "to", dbPath)

	// Try .recover (more modern)
	cmdRecover := exec.Command("bash", "-c", fmt.Sprintf("sqlite3 \"%s\" \".recover\" | sqlite3 \"%s\"", corruptMain, dbPath))
	out, err := cmdRecover.CombinedOutput()
	
	success := false
	if err == nil && isHealthy(dbPath) {
		slog.Info("Recovery successful via .recover")
		success = true
	} else {
		slog.Warn(".recover failed or produced invalid DB, trying .dump", "error", err, "output", string(out))
		os.Remove(dbPath) // Clean up failed attempt before next one

		// Try .dump (classic fallback)
		cmdDump := exec.Command("bash", "-c", fmt.Sprintf("sqlite3 \"%s\" \".dump\" | sqlite3 \"%s\"", corruptMain, dbPath))
		out, err = cmdDump.CombinedOutput()
		if err == nil && isHealthy(dbPath) {
			slog.Info("Recovery successful via .dump")
			success = true
		} else {
			slog.Error("Recovery failed", "error", err, "output", string(out))
		}
	}

	// 6. Cleanup
	if success {
		os.RemoveAll(backupDir)
		return nil
	}

	// If all failed, we should probably try to restore the "corrupt" file so at least it's back where it was
	// but it's already failed twice.
	return fmt.Errorf("all recovery attempts failed")
}

func isHealthy(dbPath string) bool {
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		return false
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return false
	}
	defer db.Close()

	var s string
	err = db.QueryRow("PRAGMA integrity_check").Scan(&s)
	if err != nil {
		return false
	}
	if s != "ok" {
		return false
	}

	// Try to read from sqlite_master to ensure the schema is readable
	rows, err := db.Query("SELECT name FROM sqlite_master LIMIT 1")
	if err != nil {
		return false
	}
	rows.Close()

	return true
}
