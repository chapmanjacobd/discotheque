package models

import (
	"log/slog"
	"strings"
)

// Trait interfaces for dynamic help filtering
type (
	QueryTrait      interface{ IsQueryTrait() }
	FilterTrait     interface{ IsFilterTrait() }
	SortTrait       interface{ IsSortTrait() }
	DisplayTrait    interface{ IsDisplayTrait() }
	PlaybackTrait   interface{ IsPlaybackTrait() }
	TextTrait       interface{ IsTextTrait() }
	SimilarityTrait interface{ IsSimilarityTrait() }
	MergeTrait      interface{ IsMergeTrait() }
	ActionTrait     interface{ IsActionTrait() }
	FTSTrait        interface{ IsFTSTrait() }
	HashingTrait    interface{ IsHashingTrait() }
	DedupeTrait     interface{ IsDedupeTrait() }
	HistoryTrait    interface{ IsHistoryTrait() }
)

// GlobalFlags are flags available to core data commands (print, search, du, etc)
type GlobalFlags struct {
	Query  string `short:"q" help:"Raw SQL query (overrides all query building)" group:"Query"`
	Limit  int    `short:"L" default:"100" help:"Limit results per database" group:"Query"`
	All    bool   `short:"a" help:"Return all results (no limit)" group:"Query"`
	Offset int    `help:"Skip N results" group:"Query"`

	// Dedupe profiles
	Audio        bool `help:"Dedupe database by artist + album + title" group:"Dedupe"`
	ExtractorID  bool `alias:"id" help:"Dedupe database by extractor_id" group:"Dedupe"`
	TitleOnly    bool `help:"Dedupe database by title" group:"Dedupe"`
	DurationOnly bool `help:"Dedupe database by duration" group:"Dedupe"`
	Filesystem   bool `alias:"fs" help:"Dedupe filesystem database (hash)" group:"Dedupe"`

	// Dedupe options
	CompareDirs        bool    `help:"Compare directories" group:"Dedupe"`
	Basename           bool    `help:"Match by basename similarity" group:"Dedupe"`
	Dirname            bool    `help:"Match by dirname similarity" group:"Dedupe"`
	MinSimilarityRatio float64 `default:"0.8" help:"Filter out matches with less than this ratio (0.7-0.9)" group:"Dedupe"`
	DedupeCmd          string  `help:"Command to run for deduplication (rmlint-style: cmd duplicate keep)" group:"Dedupe"`

	// Path filters
	Include      []string `short:"s" help:"Include paths matching pattern" group:"Filter"`
	Exclude      []string `short:"E" help:"Exclude paths matching pattern" group:"Filter"`
	Search       []string `help:"Search terms (space-separated for AND, | for OR)" group:"Filter"`
	Category     []string `help:"Filter by category" group:"Filter"`
	Genre        string   `help:"Filter by genre" group:"Filter"`
	Regex        string   `help:"Filter paths by regex pattern" group:"Filter"`
	PathContains []string `help:"Path must contain all these strings" group:"Filter"`
	Paths        []string `help:"Exact paths to include" group:"Filter"`

	// Size/Duration filters
	Size             []string `short:"S" help:"Size range (e.g., >100MB, 1GB%10)" group:"Filter"`
	Duration         []string `short:"d" help:"Duration range (e.g., >1hour, 30min%10)" group:"Filter"`
	DurationFromSize string   `help:"Constrain media to duration of videos which match any size constraints" group:"Filter"`
	Ext              []string `short:"e" help:"Filter by extensions (e.g., .mp4,.mkv)" group:"Filter"`

	// Time filters
	CreatedAfter   string `help:"Created after date (YYYY-MM-DD)" group:"Filter"`
	CreatedBefore  string `help:"Created before date (YYYY-MM-DD)" group:"Filter"`
	ModifiedAfter  string `help:"Modified after date (YYYY-MM-DD)" group:"Filter"`
	ModifiedBefore string `help:"Modified before date (YYYY-MM-DD)" group:"Filter"`
	DeletedAfter   string `help:"Deleted after date (YYYY-MM-DD)" group:"Filter"`
	DeletedBefore  string `help:"Deleted before date (YYYY-MM-DD)" group:"Filter"`
	PlayedAfter    string `help:"Last played after date (YYYY-MM-DD)" group:"Filter"`
	PlayedBefore   string `help:"Last played before date (YYYY-MM-DD)" group:"Filter"`

	// Playback state filters
	Watched      *bool  `help:"Filter by watched status (true/false)" group:"Filter"`
	Unfinished   bool   `help:"Has playhead but not finished" group:"Filter"`
	Partial      string `short:"P" help:"Filter by partial playback status" group:"Filter"`
	PlayCountMin int    `help:"Minimum play count" group:"Filter"`
	PlayCountMax int    `help:"Maximum play count" group:"Filter"`
	Completed    bool   `help:"Show only completed items" group:"Filter"`
	InProgress   bool   `help:"Show only items in progress" group:"Filter"`
	WithCaptions bool   `help:"Show only items with captions" group:"Filter"`

	// Content type filters
	VideoOnly       bool     `help:"Only video files" group:"Filter"`
	AudioOnly       bool     `help:"Only audio files" group:"Filter"`
	ImageOnly       bool     `help:"Only image files" group:"Filter"`
	TextOnly        bool     `help:"Only text/ebook files" group:"Filter"`
	Portrait        bool     `help:"Only portrait orientation files" group:"Filter"`
	ScanSubtitles   bool     `help:"Scan for external subtitles during import" group:"Filter"`
	OnlineMediaOnly bool     `help:"Exclude local media" group:"Filter"`
	LocalMediaOnly  bool     `help:"Exclude online media" group:"Filter"`
	FlexibleSearch  bool     `help:"Flexible search (fuzzy)" group:"Filter"`
	Exact           bool     `help:"Exact match for search" group:"Filter"`
	Where           []string `short:"w" help:"SQL where clause(s)" group:"Filter"`
	Exists          bool     `help:"Filter out non-existent files" group:"Filter"`

	MimeType   []string `help:"Filter by mimetype substring (e.g., video, mp4)" group:"Filter"`
	NoMimeType []string `help:"Exclude by mimetype substring" group:"Filter"`

	NoDefaultCategories bool `help:"Disable default categories" group:"Filter"`

	// Deleted status
	HideDeleted bool `default:"true" help:"Exclude deleted files from results" group:"Filter"`
	OnlyDeleted bool `help:"Include only deleted files in results" group:"Filter"`

	// Siblings
	FetchSiblings    string `short:"o" help:"Fetch siblings of matched files (each, all, if-audiobook)" group:"Filter"`
	FetchSiblingsMax int    `help:"Maximum number of siblings to fetch" group:"Filter"`

	// Sorting
	SortBy  string `short:"u" default:"path" help:"Sort by field" group:"Sort"`
	Reverse bool   `short:"V" help:"Reverse sort order" group:"Sort"`
	NatSort bool   `short:"n" help:"Use natural sorting" group:"Sort"`
	Random  bool   `short:"r" help:"Random order" group:"Sort"`

	// Display
	Columns   []string `short:"c" help:"Columns to display" group:"Display"`
	BigDirs   bool     `short:"B" help:"Aggregate by parent directory" group:"Display"`
	JSON      bool     `short:"j" help:"Output results as JSON" group:"Display"`
	Summarize bool     `help:"Print aggregate statistics" group:"Display"`
	Frequency string   `short:"f" help:"Group statistics by time frequency (daily, weekly, monthly, yearly)" group:"Display"`
	TUI       bool     `help:"Interactive TUI mode" group:"Display"`

	// Grouping
	FileCounts        string   `help:"Filter by number of files in directory (e.g., >5, 10%1)" group:"Display"`
	GroupByExtensions bool     `help:"Group by file extensions" group:"Display"`
	GroupByMimeTypes  bool     `help:"Group by mimetypes" group:"Display"`
	GroupBySize       bool     `help:"Group by size buckets" group:"Display"`
	Depth             int      `short:"D" help:"Aggregate at specific directory depth" group:"Display"`
	MinDepth          int      `default:"0" help:"Minimum depth for aggregation" group:"Display"`
	MaxDepth          int      `help:"Maximum depth for aggregation" group:"Display"`
	Parents           bool     `help:"Include parent directories in aggregation" group:"Display"`
	FoldersOnly       bool     `help:"Only show folders" group:"Display"`
	FilesOnly         bool     `help:"Only show files" group:"Display"`
	FolderSizes       []string `help:"Filter folders by total size" group:"Display"`
	FolderCounts      string   `help:"Filter folders by number of subfolders" group:"Display"`

	// Text processing and sorting (from regex_sort.py)
	RegexSort  bool     `help:"Sort by splitting lines and sorting words" alias:"rs" group:"Text"`
	Regexs     []string `help:"Regex patterns for line splitting" alias:"re" group:"Text"`
	WordSorts  []string `help:"Word sorting strategies" group:"Text"`
	LineSorts  []string `help:"Line sorting strategies" group:"Text"`
	Compat     bool     `help:"Use natsort compat mode" group:"Text"`
	Preprocess bool     `default:"true" help:"Remove junk common to filenames and URLs" group:"Text"`
	StopWords  []string `help:"List of words to ignore" group:"Text"`
	Duplicates *bool    `help:"Filter for duplicate words (true/false)" group:"Text"`
	UniqueOnly *bool    `help:"Filter for unique words (true/false)" group:"Text"`

	// Similarity clustering
	Similar         bool    `help:"Find similar files or folders" group:"Similarity"`
	SizesDelta      float64 `default:"10.0" help:"Size difference threshold (%)" group:"Similarity"`
	CountsDelta     float64 `default:"3.0" help:"File count difference threshold (%)" group:"Similarity"`
	DurationsDelta  float64 `default:"5.0" help:"Duration difference threshold (%)" group:"Similarity"`
	FilterNames     bool    `help:"Cluster by name similarity" group:"Similarity"`
	FilterSizes     bool    `help:"Cluster by size similarity" group:"Similarity"`
	FilterCounts    bool    `help:"Cluster by count similarity" group:"Similarity"`
	FilterDurations bool    `help:"Cluster by duration similarity" group:"Similarity"`
	TotalSizes      bool    `help:"Compare total sizes (folders only)" group:"Similarity"`
	TotalDurations  bool    `help:"Compare total durations (folders only)" group:"Similarity"`
	OnlyDuplicates  bool    `help:"Only show duplicate items" group:"Similarity"`
	OnlyOriginals   bool    `help:"Only show original items" group:"Similarity"`

	// Clustering
	ClusterSort bool `short:"C" help:"Group items by similarity" group:"Similarity"`
	Clusters    int  `help:"Number of clusters" group:"Similarity"`
	TFIDF       bool `help:"Use TF-IDF for clustering" group:"Similarity"`
	MoveGroups  bool `help:"Move grouped files into separate directories" group:"Similarity"`
	PrintGroups bool `help:"Print clusters as JSON" group:"Similarity"`

	// Sorting Extensions
	ReRank string `short:"k" alias:"rerank" help:"Add key/value pairs re-rank sorting by multiple attributes (COLUMN=WEIGHT)" group:"Sort"`

	// FTS options
	FTS      bool   `help:"Use full-text search if available" group:"FTS"`
	FTSTable string `default:"media_fts" help:"FTS table name" group:"FTS"`
	Related  int    `short:"R" help:"Find media related to the first result" group:"FTS"`

	// Playback
	PlayInOrder           string  `short:"O" default:"natural_ps" help:"Play media in order" group:"Playback"`
	NoPlayInOrder         bool    `help:"Don't play media in order" group:"Playback"`
	Loop                  bool    `help:"Loop playback" group:"Playback"`
	Mute                  bool    `short:"M" help:"Start playback muted" group:"Playback"`
	OverridePlayer        string  `help:"Override default player (e.g. --player 'vlc')" group:"Playback"`
	Start                 string  `help:"Start playback at specific time/percentage" group:"Playback"`
	End                   string  `help:"Stop playback at specific time/percentage" group:"Playback"`
	Volume                int     `help:"Set initial volume (0-100)" group:"Playback"`
	Fullscreen            bool    `help:"Start in fullscreen" group:"Playback"`
	NoSubtitles           bool    `help:"Disable subtitles" group:"Playback"`
	SubtitleMix           float64 `default:"0.35" help:"Probability to play no-subtitle content" group:"Playback"`
	InterdimensionalCable int     `short:"4" alias:"4dtv" help:"Duration to play (in seconds) while changing the channel" group:"Playback"`
	Speed                 float64 `default:"1.0" help:"Playback speed" group:"Playback"`
	SavePlayhead          bool    `default:"true" help:"Save playback position on quit" group:"Playback"`
	MpvSocket             string  `help:"Mpv socket path" group:"Playback"`
	WatchLaterDir         string  `help:"Mpv watch_later directory" group:"Playback"`

	PlayerArgsSub   []string `help:"Player arguments for videos with subtitles" group:"Playback"`
	PlayerArgsNoSub []string `help:"Player arguments for videos without subtitles" group:"Playback"`

	Cmd0   string `help:"Command to run if mpv exits with code 0" group:"Action"`
	Cmd1   string `help:"Command to run if mpv exits with code 1" group:"Action"`
	Cmd2   string `help:"Command to run if mpv exits with code 2" group:"Action"`
	Cmd3   string `help:"Command to run if mpv exits with code 3" group:"Action"`
	Cmd4   string `help:"Command to run if mpv exits with code 4" group:"Action"`
	Cmd5   string `help:"Command to run if mpv exits with code 5" group:"Action"`
	Cmd6   string `help:"Command to run if mpv exits with code 6" group:"Action"`
	Cmd7   string `help:"Command to run if mpv exits with code 7" group:"Action"`
	Cmd8   string `help:"Command to run if mpv exits with code 8" group:"Action"`
	Cmd9   string `help:"Command to run if mpv exits with code 9" group:"Action"`
	Cmd10  string `help:"Command to run if mpv exits with code 10" group:"Action"`
	Cmd11  string `help:"Command to run if mpv exits with code 11" group:"Action"`
	Cmd12  string `help:"Command to run if mpv exits with code 12" group:"Action"`
	Cmd13  string `help:"Command to run if mpv exits with code 13" group:"Action"`
	Cmd14  string `help:"Command to run if mpv exits with code 14" group:"Action"`
	Cmd15  string `help:"Command to run if mpv exits with code 15" group:"Action"`
	Cmd20  string `help:"Command to run if mpv exits with code 20" group:"Action"`
	Cmd127 string `help:"Command to run if mpv exits with code 127" group:"Action"`

	Interactive bool `short:"I" help:"Interactive decision making after playback" group:"Action"`
	Trash       bool `help:"Trash files after action" group:"Action"`

	// Hashing
	HashGap       float64 `default:"0.1" help:"Gap between segments (0.0-1.0 as percentage of file size, or absolute bytes if >1)" group:"Hashing"`
	HashChunkSize int64   `help:"Size of each segment to hash" group:"Hashing"`
	HashThreads   int     `default:"1" help:"Number of threads to use for hashing a single file" group:"Hashing"`

	// Syncweb
	SyncwebURL      string `help:"Syncweb/Syncthing API URL" group:"Syncweb" env:"SYNCWEB_URL"`
	SyncwebAPIKey   string `help:"Syncweb/Syncthing API Key" group:"Syncweb" env:"SYNCWEB_API_KEY"`
	SyncwebHome     string `help:"Syncweb home directory" group:"Syncweb" env:"SYNCWEB_HOME"`
	SyncwebPublic_  string `kong:"-" env:"SYNCWEB_PUBLIC"`
	SyncwebPrivate_ string `kong:"-" env:"SYNCWEB_PRIVATE"`

	// Chromecast
	Cast          bool   `help:"Cast to chromecast groups" group:"Playback"`
	CastDevice    string `alias:"cast-to" help:"Chromecast device name" group:"Playback"`
	CastWithLocal bool   `help:"Play music locally at the same time as chromecast" group:"Playback"`

	// Database merging and filtering
	OnlyTables        []string `short:"t" help:"Comma separated specific table(s)" group:"Merge"`
	PrimaryKeys       []string `help:"Comma separated primary keys" group:"Merge"`
	BusinessKeys      []string `help:"Comma separated business keys" group:"Merge"`
	Upsert            bool     `help:"Upsert rows on conflict" group:"Merge"`
	Ignore            bool     `help:"Ignore rows on conflict (only-new-rows)" group:"Merge"`
	OnlyNewRows       bool     `kong:"-"` // Alias for Ignore
	OnlyTargetColumns bool     `help:"Only copy columns that exist in target" group:"Merge"`
	SkipColumns       []string `help:"Columns to skip during merge" group:"Merge"`

	// Actions
	PostAction   string `help:"Post-action: none, delete, mark-deleted, move, copy" group:"Action"`
	DeleteFiles  bool   `help:"Delete files after action" group:"Action"`
	DeleteRows   bool   `help:"Delete rows from database" group:"Action"`
	MarkDeleted  bool   `help:"Mark as deleted in database" group:"Action"`
	MoveTo       string `help:"Move files to directory" group:"Action"`
	CopyTo       string `help:"Copy files to directory" group:"Action"`
	ActionLimit  int    `help:"Stop after N files" group:"Action"`
	ActionSize   string `help:"Stop after N bytes (e.g., 10GB)" group:"Action"`
	TrackHistory bool   `default:"true" help:"Track playback history" group:"Action"`

	// Common options
	Verbose      bool   `short:"v" help:"Enable verbose logging"`
	Simulate     bool   `help:"Dry run; don't actually do anything"`
	DryRun       bool   `kong:"-"` // Alias for Simulate
	NoConfirm    bool   `short:"y" help:"Don't ask for confirmation"`
	Yes          bool   `kong:"-"` // Alias for NoConfirm
	Timeout      string `short:"T" help:"Quit after N minutes/seconds"`
	Threads      int    `help:"Use N threads for parallel processing"`
	IgnoreErrors bool   `short:"i" help:"Ignore errors and continue to next file"`
}

// ControlFlags are a subset of flags for simple control commands
type ControlFlags struct {
	MpvSocket  string `help:"Mpv socket path" group:"Playback"`
	CastDevice string `alias:"cast-to" help:"Chromecast device name" group:"Playback"`
	Verbose    bool   `short:"v" help:"Enable verbose logging"`
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
