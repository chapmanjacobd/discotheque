//go:build !windows

package utils

import (
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"syscall"
)

// GetCommandPath returns the absolute path to a command
func GetCommandPath(name string) string {
	exeDir := getExecutableDir()
	if exeDir != "" {
		path := filepath.Join(exeDir, name)
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	path, err := exec.LookPath(name)
	if err != nil {
		return ""
	}
	return path
}

// watchResize listens for terminal resize events
func (t *TerminalSize) watchResize() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGWINCH)
	go func() {
		for range sigChan {
			t.updateSize()
		}
	}()
}
