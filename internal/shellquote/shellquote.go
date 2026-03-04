package shellquote

import "strings"

// ShellQuote returns a shell-escaped version of the string
func ShellQuote(s string) string {
	if s == "" {
		return "''"
	}
	// If it doesn't contain any special characters, return it as is
	safeChars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-_./"
	allSafe := true
	for _, r := range s {
		if !strings.ContainsRune(safeChars, r) {
			allSafe = false
			break
		}
	}
	if allSafe {
		return s
	}

	return "'" + strings.ReplaceAll(s, "'", "'\\''") + "'"
}
