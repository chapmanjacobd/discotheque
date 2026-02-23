package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/chapmanjacobd/discotheque/internal/models"
)

func main() {
	// Pointer helpers
	str := func(s string) *string { return &s }
	int64p := func(i int64) *int64 { return &i }
	float64p := func(f float64) *float64 { return &f }

	type Mocks struct {
		Media      []models.MediaWithDB `json:"media"`
		Databases  models.DatabaseInfo  `json:"databases"`
		Categories []models.CatStat     `json:"categories"`
		Genres     []models.GenreStat   `json:"genres"`
		Ratings    []models.RatStat     `json:"ratings"`
		Playlists  []string             `json:"playlists"`
	}

	mocks := Mocks{
		Media: []models.MediaWithDB{
			{
				Media: models.Media{
					Path:     "video1.mp4",
					Type:     str("video/mp4"),
					Size:     int64p(1024 * 1024 * 50),
					Duration: int64p(120),
					Score:    float64p(5),
				},
				DB: "test.db",
			},
			{
				Media: models.Media{
					Path:     "audio1.mp3",
					Type:     str("audio/mpeg"),
					Size:     int64p(1024 * 1024 * 5),
					Duration: int64p(180),
					Score:    float64p(4),
				},
				DB: "test.db",
			},
			{
				Media: models.Media{
					Path: "image1.jpg",
					Type: str("image/jpeg"),
					Size: int64p(1024 * 500),
				},
				DB: "test.db",
			},
		},
		Databases: models.DatabaseInfo{
			Databases: []string{"test.db"},
			Trashcan:  true,
			ReadOnly:  false,
			Dev:       false,
		},
		Categories: []models.CatStat{
			{Category: "comedy", Count: 5},
			{Category: "music", Count: 3},
		},
		Genres: []models.GenreStat{
			{Genre: "Rock", Count: 10},
			{Genre: "Jazz", Count: 2},
		},
		Ratings: []models.RatStat{
			{Rating: 5, Count: 1},
			{Rating: 0, Count: 10},
		},
		Playlists: []string{"My Playlist"},
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(mocks); err != nil {
		fmt.Fprintf(os.Stderr, "Error encoding JSON: %v\n", err)
		os.Exit(1)
	}
}
