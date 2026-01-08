package aggregate

import (
	"path/filepath"
	"sort"

	"github.com/chapmanjacobd/discotheque/internal/db"
)

type FolderStats struct {
	Path          string
	Count         int
	TotalSize     int64
	TotalDuration int32
	AvgSize       int64
	AvgDuration   int32
	Files         []db.Media
}

func ByFolder(media []db.Media) []FolderStats {
	folders := make(map[string]*FolderStats)

	for _, m := range media {
		parent := filepath.Dir(m.Path)
		if _, exists := folders[parent]; !exists {
			folders[parent] = &FolderStats{
				Path:  parent,
				Files: []db.Media{},
			}
		}

		f := folders[parent]
		f.Count++
		f.TotalSize += m.Size
		f.TotalDuration += m.Duration
		f.Files = append(f.Files, m)
	}

	var result []FolderStats
	for _, f := range folders {
		if f.Count > 0 {
			f.AvgSize = f.TotalSize / int64(f.Count)
			f.AvgDuration = f.TotalDuration / int32(f.Count)
		}
		result = append(result, *f)
	}

	return result
}

func SortFolders(folders []FolderStats, by string, reverse bool) {
	less := func(i, j int) bool {
		switch by {
		case "count":
			return folders[i].Count < folders[j].Count
		case "size":
			return folders[i].TotalSize < folders[j].TotalSize
		case "duration":
			return folders[i].TotalDuration < folders[j].TotalDuration
		default:
			return folders[i].Path < folders[j].Path
		}
	}

	if reverse {
		sort.Slice(folders, func(i, j int) bool { return !less(i, j) })
	} else {
		sort.Slice(folders, less)
	}
}
