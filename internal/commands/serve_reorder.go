package commands

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"sort"
	"strings"

	database "github.com/chapmanjacobd/discoteca/internal/db"
	"github.com/chapmanjacobd/discoteca/internal/models"
)

type playlistItem struct {
	DBPath      string
	PlaylistID  int64
	MediaPath   string
	TrackNumber int64
	TimeAdded   int64
}

func (c *ServeCmd) processPlaylistItems(
	ctx context.Context,
	queries *database.Queries,
	playlistID int64,
	dbPath string,
) ([]playlistItem, error) {
	items, err := queries.GetPlaylistItems(ctx, playlistID)
	if err != nil {
		return nil, err
	}

	var res []playlistItem
	for _, item := range items {
		tn := int64(0)
		if item.TrackNumber.Valid {
			tn = item.TrackNumber.Int64
		}
		ta := int64(0)
		if item.TimeAdded.Valid {
			ta = item.TimeAdded.Int64
		}

		res = append(res, playlistItem{
			DBPath:      dbPath,
			PlaylistID:  playlistID,
			MediaPath:   item.Path,
			TrackNumber: tn,
			TimeAdded:   ta,
		})
	}
	return res, nil
}

func (c *ServeCmd) gatherPlaylistItems(ctx context.Context, playlistTitle string) ([]playlistItem, error) {
	var allItems []playlistItem
	for _, dbPath := range c.Databases {
		err := c.execDB(ctx, dbPath, func(ctx context.Context, sqlDB *sql.DB) error {
			queries := database.New(sqlDB)
			pls, err := queries.GetPlaylists(ctx)
			if err != nil {
				return err
			}

			var playlistID int64 = -1
			for _, p := range pls {
				if p.Title.Valid && strings.EqualFold(p.Title.String, playlistTitle) {
					playlistID = p.ID
					break
				}
			}

			if playlistID == -1 {
				return nil
			}

			items, err := c.processPlaylistItems(ctx, queries, playlistID, dbPath)
			if err != nil {
				return err
			}
			allItems = append(allItems, items...)
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	return allItems, nil
}

func (c *ServeCmd) reorderItems(allItems []playlistItem, mediaPath string, newIndex int) ([]playlistItem, error) {
	// Sort globally to establish current order matching frontend
	sort.Slice(allItems, func(i, j int) bool {
		if allItems[i].TrackNumber != allItems[j].TrackNumber {
			return allItems[i].TrackNumber < allItems[j].TrackNumber
		}
		// Match frontend sort: Path
		return allItems[i].MediaPath < allItems[j].MediaPath
	})

	// Find item to move
	currentIndex := -1
	var itemToMove playlistItem

	for i, item := range allItems {
		if item.MediaPath == mediaPath {
			currentIndex = i
			itemToMove = item
			break
		}
	}

	if currentIndex == -1 {
		return nil, errors.New("item not found in playlist")
	}

	// Reorder list in memory
	newItems := make([]playlistItem, 0, len(allItems))
	newItems = append(newItems, allItems[:currentIndex]...)
	newItems = append(newItems, allItems[currentIndex+1:]...)

	// Clamp
	if newIndex < 0 {
		newIndex = 0
	}
	if newIndex > len(newItems) {
		newIndex = len(newItems)
	}

	// Insert
	finalItems := make([]playlistItem, 0, len(allItems))
	finalItems = append(finalItems, newItems[:newIndex]...)
	finalItems = append(finalItems, itemToMove)
	finalItems = append(finalItems, newItems[newIndex:]...)

	return finalItems, nil
}

func (c *ServeCmd) updateTrackNumbers(ctx context.Context, items []playlistItem) {
	for i, item := range items {
		newTrackNum := int64(i + 1)
		err := c.execDB(ctx, item.DBPath, func(ctx context.Context, sqlDB *sql.DB) error {
			queries := database.New(sqlDB)
			return queries.AddPlaylistItem(ctx, database.AddPlaylistItemParams{
				PlaylistID:  item.PlaylistID,
				MediaPath:   item.MediaPath,
				TrackNumber: sql.NullInt64{Int64: newTrackNum, Valid: true},
			})
		})
		if err != nil {
			models.Log.Error("Failed to update track number", "db", item.DBPath, "path", item.MediaPath, "error", err)
		}
	}
}

func (c *ServeCmd) HandlePlaylistReorder(w http.ResponseWriter, r *http.Request) {
	if c.ReadOnly {
		http.Error(w, "Read-only mode", http.StatusForbidden)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		PlaylistTitle string `json:"playlist_title"`
		MediaPath     string `json:"media_path"`
		NewIndex      int    `json:"new_index"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	allItems, err := c.gatherPlaylistItems(r.Context(), req.PlaylistTitle)
	if err != nil {
		models.Log.Error("Failed to gather playlist items", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if len(allItems) == 0 {
		http.Error(w, "Playlist not found or empty", http.StatusNotFound)
		return
	}

	finalItems, err := c.reorderItems(allItems, req.MediaPath, req.NewIndex)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	c.updateTrackNumbers(r.Context(), finalItems)
	w.WriteHeader(http.StatusOK)
}
