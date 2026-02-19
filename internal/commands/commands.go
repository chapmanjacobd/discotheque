package commands

import (
	"context"
	"database/sql"
	"embed"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/alecthomas/kong"
	"github.com/chapmanjacobd/discotheque/internal/aggregate"
	"github.com/chapmanjacobd/discotheque/internal/db"
	"github.com/chapmanjacobd/discotheque/internal/fs"
	"github.com/chapmanjacobd/discotheque/internal/history"
	"github.com/chapmanjacobd/discotheque/internal/metadata"
	"github.com/chapmanjacobd/discotheque/internal/models"
	"github.com/chapmanjacobd/discotheque/internal/query"
	"github.com/chapmanjacobd/discotheque/internal/tui"
	"github.com/chapmanjacobd/discotheque/internal/utils"
	tea "github.com/charmbracelet/bubbletea"
)

//go:embed schema.sql
var schemaFS embed.FS

type PrintCmd struct {
	models.GlobalFlags
	Args []string `arg:"" required:"" help:"Database file(s) or files/directories to scan"`

	Databases []string `kong:"-"`
	ScanPaths []string `kong:"-"`
}

func (c *PrintCmd) AfterApply() error {
	for _, arg := range c.Args {
		if strings.HasSuffix(arg, ".db") && utils.IsSQLite(arg) {
			c.Databases = append(c.Databases, arg)
		} else {
			c.ScanPaths = append(c.ScanPaths, arg)
		}
	}
	return nil
}

func (c *PrintCmd) Run(ctx *kong.Context) error {
	models.SetupLogging(c.Verbose)

	var allMedia []models.MediaWithDB

	// Handle databases
	if len(c.Databases) > 0 {
		dbMedia, err := query.MediaQuery(context.Background(), c.Databases, c.GlobalFlags)
		if err != nil {
			return err
		}
		allMedia = append(allMedia, dbMedia...)
	}

	// Handle paths
	for _, root := range c.ScanPaths {
		err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				return nil
			}

			meta, err := metadata.Extract(context.Background(), path, c.ScanSubtitles)
			if err != nil {
				return nil
			}
			allMedia = append(allMedia, models.MediaWithDB{
				Media: models.Media{
					Path:         meta.Media.Path,
					Title:        models.NullStringPtr(meta.Media.Title),
					Type:         models.NullStringPtr(meta.Media.Type),
					Size:         models.NullInt64Ptr(meta.Media.Size),
					Duration:     models.NullInt64Ptr(meta.Media.Duration),
					TimeCreated:  models.NullInt64Ptr(meta.Media.TimeCreated),
					TimeModified: models.NullInt64Ptr(meta.Media.TimeModified),
				},
			})
			return nil
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error walking %s: %v\n", root, err)
		}
	}

	var media []models.MediaWithDB
	media = query.FilterMedia(allMedia, c.GlobalFlags)
	HideRedundantFirstPlayed(media)

	isAggregated := c.BigDirs || c.GroupByExtensions || c.GroupByMimeTypes || c.GroupBySize || c.Depth > 0 || c.Parents

	if c.JSON {
		if isAggregated {
			folders := query.AggregateMedia(media, c.GlobalFlags)
			query.SortFolders(folders, c.SortBy, c.Reverse)
			encoder := json.NewEncoder(os.Stdout)
			encoder.SetIndent("", "  ")
			return encoder.Encode(folders)
		}
		if c.Summarize {
			summary := query.SummarizeMedia(media)
			encoder := json.NewEncoder(os.Stdout)
			encoder.SetIndent("", "  ")
			return encoder.Encode(summary)
		}
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(media)
	}

	if c.Summarize {
		summary := query.SummarizeMedia(media)
		for _, s := range summary {
			fmt.Printf("%s: %d files, %s, %s\n",
				s.Label, s.Count, utils.FormatSize(s.TotalSize), utils.FormatDuration(int(s.TotalDuration)))
		}
		if !isAggregated {
			fmt.Println()
		}
	}

	if isAggregated {
		folders := query.AggregateMedia(media, c.GlobalFlags)
		query.SortFolders(folders, c.SortBy, c.Reverse)
		return PrintFolders(c.Columns, folders)
	}

	if c.RegexSort {
		media = query.RegexSortMedia(media, c.GlobalFlags)
	} else {
		query.SortMedia(media, c.SortBy, c.Reverse, c.NatSort)
	}
	return PrintMedia(c.Columns, media)
}

type DiskUsageCmd struct {
	models.GlobalFlags
	Args []string `arg:"" required:"" help:"Database file(s) or files/directories to scan"`

	Databases []string `kong:"-"`
	ScanPaths []string `kong:"-"`
}

func (c *DiskUsageCmd) AfterApply() error {
	for _, arg := range c.Args {
		if strings.HasSuffix(arg, ".db") && utils.IsSQLite(arg) {
			c.Databases = append(c.Databases, arg)
		} else {
			c.ScanPaths = append(c.ScanPaths, arg)
		}
	}
	return nil
}

func (c *DiskUsageCmd) Run(ctx *kong.Context) error {
	models.SetupLogging(c.Verbose)

	var allMedia []models.MediaWithDB

	// Handle databases
	if len(c.Databases) > 0 {
		dbMedia, err := query.MediaQuery(context.Background(), c.Databases, c.GlobalFlags)
		if err != nil {
			return err
		}
		allMedia = append(allMedia, dbMedia...)
	}

	// Handle paths
	for _, root := range c.ScanPaths {
		err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				return nil
			}

			meta, err := metadata.Extract(context.Background(), path, c.ScanSubtitles)
			if err != nil {
				return nil
			}
			allMedia = append(allMedia, models.MediaWithDB{
				Media: models.Media{
					Path:         meta.Media.Path,
					Title:        models.NullStringPtr(meta.Media.Title),
					Type:         models.NullStringPtr(meta.Media.Type),
					Size:         models.NullInt64Ptr(meta.Media.Size),
					Duration:     models.NullInt64Ptr(meta.Media.Duration),
					TimeCreated:  models.NullInt64Ptr(meta.Media.TimeCreated),
					TimeModified: models.NullInt64Ptr(meta.Media.TimeModified),
				},
			})
			return nil
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error walking %s: %v\n", root, err)
		}
	}

	if c.TUI {
		if len(allMedia) == 0 {
			return fmt.Errorf("no media found")
		}

		m := tui.NewDUModel(allMedia, c.GlobalFlags)
		p := tea.NewProgram(m, tea.WithAltScreen())
		_, err := p.Run()
		return err
	}

	// Disk usage is essentially Print with aggregation by default if no depth specified
	if !c.BigDirs && !c.GroupByExtensions && !c.GroupByMimeTypes && !c.GroupBySize && c.Depth == 0 && !c.Parents {
		c.BigDirs = true
	}
	printCmd := PrintCmd{GlobalFlags: c.GlobalFlags, Databases: c.Databases, ScanPaths: c.ScanPaths}
	return printCmd.Run(ctx)
}

type SimilarFilesCmd struct {
	models.GlobalFlags
	Databases []string `arg:"" required:"" help:"SQLite database files" type:"existingfile"`
}

func (c *SimilarFilesCmd) Run(ctx *kong.Context) error {
	models.SetupLogging(c.Verbose)
	media, err := query.MediaQuery(context.Background(), c.Databases, c.GlobalFlags)
	if err != nil {
		return err
	}

	media = query.FilterMedia(media, c.GlobalFlags)

	// Defaults for similar files
	if !c.FilterSizes && !c.FilterDurations && !c.FilterNames {
		c.FilterSizes = true
		c.FilterDurations = true
	}

	groups := aggregate.ClusterByNumbers(c.GlobalFlags, media)

	if c.OnlyOriginals || c.OnlyDuplicates {
		for i, g := range groups {
			if c.OnlyOriginals {
				groups[i].Files = g.Files[:1]
			} else if c.OnlyDuplicates {
				groups[i].Files = g.Files[1:]
			}
		}
	}

	return PrintFolders(c.Columns, groups)
}

type SimilarFoldersCmd struct {
	models.GlobalFlags
	Databases []string `arg:"" required:"" help:"SQLite database files" type:"existingfile"`
}

func (c *SimilarFoldersCmd) Run(ctx *kong.Context) error {
	models.SetupLogging(c.Verbose)
	media, err := query.MediaQuery(context.Background(), c.Databases, c.GlobalFlags)
	if err != nil {
		return err
	}

	media = query.FilterMedia(media, c.GlobalFlags)

	// Defaults for similar folders
	if !c.FilterSizes && !c.FilterDurations && !c.FilterNames && !c.FilterCounts {
		c.FilterCounts = true
		c.FilterSizes = true
	}

	folders := query.AggregateMedia(media, c.GlobalFlags)
	groups := aggregate.ClusterFoldersByNumbers(c.GlobalFlags, folders)

	return PrintFolders(c.Columns, groups)
}

type WatchCmd struct {
	models.GlobalFlags
	Databases []string `arg:"" required:"" help:"SQLite database files" type:"existingfile"`
}

func (c *WatchCmd) Run(ctx *kong.Context) error {
	models.SetupLogging(c.Verbose)
	media, err := query.MediaQuery(context.Background(), c.Databases, c.GlobalFlags)
	if err != nil {
		return err
	}

	media = query.FilterMedia(media, c.GlobalFlags)
	query.SortMedia(media, c.SortBy, c.Reverse, c.NatSort)

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
	if c.Mute {
		args = append(args, "--mute=yes")
	}
	if c.Loop {
		args = append(args, "--loop-file=inf")
	}

	ipcSocket := c.MpvSocket
	if ipcSocket == "" {
		ipcSocket = utils.GetMpvWatchSocket()
	}
	args = append(args, fmt.Sprintf("--input-ipc-server=%s", ipcSocket))

	// Add file paths
	var paths []string
	for _, m := range media {
		if utils.FileExists(m.Path) {
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

	startTime := time.Now()
	if err := cmd.Run(); err != nil {
		return err
	}

	// Update history
	if c.TrackHistory {
		for _, m := range media {
			mediaDuration := 0
			if m.Duration != nil {
				mediaDuration = int(*m.Duration)
			}
			existingPlayhead := 0
			if m.Playhead != nil {
				existingPlayhead = int(*m.Playhead)
			}
			playhead := utils.GetPlayhead(c.GlobalFlags, m.Path, startTime, existingPlayhead, mediaDuration)

			if err := history.UpdateHistorySimple(m.DB, []string{m.Path}, playhead, false); err != nil {
				slog.Error("Warning: failed to update history", "path", m.Path, "error", err)
			}
		}
	}

	return ExecutePostAction(c.GlobalFlags, media)
}

type ListenCmd struct {
	models.GlobalFlags
	Databases []string `arg:"" required:"" help:"SQLite database files" type:"existingfile"`
}

func (c *ListenCmd) Run(ctx *kong.Context) error {
	models.SetupLogging(c.Verbose)
	media, err := query.MediaQuery(context.Background(), c.Databases, c.GlobalFlags)
	if err != nil {
		return err
	}

	media = query.FilterMedia(media, c.GlobalFlags)
	query.SortMedia(media, c.SortBy, c.Reverse, c.NatSort)

	args := []string{"mpv", "--video=no"}

	if c.Volume > 0 {
		args = append(args, fmt.Sprintf("--volume=%d", c.Volume))
	}
	if c.Speed != 1.0 {
		args = append(args, fmt.Sprintf("--speed=%.2f", c.Speed))
	}
	if c.Mute {
		args = append(args, "--mute=yes")
	}
	if c.Loop {
		args = append(args, "--loop-file=inf")
	}

	ipcSocket := c.MpvSocket
	if ipcSocket == "" {
		ipcSocket = utils.GetMpvWatchSocket()
	}
	args = append(args, fmt.Sprintf("--input-ipc-server=%s", ipcSocket))

	for _, m := range media {
		if utils.FileExists(m.Path) {
			args = append(args, m.Path)
		}
	}

	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	startTime := time.Now()
	if err := cmd.Run(); err != nil {
		return err
	}

	if c.TrackHistory {
		for _, m := range media {
			mediaDuration := 0
			if m.Duration != nil {
				mediaDuration = int(*m.Duration)
			}
			existingPlayhead := 0
			if m.Playhead != nil {
				existingPlayhead = int(*m.Playhead)
			}
			playhead := utils.GetPlayhead(c.GlobalFlags, m.Path, startTime, existingPlayhead, mediaDuration)

			history.UpdateHistorySimple(m.DB, []string{m.Path}, playhead, false)
		}
	}

	return ExecutePostAction(c.GlobalFlags, media)
}

type OpenCmd struct {
	models.GlobalFlags
	Databases []string `arg:"" required:"" help:"SQLite database files" type:"existingfile"`
}

func (c *OpenCmd) Run(ctx *kong.Context) error {
	models.SetupLogging(c.Verbose)
	media, err := query.MediaQuery(context.Background(), c.Databases, c.GlobalFlags)
	if err != nil {
		return err
	}

	media = query.FilterMedia(media, c.GlobalFlags)

	for _, m := range media {
		if !utils.FileExists(m.Path) {
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

	return ExecutePostAction(c.GlobalFlags, media)
}

type BrowseCmd struct {
	models.GlobalFlags
	Databases []string `arg:"" required:"" help:"SQLite database files" type:"existingfile"`
	Browser   string   `help:"Browser to use"`
}

func (c *BrowseCmd) Run(ctx *kong.Context) error {
	models.SetupLogging(c.Verbose)
	media, err := query.MediaQuery(context.Background(), c.Databases, c.GlobalFlags)
	if err != nil {
		return err
	}

	media = query.FilterMedia(media, c.GlobalFlags)

	browser := c.Browser
	if browser == "" {
		browser = utils.GetDefaultBrowser()
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

type StatsCmd struct {
	models.GlobalFlags
	Facet     string   `arg:"" required:"" help:"One of: watched, deleted, created, modified"`
	Databases []string `arg:"" required:"" help:"SQLite database files" type:"existingfile"`
}

func (c *StatsCmd) Run(ctx *kong.Context) error {
	models.SetupLogging(c.Verbose)

	timeCol := "time_last_played"
	switch c.Facet {
	case "deleted":
		timeCol = "time_deleted"
		c.MarkDeleted = true // Ensure we don't hide deleted in query
	case "created":
		timeCol = "time_created"
	case "modified":
		timeCol = "time_modified"
	}

	for _, dbPath := range c.Databases {
		sqlDB, err := db.Connect(dbPath)
		if err != nil {
			return err
		}
		defer sqlDB.Close()

		if err := InitDB(sqlDB); err != nil {
			return fmt.Errorf("failed to initialize database %s: %w", dbPath, err)
		}

		if c.Frequency != "" {
			stats, err := query.HistoricalUsage(context.Background(), dbPath, c.Frequency, timeCol)
			if err != nil {
				return err
			}

			if c.JSON {
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				if err := encoder.Encode(stats); err != nil {
					return err
				}
				continue
			}

			fmt.Printf("%s media (%s) for %s:\n", utils.Title(c.Facet), c.Frequency, dbPath)
			if err := PrintFrequencyStats(stats); err != nil {
				return err
			}
			continue
		}

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
		fmt.Printf("  Total Size:       %s\n", utils.FormatSize(utils.GetInt64(stats.TotalSize)))
		fmt.Printf("  Total Duration:   %s\n", utils.FormatDuration(int(utils.GetInt64(stats.TotalDuration))))
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
					utils.FormatSize(utils.GetInt64(ts.TotalSize)),
					utils.FormatDuration(int(utils.GetInt64(ts.TotalDuration))))
			}
		}
		fmt.Println()
	}
	return nil
}

type PlaylistsCmd struct {
	models.GlobalFlags
	Databases []string `arg:"" required:"" help:"SQLite database files" type:"existingfile"`
}

func (c *PlaylistsCmd) Run(ctx *kong.Context) error {
	models.SetupLogging(c.Verbose)
	for _, dbPath := range c.Databases {
		sqlDB, err := db.Connect(dbPath)
		if err != nil {
			return err
		}
		defer sqlDB.Close()

		queries := db.New(sqlDB)
		playlists, err := queries.GetPlaylists(context.Background())
		if err != nil {
			return err
		}

		if c.JSON {
			encoder := json.NewEncoder(os.Stdout)
			encoder.SetIndent("", "  ")
			if err := encoder.Encode(playlists); err != nil {
				return err
			}
			continue
		}

		fmt.Printf("Playlists in %s:\n", dbPath)
		for _, pl := range playlists {
			fmt.Printf("  %s (%s)\n", utils.StringValue(models.NullStringPtr(pl.Path)), utils.StringValue(models.NullStringPtr(pl.ExtractorKey)))
		}
		fmt.Println()
	}
	return nil
}

type SearchCmd struct {
	models.GlobalFlags
	Databases []string `arg:"" required:"" help:"SQLite database files" type:"existingfile"`
}

func (c *SearchCmd) Run(ctx *kong.Context) error {
	models.SetupLogging(c.Verbose)
	c.FTS = true // Force FTS for search command
	media, err := query.MediaQuery(context.Background(), c.Databases, c.GlobalFlags)
	if err != nil {
		return err
	}

	media = query.FilterMedia(media, c.GlobalFlags)

	if c.JSON {
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(media)
	}

	return PrintMedia(c.Columns, media)
}

func PrintFrequencyStats(stats []query.FrequencyStats) error {
	fmt.Printf("%-15s\t%-10s\t%-10s\t%-15s\n", "Period", "Count", "Size", "Duration")
	for _, s := range stats {
		fmt.Printf("%-15s\t%-10d\t%-10s\t%-15s\n",
			s.Label, s.Count, utils.FormatSize(s.TotalSize), utils.FormatDuration(int(s.TotalDuration)))
	}
	return nil
}

type HistoryCmd struct {
	models.GlobalFlags
	Databases []string `arg:"" required:"" help:"SQLite database files" type:"existingfile"`
}

func HideRedundantFirstPlayed(media []models.MediaWithDB) {
	for i := range media {
		if media[i].PlayCount != nil && *media[i].PlayCount <= 1 {
			media[i].TimeFirstPlayed = nil
		}
	}
}

func (c *HistoryCmd) Run(ctx *kong.Context) error {
	models.SetupLogging(c.Verbose)
	// Set default sort for history
	if c.SortBy == "path" || c.SortBy == "" {
		c.SortBy = "time_last_played"
		c.Reverse = true
	}

	media, err := query.MediaQuery(context.Background(), c.Databases, c.GlobalFlags)
	if err != nil {
		return err
	}

	// Filter for only watched items if not otherwise specified
	if c.Watched == nil && !c.InProgress && !c.Completed {
		watched := true
		c.Watched = &watched
	}

	media = query.FilterMedia(media, c.GlobalFlags)
	HideRedundantFirstPlayed(media)

	if c.JSON {
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(media)
	}

	if c.Completed {
		fmt.Println("Completed:")
	} else if c.InProgress {
		fmt.Println("In progress:")
	} else {
		fmt.Println("History:")
	}

	if c.DeleteRows {
		for _, dbPath := range c.Databases {
			var paths []string
			for _, m := range media {
				if m.DB == dbPath {
					paths = append(paths, m.Path)
				}
			}
			if len(paths) > 0 {
				if err := history.DeleteHistoryByPaths(dbPath, paths); err != nil {
					return err
				}
			}
		}
		fmt.Printf("Deleted history for %d items\n", len(media))
		return nil
	}

	if c.Partial != "" {
		query.SortHistory(media, c.Partial, c.Reverse)
	} else {
		query.SortMedia(media, c.SortBy, c.Reverse, c.NatSort)
	}
	return PrintMedia(c.Columns, media)
}

type HistoryAddCmd struct {
	models.GlobalFlags
	Args []string `arg:"" name:"args" required:"" help:"Database file followed by paths to mark as played"`

	Paths    []string `kong:"-"`
	Database string   `kong:"-"`
}

func (c *HistoryAddCmd) AfterApply() error {
	if len(c.Args) < 2 {
		return fmt.Errorf("at least one database file and one path are required")
	}
	c.Database = c.Args[0]
	c.Paths = c.Args[1:]
	return nil
}

func (c *HistoryAddCmd) Run(ctx *kong.Context) error {
	models.SetupLogging(c.Verbose)

	var absPaths []string
	for _, p := range c.Paths {
		abs, err := filepath.Abs(p)
		if err == nil {
			absPaths = append(absPaths, abs)
		} else {
			absPaths = append(absPaths, p)
		}
	}

	err := history.UpdateHistorySimple(c.Database, absPaths, 0, true)
	if err == nil {
		slog.Info("History added", "count", len(absPaths), "database", c.Database)
	}
	return err
}

type OptimizeCmd struct {
	models.GlobalFlags
	Databases []string `arg:"" required:"" help:"SQLite database files" type:"existingfile"`
}

func (c *OptimizeCmd) Run(ctx *kong.Context) error {
	models.SetupLogging(c.Verbose)
	for _, dbPath := range c.Databases {
		slog.Info("Optimizing database", "path", dbPath)
		sqlDB, err := db.Connect(dbPath)
		if err != nil {
			return err
		}
		defer sqlDB.Close()

		slog.Info("Running VACUUM...")
		if _, err := sqlDB.Exec("VACUUM"); err != nil {
			return fmt.Errorf("VACUUM failed on %s: %w", dbPath, err)
		}

		slog.Info("Running ANALYZE...")
		if _, err := sqlDB.Exec("ANALYZE"); err != nil {
			return fmt.Errorf("ANALYZE failed on %s: %w", dbPath, err)
		}

		slog.Info("Optimizing FTS index...")
		// FTS5 optimize command
		if _, err := sqlDB.Exec("INSERT INTO media_fts(media_fts) VALUES('optimize')"); err != nil {
			slog.Warn("FTS optimize failed (maybe table doesn't exist?)", "path", dbPath, "error", err)
		}

		slog.Info("Optimization complete", "path", dbPath)
	}
	return nil
}

type AddCmd struct {
	models.GlobalFlags
	Args     []string `arg:"" name:"args" required:"" help:"Database file followed by paths to scan"`
	Parallel int      `short:"p" help:"Number of parallel extractors (default: CPU count * 4)"`

	ScanPaths []string `kong:"-"`
	Database  string   `kong:"-"`
}

func (c *AddCmd) AfterApply() error {
	if len(c.Args) < 2 {
		return fmt.Errorf("at least one database file and one path to scan are required")
	}

	// Smart DB detection: first arg MUST be a database for 'add'
	isDB := strings.HasSuffix(c.Args[0], ".db") && (utils.IsSQLite(c.Args[0]) || !utils.FileExists(c.Args[0]))
	if isDB {
		c.Database = c.Args[0]
		c.ScanPaths = c.Args[1:]
	} else {
		return fmt.Errorf("first argument must be a database file (e.g. .db): %s", c.Args[0])
	}

	if c.Parallel <= 0 {
		c.Parallel = runtime.NumCPU() * 4
	}
	return nil
}

func (c *AddCmd) Run(ctx *kong.Context) error {
	models.SetupLogging(c.Verbose)
	dbPath := c.Database
	c.ScanPaths = utils.ExpandStdin(c.ScanPaths)

	sqlDB, err := db.Connect(dbPath)
	if err != nil {
		return err
	}
	defer sqlDB.Close()

	if err := InitDB(sqlDB); err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}

	queries := db.New(sqlDB)

	// Step 0: Load existing playlists (roots) to avoid redundant scans
	existingPlaylists, _ := queries.GetPlaylists(context.Background())

	// Step 1: Load all existing metadata into memory for O(1) checks
	slog.Info("Loading existing metadata from database...")
	existingMedia, err := queries.GetAllMediaMetadata(context.Background())
	if err != nil {
		return fmt.Errorf("failed to load existing metadata: %w", err)
	}

	type meta struct {
		size    int64
		mtime   int64
		deleted bool
	}
	metaCache := make(map[string]meta, len(existingMedia))
	for _, m := range existingMedia {
		metaCache[m.Path] = meta{
			size:    m.Size.Int64,
			mtime:   m.TimeModified.Int64,
			deleted: m.TimeDeleted.Int64 > 0,
		}
	}
	slog.Info("Loaded metadata cache", "count", len(metaCache))

	for _, root := range c.ScanPaths {
		absRoot, err := filepath.Abs(root)
		if err != nil {
			slog.Error("Failed to get absolute path", "path", root, "error", err)
			continue
		}

		// Check if this path or a parent is already a playlist
		isSubpath := false
		for _, pl := range existingPlaylists {
			if pl.Path.Valid && (absRoot == pl.Path.String || strings.HasPrefix(absRoot, pl.Path.String+string(filepath.Separator))) {
				slog.Info("Path already covered by existing scan root", "path", absRoot, "root", pl.Path.String)
				isSubpath = true
				break
			}
		}
		if isSubpath {
			continue
		}

		// Record this new scan root
		queries.InsertPlaylist(context.Background(), db.InsertPlaylistParams{
			Path:         sql.NullString{String: absRoot, Valid: true},
			ExtractorKey: sql.NullString{String: "Local", Valid: true},
		})

		slog.Info("Scanning", "path", absRoot)
		foundFiles, err := fs.FindMedia(absRoot)
		if err != nil {
			return err
		}

		slog.Info("Checking for updates", "count", len(foundFiles))

		// Step 2: Identify which files actually need probing using the cache
		var toProbe []string
		skipped := 0
		for path, stat := range foundFiles {
			if len(c.Ext) > 0 {
				matched := false
				for _, e := range c.Ext {
					if strings.EqualFold(filepath.Ext(path), e) {
						matched = true
						break
					}
				}
				if !matched {
					continue
				}
			}

			if existing, ok := metaCache[path]; ok {
				// Record exists, check if it's still valid
				if !existing.deleted && existing.size == stat.Size() && existing.mtime == stat.ModTime().Unix() {
					skipped++
					continue
				}
			}
			toProbe = append(toProbe, path)
		}

		if skipped > 0 {
			slog.Info("Skipped unchanged files", "count", skipped)
		}

		if len(toProbe) == 0 {
			continue
		}

		slog.Info("Extracting metadata", "count", len(toProbe), "parallelism", c.Parallel)

		// Parallel extraction
		jobs := make(chan string, len(toProbe))
		results := make(chan *metadata.MediaMetadata, len(toProbe))
		var wg sync.WaitGroup

		for i := 0; i < c.Parallel; i++ {
			wg.Go(func() {
				for path := range jobs {
					res, err := metadata.Extract(context.Background(), path, c.ScanSubtitles)
					if err != nil {
						slog.Error("Metadata extraction failed", "path", path, "error", err)
						continue
					}
					results <- res
				}
			})
		}

		go func() {
			for _, f := range toProbe {
				jobs <- f
			}
			close(jobs)
		}()

		go func() {
			wg.Wait()
			close(results)
		}()

		count := 0
		batchSize := 100
		var currentBatch []*metadata.MediaMetadata

		flush := func() error {
			if len(currentBatch) == 0 {
				return nil
			}
			tx, err := sqlDB.Begin()
			if err != nil {
				return err
			}
			defer tx.Rollback()

			qtx := queries.WithTx(tx)
			for _, res := range currentBatch {
				if err := qtx.UpsertMedia(context.Background(), res.Media); err != nil {
					slog.Error("Database upsert failed", "path", res.Media.Path, "error", err)
				}
				for _, cap := range res.Captions {
					if err := qtx.InsertCaption(context.Background(), cap); err != nil {
						slog.Error("Caption insertion failed", "path", res.Media.Path, "error", err)
					}
				}
			}
			return tx.Commit()
		}

		for res := range results {
			currentBatch = append(currentBatch, res)
			if len(currentBatch) >= batchSize {
				if err := flush(); err != nil {
					slog.Error("Failed to commit batch", "error", err)
				}
				currentBatch = currentBatch[:0]
			}

			count++
			if count%10 == 0 || count == len(toProbe) {
				fmt.Printf("\rProcessed %d/%d", count, len(toProbe))
			}
		}
		// Final flush
		if err := flush(); err != nil {
			slog.Error("Failed to commit final batch", "error", err)
		}
		fmt.Println()
	}

	return nil
}

type CheckCmd struct {
	models.GlobalFlags
	Args   []string `arg:"" required:"" help:"Database file followed by optional paths to check"`
	DryRun bool     `help:"Don't actually mark files as deleted"`

	CheckPaths []string `kong:"-"`
	Databases  []string `kong:"-"`
}

func (c *CheckCmd) AfterApply() error {
	if len(c.Args) < 1 {
		return fmt.Errorf("at least one database file is required")
	}

	if utils.IsSQLite(c.Args[0]) || strings.HasSuffix(c.Args[0], ".db") {
		c.Databases = []string{c.Args[0]}
		if len(c.Args) > 1 {
			c.CheckPaths = c.Args[1:]
		}
	} else {
		// Fallback: first is DB
		c.Databases = []string{c.Args[0]}
		if len(c.Args) > 1 {
			c.CheckPaths = c.Args[1:]
		}
	}
	return nil
}

func (c *CheckCmd) Run(ctx *kong.Context) error {
	models.SetupLogging(c.Verbose)
	c.CheckPaths = utils.ExpandStdin(c.CheckPaths)

	// If paths provided, build a presence set
	var presenceSet map[string]bool
	var absCheckPaths []string
	if len(c.CheckPaths) > 0 {
		presenceSet = make(map[string]bool)
		for _, root := range c.CheckPaths {
			absRoot, err := filepath.Abs(root)
			if err != nil {
				return err
			}
			absCheckPaths = append(absCheckPaths, absRoot)
			slog.Info("Scanning filesystem for presence set", "path", absRoot)
			err = filepath.WalkDir(absRoot, func(path string, d os.DirEntry, err error) error {
				if err == nil && !d.IsDir() {
					absPath, _ := filepath.Abs(path)
					presenceSet[absPath] = true
				}
				return nil
			})
			if err != nil {
				return err
			}
		}
	}

	for _, dbPath := range c.Databases {
		sqlDB, err := db.Connect(dbPath)
		if err != nil {
			return err
		}
		defer sqlDB.Close()

		if err := InitDB(sqlDB); err != nil {
			return fmt.Errorf("failed to initialize database %s: %w", dbPath, err)
		}

		queries := db.New(sqlDB)
		allMedia, err := queries.GetMedia(context.Background(), 1000000)
		if err != nil {
			return err
		}

		slog.Info("Checking files", "count", len(allMedia), "database", dbPath)

		missingCount := 0
		now := time.Now().Unix()

		for _, m := range allMedia {
			isMissing := false

			if presenceSet != nil {
				// Only check files that are within the scanned roots
				inScannedRoot := false
				for _, root := range absCheckPaths {
					if strings.HasPrefix(m.Path, root) {
						inScannedRoot = true
						break
					}
				}

				if inScannedRoot {
					if !presenceSet[m.Path] {
						isMissing = true
					}
				} else {
					// Outside scanned roots, skip or use Stat?
					// For safety, if user provided roots, we only check files in those roots.
					continue
				}
			} else {
				// No presence set, fallback to individual Stats
				if !utils.FileExists(m.Path) {
					isMissing = true
				}
			}

			if isMissing {
				missingCount++
				if !c.DryRun {
					slog.Debug("Marking missing file as deleted", "path", m.Path)
					if err := queries.MarkDeleted(context.Background(), db.MarkDeletedParams{
						TimeDeleted: sql.NullInt64{Int64: now, Valid: true},
						Path:        m.Path,
					}); err != nil {
						slog.Error("Failed to mark file as deleted", "path", m.Path, "error", err)
					}
				} else {
					fmt.Printf("[Dry-run] Missing: %s\n", m.Path)
				}
			}
		}

		if c.DryRun {
			slog.Info("Check complete (dry-run)", "missing", missingCount)
		} else {
			slog.Info("Check complete", "marked_deleted", missingCount)
		}
	}
	return nil
}

func PrintMedia(columns []string, media []models.MediaWithDB) error {
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
				row = append(row, utils.StringValue(m.Title))
			case "duration":
				row = append(row, utils.FormatDuration(int(utils.Int64Value(m.Duration))))
			case "size":
				row = append(row, utils.FormatSize(utils.Int64Value(m.Size)))
			case "play_count":
				row = append(row, fmt.Sprintf("%d", utils.Int64Value(m.PlayCount)))
			case "playhead":
				row = append(row, utils.FormatDuration(int(utils.Int64Value(m.Playhead))))
			case "time_last_played":
				row = append(row, utils.FormatTime(utils.Int64Value(m.TimeLastPlayed)))
			case "db":
				row = append(row, filepath.Base(m.DB))
			}
		}
		fmt.Println(strings.Join(row, "\t"))
	}

	fmt.Printf("\n%d media files\n", len(media))
	return nil
}

func PrintFolders(columns []string, folders []models.FolderStats) error {
	if len(columns) == 0 {
		columns = []string{"path", "exists_count", "size", "duration"}
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
			case "exists_count":
				row = append(row, fmt.Sprintf("%d", f.ExistsCount))
			case "deleted_count":
				row = append(row, fmt.Sprintf("%d", f.DeletedCount))
			case "played_count":
				row = append(row, fmt.Sprintf("%d", f.PlayedCount))
			case "size":
				row = append(row, utils.FormatSize(f.TotalSize))
			case "duration":
				row = append(row, utils.FormatDuration(int(f.TotalDuration)))
			case "avg_size":
				row = append(row, utils.FormatSize(f.AvgSize))
			case "avg_duration":
				row = append(row, utils.FormatDuration(int(f.AvgDuration)))
			case "median_size":
				row = append(row, utils.FormatSize(f.MedianSize))
			case "median_duration":
				row = append(row, utils.FormatDuration(int(f.MedianDuration)))
			case "folder_count":
				row = append(row, fmt.Sprintf("%d", f.FolderCount))
			}
		}
		fmt.Println(strings.Join(row, "\t"))
	}

	fmt.Printf("\n%d groups\n", len(folders))
	return nil
}

// ExecutePostAction executes actions after a command
func ExecutePostAction(flags models.GlobalFlags, media []models.MediaWithDB) error {
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

	if action == "" || action == "none" {
		return nil
	}

	var sizeLimit int64 = 0
	if flags.ActionSize != "" {
		if sl, err := utils.HumanToBytes(flags.ActionSize); err == nil {
			sizeLimit = sl
		}
	}

	var totalSize int64 = 0
	var count int = 0

	for _, m := range media {
		if flags.ActionLimit > 0 && count >= flags.ActionLimit {
			slog.Info("Action limit reached", "limit", flags.ActionLimit)
			break
		}
		if sizeLimit > 0 && totalSize >= sizeLimit {
			slog.Info("Action size limit reached", "limit", flags.ActionSize)
			break
		}

		var err error
		var size int64 = 0
		if m.Size != nil {
			size = *m.Size
		}

		switch action {
		case "delete":
			err = DeleteMediaItem(m)
		case "mark-deleted":
			err = MarkDeletedItem(m)
		case "move":
			err = MoveMediaItem(flags.MoveTo, m)
		case "copy":
			err = CopyMediaItem(flags.CopyTo, m)
		}

		if err != nil {
			slog.Error("Post-action failed", "path", m.Path, "error", err)
		} else {
			count++
			totalSize += size
		}
	}

	if count > 0 {
		fmt.Printf("\n%s %d files (%s total)\n", action, count, utils.FormatSize(totalSize))
	}

	return nil
}

func DeleteMediaItem(m models.MediaWithDB) error {
	if utils.FileExists(m.Path) {
		if err := os.Remove(m.Path); err != nil {
			return err
		}
		fmt.Printf("Deleted: %s\n", m.Path)
	}
	return nil
}

func MarkDeletedItem(m models.MediaWithDB) error {
	sqlDB, err := db.Connect(m.DB)
	if err != nil {
		return err
	}
	defer sqlDB.Close()

	now := time.Now().Unix()
	_, err = sqlDB.Exec("UPDATE media SET time_deleted = ? WHERE path = ?", now, m.Path)
	if err == nil {
		fmt.Printf("Marked deleted: %s\n", m.Path)
	}
	return err
}

func MoveMediaItem(destDir string, m models.MediaWithDB) error {
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return err
	}

	if !utils.FileExists(m.Path) {
		return fmt.Errorf("file not found")
	}

	dest := filepath.Join(destDir, filepath.Base(m.Path))
	if err := os.Rename(m.Path, dest); err != nil {
		return err
	}

	// Update database
	sqlDB, err := db.Connect(m.DB)
	if err != nil {
		return err
	}
	defer sqlDB.Close()
	_, err = sqlDB.Exec("UPDATE media SET path = ? WHERE path = ?", dest, m.Path)
	if err == nil {
		fmt.Printf("Moved: %s -> %s\n", m.Path, dest)
	}
	return err
}

func CopyMediaItem(destDir string, m models.MediaWithDB) error {
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return err
	}

	if !utils.FileExists(m.Path) {
		return fmt.Errorf("file not found")
	}

	dest := filepath.Join(destDir, filepath.Base(m.Path))
	data, err := os.ReadFile(m.Path)
	if err != nil {
		return err
	}
	if err := os.WriteFile(dest, data, 0o644); err != nil {
		return err
	}

	fmt.Printf("Copied: %s -> %s\n", m.Path, dest)
	return nil
}
