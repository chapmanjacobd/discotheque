package utils

import (
	"os"
	"os/signal"
	"sync"
	"syscall"

	"golang.org/x/term"
)

// TerminalWidth tracks the current terminal width and updates on resize
type TerminalWidth struct {
	mu       sync.RWMutex
	width    int
	initOnce sync.Once
}

var terminalWidth TerminalWidth

// GetWidth returns the current terminal width
func GetTerminalWidth() int {
	terminalWidth.initOnce.Do(func() {
		terminalWidth.updateWidth()
		terminalWidth.watchResize()
	})
	terminalWidth.mu.RLock()
	defer terminalWidth.mu.RUnlock()
	return terminalWidth.width
}

// watchResize listens for terminal resize events
func (t *TerminalWidth) watchResize() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGWINCH)
	go func() {
		for range sigChan {
			t.updateWidth()
		}
	}()
}

// updateWidth gets the current terminal width
func (t *TerminalWidth) updateWidth() {
	w, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		w = 80 // Fallback
	}
	t.mu.Lock()
	t.width = w
	t.mu.Unlock()
}

// TruncateMiddle truncates a string in the middle with ellipsis
func TruncateMiddle(s string, max int) string {
	if s == "" {
		return ""
	}
	if len(s) <= max {
		return s
	}
	half := max / 2
	if half < 2 {
		half = 2
	}
	return s[:half-1] + "…" + s[len(s)-half+1:]
}

// GetPathDisplayWidth returns the recommended width for displaying paths
// Takes into account terminal width and leaves room for other columns
func GetPathDisplayWidth() int {
	width := GetTerminalWidth()
	// Reserve space for table columns (about 70 chars for table)
	// Use remaining space for path, but cap at reasonable length
	pathWidth := min(max(width-70, 40), 80)
	return pathWidth
}
