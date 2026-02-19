package fs

import (
	"os"
	"path/filepath"
	"strings"
)

var MediaExtensions = map[string]bool{
	".mp4":  true,
	".mkv":  true,
	".avi":  true,
	".mov":  true,
	".webm": true,
	".flv":  true,
	".m4v":  true,
	".mpg":  true,
	".mpeg": true,
	".mp3":  true,
	".flac": true,
	".m4a":  true,
	".opus": true,
	".ogg":  true,
	".wav":  true,
}

func FindMedia(root string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if MediaExtensions[ext] {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}
