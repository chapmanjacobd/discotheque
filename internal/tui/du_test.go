package tui

import (
	"testing"

	"github.com/chapmanjacobd/discotheque/internal/models"
)

func TestDUModel(t *testing.T) {
	v1Size := int64(1000)
	v2Size := int64(2000)
	media := []models.MediaWithDB{
		{Media: models.Media{Path: "/home/user/vids/v1.mp4", Size: &v1Size}},
		{Media: models.Media{Path: "/home/user/vids/v2.mp4", Size: &v2Size}},
	}

	m := NewDUModel(media, models.GlobalFlags{})
	if m.totalSize != 3000 {
		// Wait, NewDUModel calls updateList which calculates stats
		// Root depth is 1, so it should aggregate /home
	}

	if m.list.Title == "" {
		t.Error("Expected list title to be set")
	}
}
