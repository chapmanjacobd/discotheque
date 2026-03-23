package utils

import (
	"os"
	"path/filepath"
	"sync"

	"golang.org/x/term"
)

// commandExistsCache caches the result of command existence checks
var commandExistsCache sync.Map

// getExecutableDir returns the directory containing the current executable
var getExecutableDir = sync.OnceValue(func() string {
	exe, err := os.Executable()
	if err != nil {
		return ""
	}
	return filepath.Dir(exe)
})

// CommandExists checks if a command is available in PATH or common installation paths
// Results are cached to avoid repeated syscalls
func CommandExists(cmd string) bool {
	if cached, ok := commandExistsCache.Load(cmd); ok {
		return cached.(bool)
	}

	path := GetCommandPath(cmd)
	exists := path != ""
	commandExistsCache.Store(cmd, exists)
	return exists
}

// TerminalWidth tracks the current terminal width and updates on resize
type TerminalWidth struct {
	mu       sync.RWMutex
	width    int
	initOnce sync.Once
}

var terminalWidth TerminalWidth

// GetTerminalWidth returns the current terminal width
func GetTerminalWidth() int {
	terminalWidth.initOnce.Do(func() {
		terminalWidth.updateWidth()
		terminalWidth.watchResize()
	})
	terminalWidth.mu.RLock()
	defer terminalWidth.mu.RUnlock()
	return terminalWidth.width
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

// GetClearLineSequence returns the escape sequence to clear/overwrite a line.
// We use \x1b[K (Erase from cursor to end of line) which is standard for overwriting.
func GetClearLineSequence() string {
	return "\033[K"
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
