package aggregate

import (
	"testing"

	"github.com/chapmanjacobd/discotheque/internal/db"
)

func TestByFolder(t *testing.T) {
	media := []db.Media{
		{Path: "/movies/action/movie1.mp4", Size: 1000, Duration: 7200},
		{Path: "/movies/action/movie2.mp4", Size: 2000, Duration: 6000},
		{Path: "/movies/comedy/movie3.mp4", Size: 1500, Duration: 5400},
		{Path: "/tv/show/episode1.mp4", Size: 500, Duration: 1800},
	}

	folders := ByFolder(media)

	if len(folders) != 3 {
		t.Errorf("Expected 3 folders, got %d", len(folders))
	}

	// Find action folder
	var actionFolder *FolderStats
	for i := range folders {
		if folders[i].Path == "/movies/action" {
			actionFolder = &folders[i]
			break
		}
	}

	if actionFolder == nil {
		t.Fatal("Action folder not found")
	}

	if actionFolder.Count != 2 {
		t.Errorf("Expected count 2, got %d", actionFolder.Count)
	}

	if actionFolder.TotalSize != 3000 {
		t.Errorf("Expected total size 3000, got %d", actionFolder.TotalSize)
	}

	if actionFolder.AvgSize != 1500 {
		t.Errorf("Expected avg size 1500, got %d", actionFolder.AvgSize)
	}

	if actionFolder.TotalDuration != 13200 {
		t.Errorf("Expected total duration 13200, got %d", actionFolder.TotalDuration)
	}
}

func TestSortFolders(t *testing.T) {
	folders := []FolderStats{
		{Path: "/a", Count: 3, TotalSize: 3000},
		{Path: "/b", Count: 1, TotalSize: 5000},
		{Path: "/c", Count: 2, TotalSize: 1000},
	}

	SortFolders(folders, "count", false)

	if folders[0].Path != "/b" || folders[1].Path != "/c" || folders[2].Path != "/a" {
		t.Errorf("Sort by count failed: got %v", []string{folders[0].Path, folders[1].Path, folders[2].Path})
	}

	SortFolders(folders, "size", true)

	if folders[0].Path != "/b" {
		t.Errorf("Expected /b first when sorting by size desc, got %s", folders[0].Path)
	}
}
