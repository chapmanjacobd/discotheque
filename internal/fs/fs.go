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
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".gif":  true,
	".webp": true,
	".tiff": true,
	".bmp":  true,
}

func FindMedia(root string) (map[string]os.FileInfo, error) {
	files := make(map[string]os.FileInfo)
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if MediaExtensions[ext] {
			info, err := d.Info()
			if err != nil {
				return nil // Skip files we can't access
			}
			files[path] = info
		}
		return nil
	})
	return files, err
}
