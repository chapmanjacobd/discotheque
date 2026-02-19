package models

import (
	"log/slog"
	"strings"
)

// GlobalFlags are flags available to all commands
type GlobalFlags struct {
	Query  string `short:"q" help:"Raw SQL query (overrides all query building)"`
	Limit  int    `short:"L" default:"100" help:"Limit results per database"`
	Offset int    `help:"Skip N results"`

	// Path filters
	Include      []string `short:"s" help:"Include paths matching pattern"`
	Exclude      []string `short:"E" help:"Exclude paths matching pattern"`
	Search       []string `help:"Search terms (space-separated for AND, | for OR)"`
	Regex        string   `help:"Filter paths by regex pattern"`
	PathContains []string `help:"Path must contain all these strings"`

	// Size/Duration filters
	Size        []string `short:"S" help:"Size range (e.g., >100MB, 1GB%10)"`
	Duration    []string `short:"d" help:"Duration range (e.g., >1hour, 30min%10)"`
	MinSize     string   `help:"(Deprecated) Minimum file size"`
	MaxSize     string   `help:"(Deprecated) Maximum file size"`
	MinDuration int      `help:"(Deprecated) Minimum duration in seconds"`
	MaxDuration int      `help:"(Deprecated) Maximum duration in seconds"`
	DurationFromSize string `help:"Constrain media to duration of videos which match any size constraints"`
	Ext         []string `short:"e" help:"Filter by extensions (e.g., .mp4,.mkv)"`

	// Time filters
	CreatedAfter  string `help:"Created after date (YYYY-MM-DD)"`
	CreatedBefore string `help:"Created before date (YYYY-MM-DD)"`
	ModifiedAfter string `help:"Modified after date (YYYY-MM-DD)"`
	ModifiedBefore string `help:"Modified before date (YYYY-MM-DD)"`
	DeletedAfter  string `help:"Deleted after date (YYYY-MM-DD)"`
	DeletedBefore string `help:"Deleted before date (YYYY-MM-DD)"`
	PlayedAfter   string `help:"Last played after date (YYYY-MM-DD)"`
	PlayedBefore  string `help:"Last played before date (YYYY-MM-DD)"`

	// Playback state filters
	Watched      *bool `help:"Filter by watched status (true/false)"`
	Unfinished   bool  `help:"Has playhead but not finished"`
	Partial      string `short:"P" help:"Filter by partial playback status"`
	PlayCountMin int   `help:"Minimum play count"`
	PlayCountMax int   `help:"Maximum play count"`

	// Content type filters
	VideoOnly bool `help:"Only video files"`
	AudioOnly bool `help:"Only audio files"`
	ImageOnly bool `help:"Only image files"`
	Portrait  bool `help:"Only portrait orientation files"`
	ScanSubtitles   bool `help:"Scan for external subtitles during import"`
	OnlineMediaOnly bool `help:"Exclude local media"`
	LocalMediaOnly  bool `help:"Exclude online media"`

	MimeType   []string `help:"Filter by mimetype substring (e.g., video, mp4)"`
	NoMimeType []string `help:"Exclude by mimetype substring"`

	// Deleted status
	HideDeleted bool `default:"true" help:"Exclude deleted files from results"`
	OnlyDeleted bool `help:"Include only deleted files in results"`

	// Siblings
	FetchSiblings    string `short:"o" help:"Fetch siblings of matched files (each, all, if-audiobook)"`
	FetchSiblingsMax int    `help:"Maximum number of siblings to fetch"`

	// Ordering
	PlayInOrder   string `short:"O" default:"natural_ps" help:"Play media in order"`
	NoPlayInOrder bool   `help:"Don't play media in order"`

	// Sorting
	SortBy  string `short:"u" default:"path" help:"Sort by field"`
	Reverse bool   `short:"V" help:"Reverse sort order"`
	NatSort bool   `short:"n" help:"Use natural sorting"`
	Random  bool   `short:"r" help:"Random order"`

	// Display
	Columns   []string `short:"c" help:"Columns to display"`
	BigDirs   bool     `short:"B" help:"Aggregate by parent directory"`
		JSON      bool     `short:"j" help:"Output results as JSON"`
		Summarize bool     `help:"Print aggregate statistics"`
		Frequency string   `short:"f" help:"Group statistics by time frequency (daily, weekly, monthly, yearly)"`
		TUI       bool     `help:"Interactive TUI mode"`
	

	// Grouping
	FileCounts        string   `help:"Filter by number of files in directory (e.g., >5, 10%1)"`
	GroupByExtensions bool     `help:"Group by file extensions"`
	GroupByMimeTypes  bool     `help:"Group by mimetypes"`
	GroupBySize       bool     `help:"Group by size buckets"`
	Depth             int      `short:"D" help:"Aggregate at specific directory depth"`
	MinDepth          int      `default:"0" help:"Minimum depth for aggregation"`
	MaxDepth          int      `help:"Maximum depth for aggregation"`
	Parents           bool     `help:"Include parent directories in aggregation"`
	FoldersOnly       bool     `help:"Only show folders"`
	FilesOnly         bool     `help:"Only show files"`
	FolderSizes       []string `help:"Filter folders by total size"`
	FolderCounts      string   `help:"Filter folders by number of subfolders"`

	// Text processing and sorting (from regex_sort.py)
	RegexSort  bool     `help:"Sort by splitting lines and sorting words" alias:"rs"`
	Regexs     []string `help:"Regex patterns for line splitting" alias:"re"`
	WordSorts  []string `help:"Word sorting strategies"`
	LineSorts  []string `help:"Line sorting strategies"`
	Compat     bool     `help:"Use natsort compat mode"`
	Preprocess bool     `default:"true" help:"Remove junk common to filenames and URLs"`
	StopWords  []string `help:"List of words to ignore"`
	Duplicates *bool    `help:"Filter for duplicate words (true/false)"`
	UniqueOnly *bool    `help:"Filter for unique words (true/false)"`

	// Similarity clustering
	Similar         bool    `help:"Find similar files or folders"`
	SizesDelta      float64 `default:"10.0" help:"Size difference threshold (%)"`
	CountsDelta     float64 `default:"3.0" help:"File count difference threshold (%)"`
	DurationsDelta  float64 `default:"5.0" help:"Duration difference threshold (%)"`
	FilterNames     bool    `help:"Cluster by name similarity"`
	FilterSizes     bool    `help:"Cluster by size similarity"`
	FilterCounts    bool    `help:"Cluster by count similarity"`
	FilterDurations bool    `help:"Cluster by duration similarity"`
	TotalSizes      bool    `help:"Compare total sizes (folders only)"`
	TotalDurations  bool    `help:"Compare total durations (folders only)"`
	OnlyDuplicates  bool    `help:"Only show duplicate items"`
		OnlyOriginals   bool    `help:"Only show original items"`
	
		// Clustering
		ClusterSort     bool    `short:"C" help:"Group items by similarity"`
		Clusters        int     `help:"Number of clusters"`
		TFIDF           bool    `help:"Use TF-IDF for clustering"`
		MoveGroups      bool    `help:"Move grouped files into separate directories"`
		PrintGroups     bool    `help:"Print clusters as JSON"`
	
		// Database merging and filtering
	OnlyTables        []string `short:"t" help:"Comma separated specific table(s)"`
	PrimaryKeys       []string `help:"Comma separated primary keys"`
	BusinessKeys      []string `help:"Comma separated business keys"`
	Upsert            bool     `help:"Upsert rows on conflict"`
	Ignore            bool     `help:"Ignore rows on conflict (only-new-rows)"`
	OnlyNewRows       bool     `kong:"-"` // Alias for Ignore
	OnlyTargetColumns bool     `help:"Only copy columns that exist in target"`
	SkipColumns       []string `help:"Columns to skip during merge"`
	Where             []string `short:"w" help:"SQL where clause(s)"`
	Exact             bool     `help:"Exact match for search"`
	FlexibleSearch    bool     `help:"Flexible search (fuzzy)"`

	// Actions
	PostAction   string `help:"Post-action: none, delete, mark-deleted, move, copy"`
	DeleteFiles  bool   `help:"Delete files after action"`
	DeleteRows   bool   `help:"Delete rows from database"`
	MarkDeleted  bool   `help:"Mark as deleted in database"`
	MoveTo       string `help:"Move files to directory"`
	CopyTo       string `help:"Copy files to directory"`
	ActionLimit  int    `help:"Stop after N files"`
	ActionSize   string `help:"Stop after N bytes (e.g., 10GB)"`
	Exists       bool   `help:"Filter out non-existent files"`
	TrackHistory bool   `default:"true" help:"Track playback history"`

	// FTS options
	FTS      bool   `help:"Use full-text search if available"`
	FTSTable string `default:"media_fts" help:"FTS table name"`
	Related  int    `short:"R" help:"Find media related to the first result"`
	Verbose  bool   `short:"v" help:"Enable verbose logging"`

	// Common options from arggroups.py
	Simulate       bool    `help:"Dry run; don't actually do anything"`
	DryRun         bool    `kong:"-"` // Alias for Simulate
	NoConfirm      bool    `short:"y" help:"Don't ask for confirmation"`
	Yes            bool    `kong:"-"` // Alias for NoConfirm
	Timeout        string  `short:"T" help:"Quit after N minutes/seconds"`
	Threads        int     `help:"Use N threads for parallel processing"`
	Loop           bool    `help:"Loop playback"`
		Mute         bool    `short:"M" help:"Start playback muted"`
		OverridePlayer string  `help:"Override default player (e.g. --player 'vlc')"`
		IgnoreErrors   bool    `short:"i" help:"Ignore errors and continue to next file"`
		Completed      bool    `help:"Show only completed items"`
		InProgress     bool    `help:"Show only items in progress"`
		Start          string  `help:"Start playback at specific time/percentage"`
		End            string  `help:"Stop playback at specific time/percentage"`
	Volume         int     `help:"Set initial volume (0-100)"`
	Fullscreen     bool    `help:"Start in fullscreen"`
	NoSubtitles    bool    `help:"Disable subtitles"`
	Speed          float64 `default:"1.0" help:"Playback speed"`
	SavePlayhead   bool    `default:"true" help:"Save playback position on quit"`
	MpvSocket      string  `help:"Mpv socket path"`
	WatchLaterDir  string  `help:"Mpv watch_later directory"`
}

func (g *GlobalFlags) AfterApply() error {
	if g.Simulate {
		g.DryRun = true
	}
	if g.NoConfirm {
		g.Yes = true
	}
	if g.Ignore {
		g.OnlyNewRows = true
	}
	if g.Ext != nil {
		for i, ext := range g.Ext {
			if !strings.HasPrefix(ext, ".") {
				g.Ext[i] = "." + ext
			}
		}
	}
	return nil
}

var LogLevel = &slog.LevelVar{}

func SetupLogging(verbose bool) {
	if verbose {
		LogLevel.Set(slog.LevelDebug)
	} else {
		LogLevel.Set(slog.LevelInfo)
	}
}
