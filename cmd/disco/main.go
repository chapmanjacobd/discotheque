package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/alecthomas/kong"
	"github.com/chapmanjacobd/discotheque/internal/db"
	"github.com/chapmanjacobd/discotheque/internal/fs"
	"github.com/chapmanjacobd/discotheque/internal/metadata"
	"github.com/chapmanjacobd/discotheque/internal/models"
	_ "github.com/mattn/go-sqlite3"
)

// CLI defines the command-line interface
type CLI struct {
	Add     AddCmd     `cmd:"" help:"Add media to database"`
	Print   PrintCmd   `cmd:"" help:"Print media information"`
	Watch   WatchCmd   `cmd:"" help:"Watch videos with mpv"`
	Listen  ListenCmd  `cmd:"" help:"Listen to audio with mpv"`
	Stats   StatsCmd   `cmd:"" help:"Show library statistics"`
	History HistoryCmd `cmd:"" help:"Show playback history"`
	Open    OpenCmd    `cmd:"" help:"Open files with default application"`
	Browse  BrowseCmd  `cmd:"" help:"Open URLs in browser"`
}

// GlobalFlags are flags available to all commands
type GlobalFlags struct {
	Query  string `short:"q" help:"Raw SQL query (overrides all query building)"`
	Limit  int    `short:"L" default:"100" help:"Limit results per database"`
	Offset int    `help:"Skip N results"`

	// Path filters
	Include      []string `short:"i" help:"Include paths matching pattern"`
	Exclude      []string `short:"e" help:"Exclude paths matching pattern"`
	Search       []string `help:"Search terms (space-separated for AND, | for OR)"`
	Regex        string   `help:"Filter paths by regex pattern"`
	PathContains []string `help:"Path must contain all these strings"`

	// Size/Duration filters
	MinSize     string   `help:"Minimum file size (e.g., 100MB)"`
	MaxSize     string   `help:"Maximum file size"`
	MinDuration int      `help:"Minimum duration in seconds"`
	MaxDuration int      `help:"Maximum duration in seconds"`
	Ext         []string `help:"Filter by extensions (e.g., .mp4,.mkv)"`

	// Time filters
	CreatedAfter  string `help:"Created after date (YYYY-MM-DD)"`
	CreatedBefore string `help:"Created before date (YYYY-MM-DD)"`
	ModifiedAfter string `help:"Modified after date (YYYY-MM-DD)"`
	PlayedAfter   string `help:"Last played after date (YYYY-MM-DD)"`
	PlayedBefore  string `help:"Last played before date (YYYY-MM-DD)"`

	// Playback state filters
	Watched      *bool `help:"Filter by watched status (true/false)"`
	Unfinished   bool  `help:"Has playhead but not finished"`
	PlayCountMin int   `help:"Minimum play count"`
	PlayCountMax int   `help:"Maximum play count"`

	// Content type filters
	VideoOnly bool `help:"Only video files"`
	AudioOnly bool `help:"Only audio files"`
	ImageOnly bool `help:"Only image files"`

	// Sorting
	SortBy  string `short:"s" default:"path" help:"Sort by field"`
	Reverse bool   `short:"r" help:"Reverse sort order"`
	NatSort bool   `short:"n" help:"Use natural sorting"`
	Random  bool   `help:"Random order"`

	// Display
	Columns []string `short:"c" help:"Columns to display"`
	BigDirs bool     `help:"Aggregate by parent directory"`
	JSON    bool     `short:"j" help:"Output results as JSON"`

	// Actions
	PostAction   string `help:"Post-action: none, delete, mark-deleted, move, copy"`
	DeleteFiles  bool   `help:"Delete files after action"`
	MarkDeleted  bool   `help:"Mark as deleted in database"`
	MoveTo       string `help:"Move files to directory"`
	CopyTo       string `help:"Copy files to directory"`
	Exists       bool   `help:"Filter out non-existent files"`
	TrackHistory bool   `default:"true" help:"Track playback history"`

	// FTS options
	FTS      bool   `help:"Use full-text search if available"`
	FTSTable string `default:"media_fts" help:"FTS table name"`
}

func (g *GlobalFlags) AfterApply() error {
	if g.Ext != nil {
		for i, ext := range g.Ext {
			if !strings.HasPrefix(ext, ".") {
				g.Ext[i] = "." + ext
			}
		}
	}
	return nil
}

type PrintCmd struct {
	GlobalFlags
	Databases []string `arg:"" required:"" help:"SQLite database files" type:"existingfile"`
}

type WatchCmd struct {
	GlobalFlags
	Databases    []string `arg:"" required:"" help:"SQLite database files" type:"existingfile"`
	Volume       int      `help:"Set volume (0-100)"`
	Fullscreen   bool     `short:"f" help:"Start in fullscreen"`
	NoSubtitles  bool     `help:"Disable subtitles"`
	Speed        float64  `default:"1.0" help:"Playback speed"`
	Start        string   `help:"Start time (e.g., 5:30 or 30%)"`
	SavePlayhead bool     `default:"true" help:"Save playback position"`
}

type ListenCmd struct {
	GlobalFlags
	Databases []string `arg:"" required:"" help:"SQLite database files" type:"existingfile"`
	Volume    int      `help:"Set volume (0-100)"`
	Speed     float64  `default:"1.0" help:"Playback speed"`
}

type OpenCmd struct {
	GlobalFlags
	Databases []string `arg:"" required:"" help:"SQLite database files" type:"existingfile"`
}

type BrowseCmd struct {
	GlobalFlags
	Databases []string `arg:"" required:"" help:"SQLite database files" type:"existingfile"`
	Browser   string   `help:"Browser to use"`
}

type AddCmd struct {
	GlobalFlags
	Args     []string `arg:"" name:"args" required:"" help:"Paths to scan followed by the database file"`
	Parallel int      `short:"p" default:"4" help:"Number of parallel extractors"`

	ScanPaths []string `kong:"-"`
	Database  string   `kong:"-"`
}

func (c *AddCmd) AfterApply() error {
	if len(c.Args) < 2 {
		return fmt.Errorf("at least one path to scan and one database file are required")
	}
	c.ScanPaths = c.Args[:len(c.Args)-1]
	c.Database = c.Args[len(c.Args)-1]
	return nil
}

func (c *AddCmd) Run(ctx *kong.Context) error {
	dbPath := c.Database

	sqlDB, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return err
	}
	defer sqlDB.Close()

	queries := db.New(sqlDB)

	for _, root := range c.ScanPaths {
		fmt.Printf("Scanning %s...\n", root)
		files, err := fs.FindMedia(root)
		if err != nil {
			return err
		}

		// Filter by Ext if specified
		if len(c.Ext) > 0 {
			var filtered []string
			extMap := make(map[string]bool)
			for _, e := range c.Ext {
				extMap[strings.ToLower(e)] = true
			}
			for _, f := range files {
				if extMap[strings.ToLower(filepath.Ext(f))] {
					filtered = append(filtered, f)
				}
			}
			files = filtered
		}

		fmt.Printf("Found %d media files. Extracting metadata...\n", len(files))

		// Parallel extraction
		jobs := make(chan string, len(files))
		results := make(chan *db.UpsertMediaParams, len(files))
		var wg sync.WaitGroup

		for i := 0; i < c.Parallel; i++ {
			wg.Go(func() {
				for path := range jobs {
					params, err := metadata.Extract(context.Background(), path)
					if err != nil {
						fmt.Fprintf(os.Stderr, "Error extracting %s: %v\n", path, err)
						continue
					}
					results <- params
				}
			})
		}

		go func() {
			for _, f := range files {
				jobs <- f
			}
			close(jobs)
		}()

		go func() {
			wg.Wait()
			close(results)
		}()

		count := 0
		for params := range results {
			if err := queries.UpsertMedia(context.Background(), *params); err != nil {
				fmt.Fprintf(os.Stderr, "Error saving %s: %v\n", params.Path, err)
			}
			count++
			if count%10 == 0 || count == len(files) {
				fmt.Printf("\rProcessed %d/%d", count, len(files))
			}
		}
		fmt.Println()
	}

	return nil
}

// FolderStats aggregates media by folder
type FolderStats struct {
	Path          string
	Count         int
	TotalSize     int64
	TotalDuration int
	AvgSize       int64
	AvgDuration   int
	Files         []MediaWithDB
}

type StatsCmd struct {
	Databases []string `arg:"" required:"" help:"SQLite database files" type:"existingfile"`
	JSON      bool     `short:"j" help:"Output as JSON"`
}

func (c *StatsCmd) Run(ctx *kong.Context) error {
	for _, dbPath := range c.Databases {
		sqlDB, err := sql.Open("sqlite3", dbPath)
		if err != nil {
			return err
		}
		defer sqlDB.Close()

		queries := db.New(sqlDB)
		stats, err := queries.GetStats(context.Background())
		if err != nil {
			return err
		}

		typeStats, err := queries.GetStatsByType(context.Background())
		if err != nil {
			return err
		}

		if c.JSON {
			result := map[string]any{
				"database":  dbPath,
				"summary":   stats,
				"breakdown": typeStats,
			}
			encoder := json.NewEncoder(os.Stdout)
			encoder.SetIndent("", "  ")
			if err := encoder.Encode(result); err != nil {
				return err
			}
			continue
		}

		fmt.Printf("Statistics for %s:\n", dbPath)
		fmt.Printf("  Total Files:      %d\n", stats.TotalCount)
		fmt.Printf("  Total Size:       %s\n", formatSize(getInt64(stats.TotalSize)))
		fmt.Printf("  Total Duration:   %s\n", formatDuration(int(getInt64(stats.TotalDuration))))
		fmt.Printf("  Watched Files:    %d\n", stats.WatchedCount)
		fmt.Printf("  Unwatched Files:  %d\n", stats.UnwatchedCount)

		if len(typeStats) > 0 {
			fmt.Println("\n  Breakdown by Type:")
			for _, ts := range typeStats {
				t := "unknown"
				if ts.Type.Valid {
					t = ts.Type.String
				}
				fmt.Printf("    %-10s: %d files, %s, %s\n",
					t, ts.Count,
					formatSize(getInt64(ts.TotalSize)),
					formatDuration(int(getInt64(ts.TotalDuration))))
			}
		}
		fmt.Println()
	}
	return nil
}

type HistoryCmd struct {
	GlobalFlags
	Databases []string `arg:"" required:"" help:"SQLite database files" type:"existingfile"`
}

func (c *HistoryCmd) Run(ctx *kong.Context) error {
	// Set default sort for history
	if c.SortBy == "path" || c.SortBy == "" {
		c.SortBy = "time_last_played"
		c.Reverse = true
	}

	media, err := MediaQuery(context.Background(), c.Databases, c.GlobalFlags)
	if err != nil {
		return err
	}

	// Filter for only watched items if not otherwise specified
	if c.Watched == nil {
		watched := true
		c.Watched = &watched
	}

	media = FilterMedia(media, c.GlobalFlags)

	if c.JSON {
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(media)
	}

	SortMedia(media, c.SortBy, c.Reverse, c.NatSort)
	return printMedia(c.Columns, media)
}

type MediaWithDB struct {
	models.Media
	DB string
}

// QueryBuilder constructs SQL queries from flags
type QueryBuilder struct {
	Flags GlobalFlags
}

func NewQueryBuilder(flags GlobalFlags) *QueryBuilder {
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

	// Always exclude deleted unless explicitly included
	whereClauses = append(whereClauses, "COALESCE(time_deleted, 0) = 0")

	// Search terms (FTS or LIKE)
	if len(qb.Flags.Search) > 0 {
		if qb.Flags.FTS {
			// FTS match syntax
			searchTerm := strings.Join(qb.Flags.Search, " ")
			whereClauses = append(whereClauses, fmt.Sprintf("%s MATCH ?", table))
			args = append(args, searchTerm)
		} else {
			// Regular LIKE search
			for _, term := range qb.Flags.Search {
				whereClauses = append(whereClauses, "(path LIKE ? OR title LIKE ?)")
				pattern := "%" + term + "%"
				args = append(args, pattern, pattern)
			}
		}
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
	if qb.Flags.MinSize != "" {
		if minBytes, err := humanToBytes(qb.Flags.MinSize); err == nil {
			whereClauses = append(whereClauses, "size >= ?")
			args = append(args, minBytes)
		}
	}
	if qb.Flags.MaxSize != "" {
		if maxBytes, err := humanToBytes(qb.Flags.MaxSize); err == nil {
			whereClauses = append(whereClauses, "size <= ?")
			args = append(args, maxBytes)
		}
	}

	// Duration filters
	if qb.Flags.MinDuration > 0 {
		whereClauses = append(whereClauses, "duration >= ?")
		args = append(args, qb.Flags.MinDuration)
	}
	if qb.Flags.MaxDuration > 0 {
		whereClauses = append(whereClauses, "duration <= ?")
		args = append(args, qb.Flags.MaxDuration)
	}

	// Time filters
	if qb.Flags.CreatedAfter != "" {
		if ts := parseDate(qb.Flags.CreatedAfter); ts > 0 {
			whereClauses = append(whereClauses, "time_created >= ?")
			args = append(args, ts)
		}
	}
	if qb.Flags.CreatedBefore != "" {
		if ts := parseDate(qb.Flags.CreatedBefore); ts > 0 {
			whereClauses = append(whereClauses, "time_created <= ?")
			args = append(args, ts)
		}
	}
	if qb.Flags.ModifiedAfter != "" {
		if ts := parseDate(qb.Flags.ModifiedAfter); ts > 0 {
			whereClauses = append(whereClauses, "time_modified >= ?")
			args = append(args, ts)
		}
	}
	if qb.Flags.PlayedAfter != "" {
		if ts := parseDate(qb.Flags.PlayedAfter); ts > 0 {
			whereClauses = append(whereClauses, "time_last_played >= ?")
			args = append(args, ts)
		}
	}
	if qb.Flags.PlayedBefore != "" {
		if ts := parseDate(qb.Flags.PlayedBefore); ts > 0 {
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
	if qb.Flags.Unfinished {
		whereClauses = append(whereClauses, "playhead > 0 AND playhead < duration * 0.95")
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

	// Content type filters (simplified - would need better detection in real use)
	if qb.Flags.VideoOnly {
		whereClauses = append(whereClauses, "(path LIKE '%.mp4' OR path LIKE '%.mkv' OR path LIKE '%.avi' OR path LIKE '%.mov' OR path LIKE '%.webm')")
	}
	if qb.Flags.AudioOnly {
		whereClauses = append(whereClauses, "(path LIKE '%.mp3' OR path LIKE '%.flac' OR path LIKE '%.m4a' OR path LIKE '%.opus' OR path LIKE '%.ogg')")
	}
	if qb.Flags.ImageOnly {
		whereClauses = append(whereClauses, "(path LIKE '%.jpg' OR path LIKE '%.png' OR path LIKE '%.gif' OR path LIKE '%.webp')")
	}

	// Build query
	query := fmt.Sprintf("SELECT * FROM %s", table)

	if len(whereClauses) > 0 {
		query += " WHERE " + strings.Join(whereClauses, " AND ")
	}

	// Order by (if not random and not using in-memory sort)
	if !qb.Flags.Random && !qb.Flags.NatSort && qb.Flags.SortBy != "" {
		order := "ASC"
		if qb.Flags.Reverse {
			order = "DESC"
		}
		query += fmt.Sprintf(" ORDER BY %s %s", qb.Flags.SortBy, order)
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

// MediaQuery executes a query against multiple databases concurrently
func MediaQuery(ctx context.Context, dbs []string, flags GlobalFlags) ([]MediaWithDB, error) {
	qb := NewQueryBuilder(flags)
	query, args := qb.Build()

	var wg sync.WaitGroup
	results := make(chan []MediaWithDB, len(dbs))
	errors := make(chan error, len(dbs))

	for _, dbPath := range dbs {
		wg.Add(1)
		go func(path string) {
			defer wg.Done()
			media, err := queryDatabase(ctx, path, query, args)
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

	var allMedia []MediaWithDB
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

	return allMedia, nil
}

func queryDatabase(ctx context.Context, dbPath, query string, args []any) ([]MediaWithDB, error) {
	sqlDB, err := sql.Open("sqlite3", dbPath)
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
	var allMedia []MediaWithDB

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
				m.Path = getString(values[i])
			case "title":
				m.Title = sql.NullString{String: getString(values[i]), Valid: true}
			case "duration":
				m.Duration = sql.NullInt64{Int64: getInt64(values[i]), Valid: true}
			case "size":
				m.Size = sql.NullInt64{Int64: getInt64(values[i]), Valid: true}
			case "time_created":
				m.TimeCreated = sql.NullInt64{Int64: getInt64(values[i]), Valid: true}
			case "time_modified":
				m.TimeModified = sql.NullInt64{Int64: getInt64(values[i]), Valid: true}
			case "time_deleted":
				m.TimeDeleted = sql.NullInt64{Int64: getInt64(values[i]), Valid: true}
			case "time_first_played":
				m.TimeFirstPlayed = sql.NullInt64{Int64: getInt64(values[i]), Valid: true}
			case "time_last_played":
				m.TimeLastPlayed = sql.NullInt64{Int64: getInt64(values[i]), Valid: true}
			case "play_count":
				m.PlayCount = sql.NullInt64{Int64: getInt64(values[i]), Valid: true}
			case "playhead":
				m.Playhead = sql.NullInt64{Int64: getInt64(values[i]), Valid: true}
			}
		}

		allMedia = append(allMedia, MediaWithDB{
			Media: models.FromDB(m),
			DB:    dbPath,
		})
	}

	return allMedia, rows.Err()
}

func parseDate(dateStr string) int64 {
	layouts := []string{
		"2006-01-02",
		"2006-01-02 15:04:05",
		"2006-01-02 15:04",
		"01/02/2006",
	}

	for _, layout := range layouts {
		if t, err := time.Parse(layout, dateStr); err == nil {
			return t.Unix()
		}
	}
	return 0
}

// FilterMedia applies all filters to media list
func FilterMedia(media []MediaWithDB, flags GlobalFlags) []MediaWithDB {
	var filtered []MediaWithDB

	for _, m := range media {
		// Check existence
		if flags.Exists && !fileExists(m.Path) {
			continue
		}

		// Include/exclude patterns
		if len(flags.Include) > 0 && !matchesAny(m.Path, flags.Include) {
			continue
		}
		if len(flags.Exclude) > 0 && matchesAny(m.Path, flags.Exclude) {
			continue
		}

		// Size filters
		if flags.MinSize != "" {
			minBytes, _ := humanToBytes(flags.MinSize)
			if m.Size == nil || *m.Size < minBytes {
				continue
			}
		}
		if flags.MaxSize != "" {
			maxBytes, _ := humanToBytes(flags.MaxSize)
			if m.Size == nil || *m.Size > maxBytes {
				continue
			}
		}

		// Duration filters
		if flags.MinDuration > 0 && (m.Duration == nil || *m.Duration < int64(flags.MinDuration)) {
			continue
		}
		if flags.MaxDuration > 0 && (m.Duration == nil || *m.Duration > int64(flags.MaxDuration)) {
			continue
		}

		filtered = append(filtered, m)
	}

	return filtered
}

// AggregateFolders groups media by parent directory
func AggregateFolders(media []MediaWithDB) []FolderStats {
	folders := make(map[string]*FolderStats)

	for _, m := range media {
		parent := filepath.Dir(m.Path)
		if _, exists := folders[parent]; !exists {
			folders[parent] = &FolderStats{
				Path:  parent,
				Files: []MediaWithDB{},
			}
		}

		f := folders[parent]
		f.Count++
		if m.Size != nil {
			f.TotalSize += *m.Size
		}
		if m.Duration != nil {
			f.TotalDuration += int(*m.Duration)
		}
		f.Files = append(f.Files, m)
	}

	var result []FolderStats
	for _, f := range folders {
		if f.Count > 0 {
			f.AvgSize = f.TotalSize / int64(f.Count)
			f.AvgDuration = f.TotalDuration / f.Count
		}
		result = append(result, *f)
	}

	return result
}

// SortMedia sorts media using various methods
func SortMedia(media []MediaWithDB, sortBy string, reverse bool, natSort bool) {
	less := func(i, j int) bool {
		switch sortBy {
		case "path":
			if natSort {
				return naturalLess(media[i].Path, media[j].Path)
			}
			return media[i].Path < media[j].Path
		case "title":
			return stringValue(media[i].Title) < stringValue(media[j].Title)
		case "duration":
			return int64Value(media[i].Duration) < int64Value(media[j].Duration)
		case "size":
			return int64Value(media[i].Size) < int64Value(media[j].Size)
		case "time_created":
			return int64Value(media[i].TimeCreated) < int64Value(media[j].TimeCreated)
		case "time_modified":
			return int64Value(media[i].TimeModified) < int64Value(media[j].TimeModified)
		case "time_last_played":
			return int64Value(media[i].TimeLastPlayed) < int64Value(media[j].TimeLastPlayed)
		case "play_count":
			return int64Value(media[i].PlayCount) < int64Value(media[j].PlayCount)
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

func stringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func int64Value(i *int64) int64 {
	if i == nil {
		return 0
	}
	return *i
}

// SortFolders sorts folder stats
func SortFolders(folders []FolderStats, sortBy string, reverse bool) {
	less := func(i, j int) bool {
		switch sortBy {
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

// Natural sorting implementation
func naturalLess(s1, s2 string) bool {
	n1, n2 := extractNumbers(s1), extractNumbers(s2)

	idx1, idx2 := 0, 0
	for idx1 < len(n1) && idx2 < len(n2) {
		if n1[idx1].isNum && n2[idx2].isNum {
			if n1[idx1].num != n2[idx2].num {
				return n1[idx1].num < n2[idx2].num
			}
		} else {
			if n1[idx1].str != n2[idx2].str {
				return n1[idx1].str < n2[idx2].str
			}
		}
		idx1++
		idx2++
	}

	return len(n1) < len(n2)
}

type chunk struct {
	str   string
	num   int
	isNum bool
}

func extractNumbers(s string) []chunk {
	re := regexp.MustCompile(`\d+|\D+`)
	matches := re.FindAllString(s, -1)

	var chunks []chunk
	for _, m := range matches {
		if num, err := strconv.Atoi(m); err == nil {
			chunks = append(chunks, chunk{num: num, isNum: true})
		} else {
			chunks = append(chunks, chunk{str: strings.ToLower(m), isNum: false})
		}
	}
	return chunks
}

// UpdateHistory updates playback history in database
func UpdateHistory(dbPath string, paths []string, playhead int) error {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return err
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	now := time.Now().Unix()

	for _, path := range paths {
		// Check if record exists
		var exists bool
		err := tx.QueryRow("SELECT 1 FROM media WHERE path = ?", path).Scan(&exists)

		if err == sql.ErrNoRows {
			continue
		} else if err != nil {
			return err
		}

		// Update history
		_, err = tx.Exec(`
			UPDATE media
			SET time_last_played = ?,
			    time_first_played = COALESCE(time_first_played, ?),
			    play_count = COALESCE(play_count, 0) + 1,
			    playhead = ?
			WHERE path = ?
		`, now, now, playhead, path)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// Commands implementation

func (c *PrintCmd) Run(ctx *kong.Context) error {
	media, err := MediaQuery(context.Background(), c.Databases, c.GlobalFlags)
	if err != nil {
		return err
	}

	media = FilterMedia(media, c.GlobalFlags)

	if c.JSON {
		if c.BigDirs {
			folders := AggregateFolders(media)
			SortFolders(folders, c.SortBy, c.Reverse)
			encoder := json.NewEncoder(os.Stdout)
			encoder.SetIndent("", "  ")
			return encoder.Encode(folders)
		}
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(media)
	}

	if c.BigDirs {
		folders := AggregateFolders(media)
		SortFolders(folders, c.SortBy, c.Reverse)
		return printFolders(c.Columns, folders)
	}

	SortMedia(media, c.SortBy, c.Reverse, c.NatSort)
	return printMedia(c.Columns, media)
}

func (c *WatchCmd) Run(ctx *kong.Context) error {
	media, err := MediaQuery(context.Background(), c.Databases, c.GlobalFlags)
	if err != nil {
		return err
	}

	media = FilterMedia(media, c.GlobalFlags)
	SortMedia(media, c.SortBy, c.Reverse, c.NatSort)

	if len(media) == 0 {
		return fmt.Errorf("no media found")
	}

	// Build mpv command
	args := []string{"mpv"}

	if c.Volume > 0 {
		args = append(args, fmt.Sprintf("--volume=%d", c.Volume))
	}
	if c.Fullscreen {
		args = append(args, "--fullscreen")
	}
	if c.NoSubtitles {
		args = append(args, "--no-sub")
	}
	if c.Speed != 1.0 {
		args = append(args, fmt.Sprintf("--speed=%.2f", c.Speed))
	}
	if c.Start != "" {
		args = append(args, fmt.Sprintf("--start=%s", c.Start))
	}
	if c.SavePlayhead {
		args = append(args, "--save-position-on-quit")
	}

	// Add file paths
	var paths []string
	for _, m := range media {
		if fileExists(m.Path) {
			paths = append(paths, m.Path)
		}
	}

	if len(paths) == 0 {
		return fmt.Errorf("no playable files found")
	}

	args = append(args, paths...)

	// Execute mpv
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		return err
	}

	// Update history
	if c.TrackHistory {
		for _, m := range media {
			if err := UpdateHistory(m.DB, []string{m.Path}, 0); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to update history: %v\n", err)
			}
		}
	}

	return executePostAction(c.GlobalFlags, media)
}

func (c *ListenCmd) Run(ctx *kong.Context) error {
	media, err := MediaQuery(context.Background(), c.Databases, c.GlobalFlags)
	if err != nil {
		return err
	}

	media = FilterMedia(media, c.GlobalFlags)
	SortMedia(media, c.SortBy, c.Reverse, c.NatSort)

	args := []string{"mpv", "--video=no"}

	if c.Volume > 0 {
		args = append(args, fmt.Sprintf("--volume=%d", c.Volume))
	}
	if c.Speed != 1.0 {
		args = append(args, fmt.Sprintf("--speed=%.2f", c.Speed))
	}

	for _, m := range media {
		if fileExists(m.Path) {
			args = append(args, m.Path)
		}
	}

	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return err
	}

	if c.TrackHistory {
		for _, m := range media {
			UpdateHistory(m.DB, []string{m.Path}, 0)
		}
	}

	return executePostAction(c.GlobalFlags, media)
}

func (c *OpenCmd) Run(ctx *kong.Context) error {
	media, err := MediaQuery(context.Background(), c.Databases, c.GlobalFlags)
	if err != nil {
		return err
	}

	media = FilterMedia(media, c.GlobalFlags)

	for _, m := range media {
		if !fileExists(m.Path) {
			continue
		}

		var cmd *exec.Cmd
		switch runtime.GOOS {
		case "linux":
			cmd = exec.Command("xdg-open", m.Path)
		case "darwin":
			cmd = exec.Command("open", m.Path)
		case "windows":
			cmd = exec.Command("cmd", "/c", "start", m.Path)
		}

		if err := cmd.Start(); err != nil {
			return err
		}
	}

	return executePostAction(c.GlobalFlags, media)
}

func (c *BrowseCmd) Run(ctx *kong.Context) error {
	media, err := MediaQuery(context.Background(), c.Databases, c.GlobalFlags)
	if err != nil {
		return err
	}

	media = FilterMedia(media, c.GlobalFlags)

	browser := c.Browser
	if browser == "" {
		browser = getDefaultBrowser()
	}

	var urls []string
	for _, m := range media {
		if strings.HasPrefix(m.Path, "http") {
			urls = append(urls, m.Path)
		}
	}

	if len(urls) == 0 {
		return fmt.Errorf("no URLs found")
	}

	args := append([]string{browser}, urls...)
	cmd := exec.Command(args[0], args[1:]...)
	return cmd.Start()
}

// Print functions

func printMedia(columns []string, media []MediaWithDB) error {
	if len(columns) == 0 {
		columns = []string{"path", "duration", "size"}
	}

	// Print header
	fmt.Println(strings.Join(columns, "\t"))

	for _, m := range media {
		var row []string
		for _, col := range columns {
			switch col {
			case "path":
				row = append(row, m.Path)
			case "title":
				row = append(row, stringValue(m.Title))
			case "duration":
				row = append(row, formatDuration(int(int64Value(m.Duration))))
			case "size":
				row = append(row, formatSize(int64Value(m.Size)))
			case "play_count":
				row = append(row, fmt.Sprintf("%d", int64Value(m.PlayCount)))
			case "playhead":
				row = append(row, formatDuration(int(int64Value(m.Playhead))))
			case "time_last_played":
				row = append(row, formatTime(int64Value(m.TimeLastPlayed)))
			case "db":
				row = append(row, filepath.Base(m.DB))
			}
		}
		fmt.Println(strings.Join(row, "\t"))
	}

	fmt.Printf("\n%d media files\n", len(media))
	return nil
}

func printFolders(columns []string, folders []FolderStats) error {
	if len(columns) == 0 {
		columns = []string{"path", "count", "size", "duration"}
	}

	fmt.Println(strings.Join(columns, "\t"))

	for _, f := range folders {
		var row []string
		for _, col := range columns {
			switch col {
			case "path":
				row = append(row, f.Path)
			case "count":
				row = append(row, fmt.Sprintf("%d", f.Count))
			case "size":
				row = append(row, formatSize(f.TotalSize))
			case "duration":
				row = append(row, formatDuration(f.TotalDuration))
			case "avg_size":
				row = append(row, formatSize(f.AvgSize))
			case "avg_duration":
				row = append(row, formatDuration(f.AvgDuration))
			}
		}
		fmt.Println(strings.Join(row, "\t"))
	}

	fmt.Printf("\n%d folders\n", len(folders))
	return nil
}

// Post-action execution

func executePostAction(flags GlobalFlags, media []MediaWithDB) error {
	action := flags.PostAction

	if flags.DeleteFiles {
		action = "delete"
	} else if flags.MarkDeleted {
		action = "mark-deleted"
	} else if flags.MoveTo != "" {
		action = "move"
	} else if flags.CopyTo != "" {
		action = "copy"
	}

	switch action {
	case "delete":
		return deleteMedia(media)
	case "mark-deleted":
		return markDeleted(media)
	case "move":
		return moveMedia(flags.MoveTo, media)
	case "copy":
		return copyMedia(flags.CopyTo, media)
	}

	return nil
}

func deleteMedia(media []MediaWithDB) error {
	for _, m := range media {
		if fileExists(m.Path) {
			if err := os.Remove(m.Path); err != nil {
				return err
			}
			fmt.Printf("Deleted: %s\n", m.Path)
		}
	}
	return nil
}

func markDeleted(media []MediaWithDB) error {
	byDB := make(map[string][]string)
	for _, m := range media {
		byDB[m.DB] = append(byDB[m.DB], m.Path)
	}

	for dbPath, paths := range byDB {
		sqlDB, err := sql.Open("sqlite3", dbPath)
		if err != nil {
			return err
		}

		tx, _ := sqlDB.Begin()
		now := time.Now().Unix()

		for _, path := range paths {
			tx.Exec("UPDATE media SET time_deleted = ? WHERE path = ?", now, path)
		}

		tx.Commit()
		sqlDB.Close()

		fmt.Printf("Marked %d files as deleted in %s\n", len(paths), filepath.Base(dbPath))
	}
	return nil
}

func moveMedia(destDir string, media []MediaWithDB) error {
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return err
	}

	for _, m := range media {
		if !fileExists(m.Path) {
			continue
		}

		dest := filepath.Join(destDir, filepath.Base(m.Path))
		if err := os.Rename(m.Path, dest); err != nil {
			return err
		}

		// Update database
		sqlDB, _ := sql.Open("sqlite3", m.DB)
		sqlDB.Exec("UPDATE media SET path = ? WHERE path = ?", dest, m.Path)
		sqlDB.Close()

		fmt.Printf("Moved: %s -> %s\n", m.Path, dest)
	}
	return nil
}

func copyMedia(destDir string, media []MediaWithDB) error {
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return err
	}

	for _, m := range media {
		if !fileExists(m.Path) {
			continue
		}

		dest := filepath.Join(destDir, filepath.Base(m.Path))
		data, _ := os.ReadFile(m.Path)
		os.WriteFile(dest, data, 0o644)

		fmt.Printf("Copied: %s -> %s\n", m.Path, dest)
	}
	return nil
}

// Helper functions

func getString(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

func getInt(v any) int {
	if i, ok := v.(int64); ok {
		return int(i)
	}
	return 0
}

func getInt64(v any) int64 {
	if i, ok := v.(int64); ok {
		return i
	}
	return 0
}

func matchesAny(path string, patterns []string) bool {
	for _, pattern := range patterns {
		if matched, _ := filepath.Match(pattern, path); matched {
			return true
		}
		if strings.Contains(path, pattern) {
			return true
		}
	}
	return false
}

func humanToBytes(s string) (int64, error) {
	s = strings.ToUpper(strings.TrimSpace(s))

	suffixes := []struct {
		suffix string
		mult   int64
	}{
		{"KB", 1024},
		{"MB", 1024 * 1024},
		{"GB", 1024 * 1024 * 1024},
		{"TB", 1024 * 1024 * 1024 * 1024},
		{"K", 1024},
		{"M", 1024 * 1024},
		{"G", 1024 * 1024 * 1024},
		{"T", 1024 * 1024 * 1024 * 1024},
		{"B", 1},
	}

	for _, entry := range suffixes {
		if before, ok := strings.CutSuffix(s, entry.suffix); ok {
			numStr := strings.TrimSpace(before)
			num, err := strconv.ParseFloat(numStr, 64)
			if err != nil {
				return 0, err
			}
			return int64(num * float64(entry.mult)), nil
		}
	}

	num, err := strconv.ParseInt(s, 10, 64)
	return num, err
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func formatDuration(seconds int) string {
	if seconds == 0 {
		return "-"
	}
	h := seconds / 3600
	m := (seconds % 3600) / 60
	s := seconds % 60
	if h > 0 {
		return fmt.Sprintf("%d:%02d:%02d", h, m, s)
	}
	return fmt.Sprintf("%d:%02d", m, s)
}

func formatSize(bytes int64) string {
	if bytes == 0 {
		return "-"
	}
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func formatTime(timestamp int64) string {
	if timestamp == 0 {
		return "-"
	}
	t := time.Unix(timestamp, 0)
	return t.Format("2006-01-02 15:04")
}

func getDefaultBrowser() string {
	switch runtime.GOOS {
	case "linux":
		return "xdg-open"
	case "darwin":
		return "open"
	case "windows":
		return "start"
	default:
		return "xdg-open"
	}
}

func main() {
	cli := &CLI{}
	ctx := kong.Parse(cli,
		kong.Name("lb"),
		kong.Description("Library media management tool"),
		kong.UsageOnError(),
	)
	err := ctx.Run(ctx)
	ctx.FatalIfErrorf(err)
}
