package main

import (
	"log/slog"
	"os"

	"github.com/alecthomas/kong"
	"github.com/chapmanjacobd/discotheque/internal/models"
	"github.com/chapmanjacobd/discotheque/internal/utils"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	cli := &CLI{}
	parser, err := kong.New(cli,
		kong.Name("disco"),
		kong.Description("discotheque management tool"),
		kong.UsageOnError(),
	)
	if err != nil {
		panic(err)
	}

	// Dynamic help filtering: Hide flag groups not implemented by the subcommand
	var filterFlags func(*kong.Node)
	filterFlags = func(n *kong.Node) {
		if n.Type == kong.CommandNode {
			// For each flag in the command
			for _, flag := range n.Flags {
				if flag.Group == nil {
					continue
				}

				keep := false
				target := n.Target.Interface()
				switch flag.Group.Title {
				case "Query":
					_, keep = target.(models.QueryTrait)
				case "Filter":
					_, keep = target.(models.FilterTrait)
				case "Time":
					_, keep = target.(models.TimeTrait)
				case "MediaFilter":
					_, keep = target.(models.MediaFilterTrait)
				case "PathFilter":
					_, keep = target.(models.PathFilterTrait)
				case "Deleted":
					_, keep = target.(models.DeletedTrait)
				case "Aggregate":
					_, keep = target.(models.AggregateTrait)
				case "PostAction":
					_, keep = target.(models.PostActionTrait)
				case "MpvAction":
					_, keep = target.(models.MpvActionTrait)
				case "Sort":
					_, keep = target.(models.SortTrait)
				case "Display":
					_, keep = target.(models.DisplayTrait)
				case "Playback":
					_, keep = target.(models.PlaybackTrait)
				case "Text":
					_, keep = target.(models.TextTrait)
				case "Similarity":
					_, keep = target.(models.SimilarityTrait)
				case "Merge":
					_, keep = target.(models.MergeTrait)
				case "Action":
					_, keep = target.(models.ActionTrait)
				case "FTS":
					_, keep = target.(models.FTSTrait)
				case "Hashing":
					_, keep = target.(models.HashingTrait)
				case "Dedupe":
					_, keep = target.(models.DedupeTrait)
				case "History":
					_, keep = target.(models.HistoryTrait)
				case "Syncweb":
					_, keep = target.(models.SyncwebTrait)
				default:
					keep = true
				}

				if !keep {
					flag.Hidden = true
				}
			}
		}
		for _, child := range n.Children {
			filterFlags(child)
		}
	}
	filterFlags(parser.Model.Node)

	ctx, err := parser.Parse(os.Args[1:])
	if err != nil {
		parser.FatalIfErrorf(err)
	}

	// Configure logger
	logger := slog.New(&utils.PlainHandler{
		Level: models.LogLevel,
		Out:   os.Stderr,
	})
	slog.SetDefault(logger)

	err = ctx.Run(ctx)
	ctx.FatalIfErrorf(err)
}
