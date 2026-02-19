package main

import (
	"log/slog"
	"os"

	"github.com/alecthomas/kong"
	"github.com/chapmanjacobd/discotheque/internal/commands"
	"github.com/chapmanjacobd/discotheque/internal/models"
	_ "github.com/mattn/go-sqlite3"
)

// CLI defines the command-line interface
type CLI struct {
	Add            commands.AddCmd            `cmd:"" help:"Add media to database"`
	Check          commands.CheckCmd          `cmd:"" help:"Check for missing files and mark as deleted"`
	Print          commands.PrintCmd          `cmd:"" help:"Print media information"`
	Search         commands.SearchCmd         `cmd:"" help:"Search media using FTS"`
	SearchCaptions commands.SearchCaptionsCmd `cmd:"" help:"Search captions using FTS" aliases:"sc"`
	Playlists      commands.PlaylistsCmd      `cmd:"" help:"List scan roots (playlists)"`
	SearchDB       commands.SearchDBCmd       `cmd:"" help:"Search arbitrary database table" aliases:"sdb"`
	MediaCheck     commands.MediaCheckCmd     `cmd:"" help:"Check media files for corruption" aliases:"mc"`
	FilesInfo      commands.FilesInfoCmd      `cmd:"" help:"Show information about files" aliases:"fs"`
	DiskUsage      commands.DiskUsageCmd      `cmd:"" help:"Show disk usage aggregation" aliases:"du"`
	BigDirs        commands.BigDirsCmd        `cmd:"" help:"Show big directories aggregation" aliases:"bigdirs,bd"`
	SimilarFiles   commands.SimilarFilesCmd   `cmd:"" help:"Find similar files" aliases:"sf"`
	SimilarFolders commands.SimilarFoldersCmd `cmd:"" help:"Find similar folders" aliases:"sh"`
	Watch          commands.WatchCmd          `cmd:"" help:"Watch videos with mpv"`
	Listen         commands.ListenCmd         `cmd:"" help:"Listen to audio with mpv"`
	Stats          commands.StatsCmd          `cmd:"" help:"Show library statistics"`
	History        commands.HistoryCmd        `cmd:"" help:"Show playback history"`
	HistoryAdd     commands.HistoryAddCmd     `cmd:"" help:"Add paths to playback history"`
	Optimize       commands.OptimizeCmd       `cmd:"" help:"Optimize database (VACUUM, ANALYZE, FTS optimize)"`
	Tui            commands.TuiCmd            `cmd:"" help:"Interactive TUI media picker"`
	Readme         commands.ReadmeCmd         `cmd:"" help:"Generate README.md content"`
	RegexSort      commands.RegexSortCmd      `cmd:"" help:"Sort by splitting lines and sorting words" aliases:"rs"`
	ClusterSort    commands.ClusterSortCmd    `cmd:"" help:"Group items by similarity" aliases:"cs"`
	Open           commands.OpenCmd           `cmd:"" help:"Open files with default application"`
	Browse         commands.BrowseCmd         `cmd:"" help:"Open URLs in browser"`
	Now            commands.NowCmd            `cmd:"" help:"Show current mpv playback status"`
	Next           commands.NextCmd           `cmd:"" help:"Skip to next file in mpv"`
	Stop           commands.StopCmd           `cmd:"" help:"Stop mpv playback"`
	Pause          commands.PauseCmd          `cmd:"" help:"Toggle mpv pause state" aliases:"play"`
	Seek           commands.SeekCmd           `cmd:"" help:"Seek mpv playback" aliases:"ffwd,rewind"`
	MergeDBs       commands.MergeDBsCmd       `cmd:"merge-dbs" help:"Merge multiple SQLite databases" aliases:"mergedbs"`

	ExitCalled bool `kong:"-"`
}

func (c *CLI) Terminate(code int) {
	c.ExitCalled = true
}

func main() {
	cli := &CLI{}
	ctx := kong.Parse(cli,
		kong.Name("disco"),
		kong.Description("discotheque management tool"),
		kong.UsageOnError(),
	)

	// Configure logger
	opts := &slog.HandlerOptions{
		Level: models.LogLevel,
	}
	logger := slog.New(slog.NewTextHandler(os.Stderr, opts))
	slog.SetDefault(logger)

	err := ctx.Run(ctx)
	ctx.FatalIfErrorf(err)
}
