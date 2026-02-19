package utils

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
)

func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func GetDefaultBrowser() string {
	switch runtime.GOOS {
	case "linux":
		return "xdg-open"
	case "darwin":
		return "open"
	case "windows":
		return "start"
	default:
		return "xdg-open"
	}
}

func IsSQLite(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()

	header := make([]byte, 16)
	if _, err := f.Read(header); err != nil {
		return false
	}
	return string(header) == "SQLite format 3\x00"
}

func ReadLines(r io.Reader) []string {
	var lines []string
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			lines = append(lines, line)
		}
	}
	return lines
}

func ExpandStdin(paths []string) []string {
	var out []string
	for _, p := range paths {
		if p == "-" {
			out = append(out, ReadLines(os.Stdin)...)
		} else {
			out = append(out, p)
		}
	}
	return out
}

func Confirm(message string) bool {
	fmt.Printf("%s [y/N]: ", message)
	var response string
	fmt.Scanln(&response)
	response = strings.ToLower(strings.TrimSpace(response))
	return response == "y" || response == "yes"
}

func Prompt(message string) string {
	fmt.Printf("%s: ", message)
	var response string
	fmt.Scanln(&response)
	return strings.TrimSpace(response)
}

