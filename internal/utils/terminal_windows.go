//go:build windows

package utils

import (
	"os"
	"os/exec"
	"path/filepath"
)

// GetCommandPath returns the absolute path to a command, searching in common Windows installation paths if not in PATH
func GetCommandPath(name string) string {
	exeDir := getExecutableDir()
	if exeDir != "" {
		// Check both with and without .exe
		for _, path := range []string{
			filepath.Join(exeDir, name),
			filepath.Join(exeDir, name+".exe"),
		} {
			if _, err := os.Stat(path); err == nil {
				return path
			}
		}
	}

	if path, err := exec.LookPath(name); err == nil {
		return path
	}

	// Check common Windows installation paths
	var searchPaths []string
	programFiles := os.Getenv("ProgramFiles")
	programFilesX86 := os.Getenv("ProgramFiles(x86)")

	switch name {
	case "ebook-convert":
		searchPaths = append(searchPaths, filepath.Join(programFiles, "Calibre2", "ebook-convert.exe"))
	case "magick", "convert":
		if dirs, err := filepath.Glob(filepath.Join(programFiles, "ImageMagick-*")); err == nil {
			for _, dir := range dirs {
				searchPaths = append(searchPaths, filepath.Join(dir, name+".exe"))
			}
		}
	case "lsar", "unar":
		searchPaths = append(searchPaths, filepath.Join(programFiles, "Universal Extractor 2", "bin", name+".exe"))
		searchPaths = append(searchPaths, filepath.Join(programFilesX86, "Universal Extractor 2", "bin", name+".exe"))
	}

	for _, path := range searchPaths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	return ""
}

// watchResize is a no-op on Windows
func (t *TerminalSize) watchResize() {}
