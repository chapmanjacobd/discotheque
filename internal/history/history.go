package history

import (
	"context"
	"database/sql"
	"time"

	"github.com/chapmanjacobd/discotheque/internal/db"
)

type Tracker struct {
	queries *db.Queries
}

func NewTracker(database *sql.DB) *Tracker {
	return &Tracker{
		queries: db.New(database),
	}
}

func (t *Tracker) UpdatePlayback(ctx context.Context, path string, playhead int32) error {
	now := time.Now().Unix()
	return t.queries.UpdatePlayHistory(ctx, db.UpdatePlayHistoryParams{
		TimeLastPlayed:  now,
		TimeFirstPlayed: now,
		Playhead:        playhead,
		Path:            path,
	})
}

func (t *Tracker) MarkDeleted(ctx context.Context, path string) error {
	now := time.Now().Unix()
	return t.queries.MarkDeleted(ctx, db.MarkDeletedParams{
		TimeDeleted: now,
		Path:        path,
	})
}
