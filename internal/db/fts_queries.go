package db

import (
	"context"
	"database/sql"
)

// SearchMediaFTSParams are parameters for FTS media search
type SearchMediaFTSParams struct {
	Query string
	Limit int64
}

// SearchMediaFTSResult is a Media item with FTS rank
type SearchMediaFTSResult struct {
	Media Media
	Rank  float64
}

// SearchMediaFTS searches media using FTS5 with trigram-compatible queries
func (q *Queries) SearchMediaFTS(ctx context.Context, arg SearchMediaFTSParams) ([]SearchMediaFTSResult, error) {
	// Use trigram-compatible query (3-char terms for detail=none) with ranking
	query := `
SELECT m.*, media_fts.rank FROM media m, media_fts
WHERE m.rowid = media_fts.rowid
AND media_fts MATCH ?
AND m.time_deleted = 0
ORDER BY media_fts.rank DESC
LIMIT ?
`
	rows, err := q.db.QueryContext(ctx, query, arg.Query, arg.Limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []SearchMediaFTSResult
	for rows.Next() {
		var m Media
		var rank float64
		err := rows.Scan(
			&m.Path,
			&m.FtsPath,
			&m.Title,
			&m.Duration,
			&m.Size,
			&m.TimeCreated,
			&m.TimeModified,
			&m.TimeDeleted,
			&m.TimeFirstPlayed,
			&m.TimeLastPlayed,
			&m.PlayCount,
			&m.Playhead,
			&m.Type,
			&m.Width,
			&m.Height,
			&m.Fps,
			&m.VideoCodecs,
			&m.AudioCodecs,
			&m.SubtitleCodecs,
			&m.VideoCount,
			&m.AudioCount,
			&m.SubtitleCount,
			&m.Album,
			&m.Artist,
			&m.Genre,
			&m.Categories,
			&m.Description,
			&m.Language,
			&m.TimeDownloaded,
			&m.Score,
			&rank,
		)
		if err != nil {
			return nil, err
		}
		results = append(results, SearchMediaFTSResult{Media: m, Rank: rank})
	}

	return results, rows.Err()
}

// SearchCaptionsParams are parameters for caption search
type SearchCaptionsParams struct {
	Query     string
	VideoOnly bool
	AudioOnly bool
	ImageOnly bool
	TextOnly  bool
	Limit     int64
}

// SearchCaptionsRow represents a row from caption search with FTS rank
type SearchCaptionsRow struct {
	MediaPath string
	Time      sql.NullFloat64
	Text      sql.NullString
	Title     sql.NullString
	Type      sql.NullString
	Size      sql.NullInt64
	Duration  sql.NullInt64
	Rank      float64
}

// SearchCaptions searches captions using FTS5 with trigram-compatible queries
func (q *Queries) SearchCaptions(ctx context.Context, arg SearchCaptionsParams) ([]SearchCaptionsRow, error) {
	query := `
SELECT c.media_path, c.time, c.text, m.title, m.type, m.size, m.duration, captions_fts.rank
FROM captions c, captions_fts, media m
WHERE c.rowid = captions_fts.rowid
AND c.media_path = m.path
AND captions_fts MATCH ?
AND m.time_deleted = 0
AND c.text IS NOT NULL AND c.text != ''
AND (? = 0 OR m.type = 'video')
AND (? = 0 OR m.type IN ('audio', 'audiobook'))
AND (? = 0 OR m.type = 'image')
AND (? = 0 OR m.type = 'text')
ORDER BY captions_fts.rank DESC, c.media_path, c.time
LIMIT ?
`
	videoOnly := 0
	audioOnly := 0
	imageOnly := 0
	textOnly := 0
	if arg.VideoOnly {
		videoOnly = 1
	}
	if arg.AudioOnly {
		audioOnly = 1
	}
	if arg.ImageOnly {
		imageOnly = 1
	}
	if arg.TextOnly {
		textOnly = 1
	}

	rows, err := q.db.QueryContext(ctx, query, arg.Query, videoOnly, audioOnly, imageOnly, textOnly, arg.Limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []SearchCaptionsRow
	for rows.Next() {
		var r SearchCaptionsRow
		err := rows.Scan(&r.MediaPath, &r.Time, &r.Text, &r.Title, &r.Type, &r.Size, &r.Duration, &r.Rank)
		if err != nil {
			return nil, err
		}
		results = append(results, r)
	}

	return results, rows.Err()
}
