package query

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/chapmanjacobd/discotheque/internal/db"
	"github.com/chapmanjacobd/discotheque/internal/models"
	"github.com/chapmanjacobd/discotheque/internal/utils"
)

// QueryBuilder constructs SQL queries from flags
type QueryBuilder struct {
	Flags models.GlobalFlags
}

func NewQueryBuilder(flags models.GlobalFlags) *QueryBuilder {
	return &QueryBuilder{Flags: flags}
}

func (qb *QueryBuilder) Build() (string, []any) {
	// If raw query provided, use it
	if qb.Flags.Query != "" {
		return qb.Flags.Query, nil
	}

	var whereClauses []string
	var args []any

	// Base table
	table := "media"
	if qb.Flags.FTS {
		table = qb.Flags.FTSTable
	}

	// Deleted status
	if qb.Flags.OnlyDeleted {
		whereClauses = append(whereClauses, "COALESCE(time_deleted, 0) > 0")
	} else if qb.Flags.HideDeleted {
		whereClauses = append(whereClauses, "COALESCE(time_deleted, 0) = 0")
	}

	if qb.Flags.DeletedAfter != "" {
		if ts := utils.ParseDateOrRelative(qb.Flags.DeletedAfter); ts > 0 {
			whereClauses = append(whereClauses, "time_deleted >= ?")
			args = append(args, ts)
		}
	}
	if qb.Flags.DeletedBefore != "" {
		if ts := utils.ParseDateOrRelative(qb.Flags.DeletedBefore); ts > 0 {
			whereClauses = append(whereClauses, "time_deleted <= ?")
			args = append(args, ts)
		}
	}

	// Search terms (FTS or LIKE)
	allInclude := append([]string{}, qb.Flags.Search...)
	allInclude = append(allInclude, qb.Flags.Include...)

	if len(allInclude) > 0 {
		if qb.Flags.FTS {
			// FTS match syntax
			searchTerm := strings.Join(allInclude, " ")
			whereClauses = append(whereClauses, fmt.Sprintf("%s MATCH ?", table))
			args = append(args, searchTerm)
		} else {
			// Regular LIKE search
			for _, term := range allInclude {
				whereClauses = append(whereClauses, "(path LIKE ? OR title LIKE ?)")
				pattern := "%" + term + "%"
				args = append(args, pattern, pattern)
			}
		}
	}

	for _, exc := range qb.Flags.Exclude {
		whereClauses = append(whereClauses, "path NOT LIKE ? AND title NOT LIKE ?")
		pattern := "%" + exc + "%"
		args = append(args, pattern, pattern)
	}

	// Regex filter (requires regex extension or post-filter)
	if qb.Flags.Regex != "" {
		whereClauses = append(whereClauses, "path REGEXP ?")
		args = append(args, qb.Flags.Regex)
	}

	// Path contains filters
	for _, contain := range qb.Flags.PathContains {
		whereClauses = append(whereClauses, "path LIKE ?")
		args = append(args, "%"+contain+"%")
	}

	// Size filters
	for _, s := range qb.Flags.Size {
		if r, err := utils.ParseRange(s, utils.HumanToBytes); err == nil {
			if r.Value != nil {
				whereClauses = append(whereClauses, "size = ?")
				args = append(args, *r.Value)
			}
			if r.Min != nil {
				whereClauses = append(whereClauses, "size >= ?")
				args = append(args, *r.Min)
			}
			if r.Max != nil {
				whereClauses = append(whereClauses, "size <= ?")
				args = append(args, *r.Max)
			}
		}
	}
	if qb.Flags.MinSize != "" {
		if minBytes, err := utils.HumanToBytes(qb.Flags.MinSize); err == nil {
			whereClauses = append(whereClauses, "size >= ?")
			args = append(args, minBytes)
		}
	}
	if qb.Flags.MaxSize != "" {
		if maxBytes, err := utils.HumanToBytes(qb.Flags.MaxSize); err == nil {
			whereClauses = append(whereClauses, "size <= ?")
			args = append(args, maxBytes)
		}
	}

	// Duration filters
	for _, s := range qb.Flags.Duration {
		if r, err := utils.ParseRange(s, utils.HumanToSeconds); err == nil {
			if r.Value != nil {
				whereClauses = append(whereClauses, "duration = ?")
				args = append(args, *r.Value)
			}
			if r.Min != nil {
				whereClauses = append(whereClauses, "duration >= ?")
				args = append(args, *r.Min)
			}
			if r.Max != nil {
				whereClauses = append(whereClauses, "duration <= ?")
				args = append(args, *r.Max)
			}
		}
	}
	if qb.Flags.MinDuration > 0 {
		whereClauses = append(whereClauses, "duration >= ?")
		args = append(args, int64(qb.Flags.MinDuration))
	}
	if qb.Flags.MaxDuration > 0 {
		whereClauses = append(whereClauses, "duration <= ?")
		args = append(args, int64(qb.Flags.MaxDuration))
	}

	// Time filters
	if qb.Flags.CreatedAfter != "" {
		if ts := utils.ParseDateOrRelative(qb.Flags.CreatedAfter); ts > 0 {
			whereClauses = append(whereClauses, "time_created >= ?")
			args = append(args, ts)
		}
	}
	if qb.Flags.CreatedBefore != "" {
		if ts := utils.ParseDateOrRelative(qb.Flags.CreatedBefore); ts > 0 {
			whereClauses = append(whereClauses, "time_created <= ?")
			args = append(args, ts)
		}
	}
	if qb.Flags.ModifiedAfter != "" {
		if ts := utils.ParseDateOrRelative(qb.Flags.ModifiedAfter); ts > 0 {
			whereClauses = append(whereClauses, "time_modified >= ?")
			args = append(args, ts)
		}
	}
	if qb.Flags.ModifiedBefore != "" {
		if ts := utils.ParseDateOrRelative(qb.Flags.ModifiedBefore); ts > 0 {
			whereClauses = append(whereClauses, "time_modified <= ?")
			args = append(args, ts)
		}
	}
	if qb.Flags.PlayedAfter != "" {
		if ts := utils.ParseDateOrRelative(qb.Flags.PlayedAfter); ts > 0 {
			whereClauses = append(whereClauses, "time_last_played >= ?")
			args = append(args, ts)
		}
	}
	if qb.Flags.PlayedBefore != "" {
		if ts := utils.ParseDateOrRelative(qb.Flags.PlayedBefore); ts > 0 {
			whereClauses = append(whereClauses, "time_last_played <= ?")
			args = append(args, ts)
		}
	}

	// Watched status
	if qb.Flags.Watched != nil {
		if *qb.Flags.Watched {
			whereClauses = append(whereClauses, "time_last_played > 0")
		} else {
			whereClauses = append(whereClauses, "COALESCE(time_last_played, 0) = 0")
		}
	}

	// Unfinished (has playhead but presumably not done)
	if qb.Flags.Unfinished || qb.Flags.InProgress {
		whereClauses = append(whereClauses, "playhead > 0 AND playhead < duration * 0.95")
	}

	if qb.Flags.Partial != "" {
		if strings.Contains(qb.Flags.Partial, "s") {
			whereClauses = append(whereClauses, "COALESCE(time_first_played, 0) = 0")
		} else {
			whereClauses = append(whereClauses, "time_first_played > 0")
		}
	}

	if qb.Flags.Completed {
		whereClauses = append(whereClauses, "playhead >= duration * 0.95")
	}

	// Play count filters
	if qb.Flags.PlayCountMin > 0 {
		whereClauses = append(whereClauses, "play_count >= ?")
		args = append(args, qb.Flags.PlayCountMin)
	}
	if qb.Flags.PlayCountMax > 0 {
		whereClauses = append(whereClauses, "play_count <= ?")
		args = append(args, qb.Flags.PlayCountMax)
	}

	// Content type filters
	if qb.Flags.VideoOnly {
		whereClauses = append(whereClauses, "(path LIKE '%.mp4' OR path LIKE '%.mkv' OR path LIKE '%.avi' OR path LIKE '%.mov' OR path LIKE '%.webm')")
	}
	if qb.Flags.AudioOnly {
		whereClauses = append(whereClauses, "(path LIKE '%.mp3' OR path LIKE '%.flac' OR path LIKE '%.m4a' OR path LIKE '%.opus' OR path LIKE '%.ogg')")
	}
	if qb.Flags.ImageOnly {
		whereClauses = append(whereClauses, "(path LIKE '%.jpg' OR path LIKE '%.png' OR path LIKE '%.gif' OR path LIKE '%.webp')")
	}

	if qb.Flags.Portrait {
		whereClauses = append(whereClauses, "width < height")
	}

	if qb.Flags.OnlineMediaOnly {
		whereClauses = append(whereClauses, "path LIKE 'http%'")
	}
	if qb.Flags.LocalMediaOnly {
		whereClauses = append(whereClauses, "path NOT LIKE 'http%'")
	}

	// Custom WHERE clauses
	whereClauses = append(whereClauses, qb.Flags.Where...)

	if qb.Flags.DurationFromSize != "" {
		if r, err := utils.ParseRange(qb.Flags.DurationFromSize, utils.HumanToBytes); err == nil {
			var subWhere []string
			var subArgs []any
			if r.Value != nil {
				subWhere = append(subWhere, "size = ?")
				subArgs = append(subArgs, *r.Value)
			}
			if r.Min != nil {
				subWhere = append(subWhere, "size >= ?")
				subArgs = append(subArgs, *r.Min)
			}
			if r.Max != nil {
				subWhere = append(subWhere, "size <= ?")
				subArgs = append(subArgs, *r.Max)
			}

			if len(subWhere) > 0 {
				whereClauses = append(whereClauses, fmt.Sprintf("size IS NOT NULL AND duration IN (SELECT DISTINCT duration FROM media WHERE %s)", strings.Join(subWhere, " AND ")))
				args = append(args, subArgs...)
			}
		}
	}

	// Build query
	query := fmt.Sprintf("SELECT * FROM %s", table)

	if len(whereClauses) > 0 {
		query += " WHERE " + strings.Join(whereClauses, " AND ")
	}

	// Order by
	if !qb.Flags.Random && !qb.Flags.NatSort && qb.Flags.SortBy != "" {
		sortExpr := OverrideSort(qb.Flags.SortBy)
		order := "ASC"
		if qb.Flags.Reverse {
			order = "DESC"
		}
		query += fmt.Sprintf(" ORDER BY %s %s", sortExpr, order)
	} else if qb.Flags.Random {
		query += " ORDER BY RANDOM()"
	}

	// Limit and offset
	if qb.Flags.Limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", qb.Flags.Limit)
	}
	if qb.Flags.Offset > 0 {
		query += fmt.Sprintf(" OFFSET %d", qb.Flags.Offset)
	}

	return query, args
}

func OverrideSort(s string) string {
	yearMonthSQL := func(v string) string {
		return fmt.Sprintf("cast(strftime('%%Y%%m', datetime(%s, 'unixepoch')) as int)", v)
	}
	yearMonthDaySQL := func(v string) string {
		return fmt.Sprintf("cast(strftime('%%Y%%m%%d', datetime(%s, 'unixepoch')) as int)", v)
	}

	s = strings.ReplaceAll(s, "month_created", yearMonthSQL("time_created"))
	s = strings.ReplaceAll(s, "month_modified", yearMonthSQL("time_modified"))
	s = strings.ReplaceAll(s, "date_created", yearMonthDaySQL("time_created"))
	s = strings.ReplaceAll(s, "date_modified", yearMonthDaySQL("time_modified"))
	s = strings.ReplaceAll(s, "random()", "RANDOM()")
	s = strings.ReplaceAll(s, "random", "RANDOM()")
	s = strings.ReplaceAll(s, "priorityfast", "ntile(1000) over (order by size) desc, duration")
	s = strings.ReplaceAll(s, "priority", "ntile(1000) over (order by size/duration) desc")
	s = strings.ReplaceAll(s, "bitrate", "size/duration")

	return s
}

// MediaQuery executes a query against multiple databases concurrently
func MediaQuery(ctx context.Context, dbs []string, flags models.GlobalFlags) ([]models.MediaWithDB, error) {
	qb := NewQueryBuilder(flags)
	query, args := qb.Build()

	var wg sync.WaitGroup
	results := make(chan []models.MediaWithDB, len(dbs))
	errors := make(chan error, len(dbs))

	for _, dbPath := range dbs {
		wg.Add(1)
		go func(path string) {
			defer wg.Done()
			media, err := QueryDatabase(ctx, path, query, args)
			if err != nil {
				errors <- fmt.Errorf("%s: %w", path, err)
				return
			}
			results <- media
		}(dbPath)
	}

	go func() {
		wg.Wait()
		close(results)
		close(errors)
	}()

	var allMedia []models.MediaWithDB
	for media := range results {
		allMedia = append(allMedia, media...)
	}

	var errs []error
	for err := range errors {
		errs = append(errs, err)
	}
	if len(errs) > 0 {
		return allMedia, fmt.Errorf("query errors: %v", errs)
	}

	if flags.FileCounts != "" {
		allMedia = FilterEpisodic(allMedia, flags.FileCounts)
	}

	if flags.FetchSiblings != "" {
		var err error
		allMedia, err = FetchSiblings(ctx, allMedia, flags)
		if err != nil {
			return allMedia, err
		}
	}

	return allMedia, nil
}

func FetchSiblings(ctx context.Context, media []models.MediaWithDB, flags models.GlobalFlags) ([]models.MediaWithDB, error) {
	if len(media) == 0 {
		return media, nil
	}

	parentToFiles := make(map[string][]models.MediaWithDB)
	for _, m := range media {
		dir := m.Parent() + "/"
		parentToFiles[dir] = append(parentToFiles[dir], m)
	}

	var allSiblings []models.MediaWithDB
	seenPaths := make(map[string]bool)

	for dir, filesInDir := range parentToFiles {
		dbPath := filesInDir[0].DB

		limit := flags.FetchSiblingsMax
		if flags.FetchSiblings == "all" || flags.FetchSiblings == "always" {
			limit = 2000
		} else if flags.FetchSiblings == "each" {
			if limit <= 0 {
				limit = len(filesInDir)
			}
		} else if flags.FetchSiblings == "if-audiobook" {
			isAudiobook := false
			for _, f := range filesInDir {
				if strings.Contains(strings.ToLower(f.Path), "audiobook") {
					isAudiobook = true
					break
				}
			}
			if !isAudiobook {
				// Keep original files and move to next dir
				for _, f := range filesInDir {
					if !seenPaths[f.Path] {
						allSiblings = append(allSiblings, f)
						seenPaths[f.Path] = true
					}
				}
				continue
			}
			if limit <= 0 {
				limit = 2000 // default for audiobook siblings if not specified
			}
		} else if utils.IsDigit(flags.FetchSiblings) {
			if l, err := strconv.Atoi(flags.FetchSiblings); err == nil {
				limit = l
			}
		} else {
			// fallback: if not specified or unknown, just keep original
			for _, f := range filesInDir {
				if !seenPaths[f.Path] {
					allSiblings = append(allSiblings, f)
					seenPaths[f.Path] = true
				}
			}
			continue
		}

		// Fetch from DB
		query := "SELECT * FROM media WHERE time_deleted = 0 AND path LIKE ? ORDER BY path LIMIT ?"
		pattern := dir + "%"
		siblings, err := QueryDatabase(ctx, dbPath, query, []any{pattern, limit})
		if err != nil {
			return nil, err
		}

		for _, s := range siblings {
			if !seenPaths[s.Path] {
				allSiblings = append(allSiblings, s)
				seenPaths[s.Path] = true
			}
		}
	}

	return allSiblings, nil
}

// FilterEpisodic filters media based on the number of files in its directory
func FilterEpisodic(media []models.MediaWithDB, criteria string) []models.MediaWithDB {
	r, err := utils.ParseRange(criteria, func(s string) (int64, error) {
		return strconv.ParseInt(s, 10, 64)
	})
	if err != nil {
		return media
	}

	counts := make(map[string]int64)
	for _, m := range media {
		parent := m.Parent()
		counts[parent]++
	}

	var filtered []models.MediaWithDB
	for _, m := range media {
		count := counts[m.Parent()]
		if r.Matches(count) {
			filtered = append(filtered, m)
		}
	}
	return filtered
}

func QueryDatabase(ctx context.Context, dbPath, query string, args []any) ([]models.MediaWithDB, error) {
	sqlDB, err := db.Connect(dbPath)
	if err != nil {
		return nil, err
	}
	defer sqlDB.Close()

	rows, err := sqlDB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cols, _ := rows.Columns()
	var allMedia []models.MediaWithDB

	for rows.Next() {
		values := make([]any, len(cols))
		valuePtrs := make([]any, len(cols))
		for i := range cols {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, err
		}

		m := db.Media{}
		for i, col := range cols {
			if values[i] == nil {
				continue
			}

			switch strings.ToLower(col) {
			case "path":
				m.Path = utils.GetString(values[i])
			case "title":
				m.Title = sql.NullString{String: utils.GetString(values[i]), Valid: true}
			case "duration":
				m.Duration = sql.NullInt64{Int64: utils.GetInt64(values[i]), Valid: true}
			case "size":
				m.Size = sql.NullInt64{Int64: utils.GetInt64(values[i]), Valid: true}
			case "time_created":
				m.TimeCreated = sql.NullInt64{Int64: utils.GetInt64(values[i]), Valid: true}
			case "time_modified":
				m.TimeModified = sql.NullInt64{Int64: utils.GetInt64(values[i]), Valid: true}
			case "time_deleted":
				m.TimeDeleted = sql.NullInt64{Int64: utils.GetInt64(values[i]), Valid: true}
			case "time_first_played":
				m.TimeFirstPlayed = sql.NullInt64{Int64: utils.GetInt64(values[i]), Valid: true}
			case "time_last_played":
				m.TimeLastPlayed = sql.NullInt64{Int64: utils.GetInt64(values[i]), Valid: true}
			case "play_count":
				m.PlayCount = sql.NullInt64{Int64: utils.GetInt64(values[i]), Valid: true}
			case "playhead":
				m.Playhead = sql.NullInt64{Int64: utils.GetInt64(values[i]), Valid: true}
			case "album":
				m.Album = sql.NullString{String: utils.GetString(values[i]), Valid: true}
			case "artist":
				m.Artist = sql.NullString{String: utils.GetString(values[i]), Valid: true}
			case "genre":
				m.Genre = sql.NullString{String: utils.GetString(values[i]), Valid: true}
			case "mood":
				m.Mood = sql.NullString{String: utils.GetString(values[i]), Valid: true}
			case "bpm":
				m.Bpm = sql.NullInt64{Int64: utils.GetInt64(values[i]), Valid: true}
			case "key":
				m.Key = sql.NullString{String: utils.GetString(values[i]), Valid: true}
			case "decade":
				m.Decade = sql.NullString{String: utils.GetString(values[i]), Valid: true}
			case "categories":
				m.Categories = sql.NullString{String: utils.GetString(values[i]), Valid: true}
			case "city":
				m.City = sql.NullString{String: utils.GetString(values[i]), Valid: true}
			case "country":
				m.Country = sql.NullString{String: utils.GetString(values[i]), Valid: true}
			case "description":
				m.Description = sql.NullString{String: utils.GetString(values[i]), Valid: true}
			case "language":
				m.Language = sql.NullString{String: utils.GetString(values[i]), Valid: true}
			}
		}

		allMedia = append(allMedia, models.MediaWithDB{
			Media: models.FromDB(m),
			DB:    dbPath,
		})
	}

	return allMedia, rows.Err()
}

// FilterMedia applies all filters to media list
func FilterMedia(media []models.MediaWithDB, flags models.GlobalFlags) []models.MediaWithDB {
	var filtered []models.MediaWithDB

	for _, m := range media {
		// Check existence
		if flags.Exists && !utils.FileExists(m.Path) {
			continue
		}

		// Include/exclude patterns
		if len(flags.Include) > 0 && !utils.MatchesAny(m.Path, flags.Include) {
			continue
		}
		if len(flags.Exclude) > 0 && utils.MatchesAny(m.Path, flags.Exclude) {
			continue
		}

		// Size filters
		for _, s := range flags.Size {
			if r, err := utils.ParseRange(s, utils.HumanToBytes); err == nil {
				if m.Size == nil || !r.Matches(*m.Size) {
					continue
				}
			}
		}
		if flags.MinSize != "" {
			minBytes, _ := utils.HumanToBytes(flags.MinSize)
			if m.Size == nil || *m.Size < minBytes {
				continue
			}
		}
		if flags.MaxSize != "" {
			maxBytes, _ := utils.HumanToBytes(flags.MaxSize)
			if m.Size == nil || *m.Size > maxBytes {
				continue
			}
		}

		// Duration filters
		for _, s := range flags.Duration {
			if r, err := utils.ParseRange(s, utils.HumanToSeconds); err == nil {
				if m.Duration == nil || !r.Matches(*m.Duration) {
					continue
				}
			}
		}
		if flags.MinDuration > 0 && (m.Duration == nil || *m.Duration < int64(flags.MinDuration)) {
			continue
		}
		if flags.MaxDuration > 0 && (m.Duration == nil || *m.Duration > int64(flags.MaxDuration)) {
			continue
		}

		// Mimetype filters
		if len(flags.MimeType) > 0 {
			match := false
			if m.Type != nil && utils.IsMimeMatch(flags.MimeType, *m.Type) {
				match = true
			}
			if !match {
				continue
			}
		}
		if len(flags.NoMimeType) > 0 {
			if m.Type != nil && utils.IsMimeMatch(flags.NoMimeType, *m.Type) {
				continue
			}
		}

		filtered = append(filtered, m)
	}

	return filtered
}

// SortMedia sorts media using various methods
func SortMedia(media []models.MediaWithDB, sortBy string, reverse bool, natSort bool) {

	less := func(i, j int) bool {
		switch sortBy {
		case "path":
			if natSort {
				return utils.NaturalLess(media[i].Path, media[j].Path)
			}
			return media[i].Path < media[j].Path
		case "title":
			return utils.StringValue(media[i].Title) < utils.StringValue(media[j].Title)
		case "duration":
			return utils.Int64Value(media[i].Duration) < utils.Int64Value(media[j].Duration)
		case "size":
			return utils.Int64Value(media[i].Size) < utils.Int64Value(media[j].Size)
		case "time_created":
			return utils.Int64Value(media[i].TimeCreated) < utils.Int64Value(media[j].TimeCreated)
		case "time_modified":
			return utils.Int64Value(media[i].TimeModified) < utils.Int64Value(media[j].TimeModified)
		case "time_last_played":
			return utils.Int64Value(media[i].TimeLastPlayed) < utils.Int64Value(media[j].TimeLastPlayed)
		case "play_count":
			return utils.Int64Value(media[i].PlayCount) < utils.Int64Value(media[j].PlayCount)
		default:
			return media[i].Path < media[j].Path
		}
	}

	if reverse {
		sort.Slice(media, func(i, j int) bool { return !less(i, j) })
	} else {
		sort.Slice(media, less)
	}
}

// RegexSortMedia sorts media using the text processor (regex splitting and word sorting)
func RegexSortMedia(media []models.MediaWithDB, flags models.GlobalFlags) []models.MediaWithDB {
	if len(media) == 0 {
		return media
	}

	sentenceStrings := make([]string, len(media))
	mapping := make(map[string][]models.MediaWithDB)

	for i, m := range media {
		// Build a searchable sentence from path and title
		parts := []string{m.Path}
		if m.Title != nil {
			parts = append(parts, *m.Title)
		}
		sentence := utils.PathToSentence(strings.Join(parts, " "))
		sentenceStrings[i] = sentence
		mapping[sentence] = append(mapping[sentence], m)
	}

	sortedSentences := utils.TextProcessor(flags, sentenceStrings)
	
	// Reconstruct media list in sorted order
	result := make([]models.MediaWithDB, 0, len(media))
	seenCount := make(map[string]int)
	for _, s := range sortedSentences {
		idx := seenCount[s]
		if idx < len(mapping[s]) {
			result = append(result, mapping[s][idx])
			seenCount[s]++
		}
	}

	return result
}

// SortFolders sorts folder stats
func SortFolders(folders []models.FolderStats, sortBy string, reverse bool) {
	less := func(i, j int) bool {
		switch sortBy {
		case "count":
			return folders[i].Count < folders[j].Count
		case "size":
			return folders[i].TotalSize < folders[j].TotalSize
		case "duration":
			return folders[i].TotalDuration < folders[j].TotalDuration
		case "priority":
			p1 := float64(folders[i].TotalSize) / float64(utils.Max(1, folders[i].Count))
			p2 := float64(folders[j].TotalSize) / float64(utils.Max(1, folders[j].Count))
			if p1 != p2 {
				return p1 < p2
			}
			return folders[i].TotalSize < folders[j].TotalSize
		case "path":
			return folders[i].Path < folders[j].Path
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

func SummarizeMedia(media []models.MediaWithDB) []FrequencyStats {
	if len(media) == 0 {
		return nil
	}

	sizes := make([]int64, 0, len(media))
	durations := make([]int64, 0, len(media))

	for _, m := range media {
		if m.Size != nil {
			sizes = append(sizes, *m.Size)
		}
		if m.Duration != nil {
			durations = append(durations, *m.Duration)
		}
	}

	return []FrequencyStats{
		{
			Label:         "Total",
			Count:         len(media),
			TotalSize:     utils.SafeSum(sizes),
			TotalDuration: utils.SafeSum(durations),
		},
		{
			Label:         "Median",
			Count:         len(media),
			TotalSize:     int64(utils.SafeMedian(sizes)),
			TotalDuration: int64(utils.SafeMedian(durations)),
		},
	}
}

type FrequencyStats struct {
	Label         string `json:"label"`
	Count         int    `json:"count"`
	TotalSize     int64  `json:"total_size"`
	TotalDuration int64  `json:"total_duration"`
}

func HistoricalUsage(ctx context.Context, dbPath string, freq string, timeColumn string) ([]FrequencyStats, error) {
	sqlDB, err := db.Connect(dbPath)
	if err != nil {
		return nil, err
	}
	defer sqlDB.Close()

	var freqSql string
	switch freq {
	case "daily":
		freqSql = fmt.Sprintf("strftime('%%Y-%%m-%%d', datetime(%s, 'unixepoch'))", timeColumn)
	case "weekly":
		freqSql = fmt.Sprintf("strftime('%%Y-%%W', datetime(%s, 'unixepoch'))", timeColumn)
	case "monthly":
		freqSql = fmt.Sprintf("strftime('%%Y-%%m', datetime(%s, 'unixepoch'))", timeColumn)
	case "yearly":
		freqSql = fmt.Sprintf("strftime('%%Y', datetime(%s, 'unixepoch'))", timeColumn)
	default:
		return nil, fmt.Errorf("invalid frequency: %s", freq)
	}

	query := fmt.Sprintf(`
		SELECT
			%s AS label,
			COUNT(*) AS count,
			SUM(size) AS total_size,
			SUM(duration) AS total_duration
		FROM media
		WHERE %s > 0 AND time_deleted = 0
		GROUP BY label
		ORDER BY label DESC
	`, freqSql, timeColumn)

	rows, err := sqlDB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []FrequencyStats
	for rows.Next() {
		var s FrequencyStats
		var totalSize, totalDuration sql.NullInt64
		if err := rows.Scan(&s.Label, &s.Count, &totalSize, &totalDuration); err != nil {
			return nil, err
		}
		s.TotalSize = totalSize.Int64
		s.TotalDuration = totalDuration.Int64
		stats = append(stats, s)
	}
	return stats, nil
}
