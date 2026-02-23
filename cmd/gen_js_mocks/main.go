package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/chapmanjacobd/discotheque/internal/models"
)

func ptr[T any](v T) *T {
	return &v
}

func main() {
	mocks := make(map[string]interface{})

	// Media mock
	mocks["media"] = []models.MediaWithDB{
		{
			Media: models.Media{
				Path:     "video1.mp4",
				Type:     ptr("video/mp4"),
				Size:     ptr(int64(1024 * 1024 * 50)),
				Duration: ptr(int64(120)),
				Score:    ptr(float64(5)),
			},
			DB: "test.db",
		},
		{
			Media: models.Media{
				Path:     "audio1.mp3",
				Type:     ptr("audio/mpeg"),
				Size:     ptr(int64(1024 * 1024 * 5)),
				Duration: ptr(int64(180)),
				Score:    ptr(float64(4)),
			},
			DB: "test.db",
		},
		{
			Media: models.Media{
				Path: "image1.jpg",
				Type: ptr("image/jpeg"),
				Size: ptr(int64(1024 * 500)),
			},
			DB: "test.db",
		},
	}

	// Database info mock
	mocks["databases"] = models.DatabaseInfo{
		Databases:      []string{"test.db"},
		Trashcan:       true,
		GlobalProgress: true,
		Dev:            false,
	}

	// Categories mock
	mocks["categories"] = []models.CatStat{
		{Category: "comedy", Count: 5},
		{Category: "music", Count: 3},
	}

	// Genres mock
	mocks["genres"] = []models.GenreStat{
		{Genre: "Rock", Count: 10},
		{Genre: "Jazz", Count: 2},
	}

	// Ratings mock
	mocks["ratings"] = []models.RatStat{
		{Rating: 5, Count: 1},
		{Rating: 0, Count: 10},
	}

	// Playlists mock
	mocks["playlists"] = []models.Playlist{
		{ID: 1, Title: ptr("My Playlist"), DB: "test.db"},
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(mocks); err != nil {
		fmt.Fprintf(os.Stderr, "Error encoding JSON: %v\n", err)
		os.Exit(1)
	}
}
