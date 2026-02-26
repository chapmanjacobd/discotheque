package commands

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/alecthomas/kong"
	"github.com/chapmanjacobd/discotheque/internal/db"
	"github.com/chapmanjacobd/discotheque/internal/metadata"
	"github.com/chapmanjacobd/discotheque/internal/models"
	"github.com/chapmanjacobd/discotheque/internal/query"
	"github.com/chapmanjacobd/discotheque/internal/utils"
)

type MediaCheckCmd struct {
	models.GlobalFlags
	Databases []string `arg:"" required:"" help:"SQLite database files" type:"existingfile"`

	ChunkSize         float64 `help:"Chunk size in seconds. If set, recommended to use >0.1 seconds" default:"0.5"`
	Gap               string  `help:"Width between chunks to skip. Values greater than 1 are treated as number of seconds" default:"5%"`
	DeleteCorrupt     string  `help:"Delete media that is more corrupt or equal to this threshold. Values greater than 1 are treated as number of seconds"`
	FullScanIfCorrupt string  `help:"Full scan as second pass if initial scan result more corruption or equal to this threshold. Values greater than 1 are treated as number of seconds"`
	FullScan          bool    `help:"Decode the full media file"`
	AudioScan         bool    `help:"Count errors in audio track only"`
}

func (c MediaCheckCmd) IsFilterTrait()      {}
func (c MediaCheckCmd) IsPathFilterTrait()  {}
func (c MediaCheckCmd) IsMediaFilterTrait() {}
func (c MediaCheckCmd) IsDeletedTrait()     {}

func (c *MediaCheckCmd) Run(ctx *kong.Context) error {
	models.SetupLogging(c.Verbose)

	media, err := query.MediaQuery(context.Background(), c.Databases, c.GlobalFlags)
	if err != nil {
		return err
	}
	media = query.FilterMedia(media, c.GlobalFlags)

	if len(media) == 0 {
		return fmt.Errorf("no media found")
	}

	gap, _ := utils.FloatFromPercent(c.Gap)
	deleteThreshold, _ := utils.FloatFromPercent(c.DeleteCorrupt)
	fullScanThreshold, _ := utils.FloatFromPercent(c.FullScanIfCorrupt)

	for _, m := range media {
		corruption := 0.0
		duration := 0.0
		if m.Duration != nil {
			duration = float64(*m.Duration)
		}

		if c.FullScan {
			corruption, err = metadata.DecodeFullScan(context.Background(), m.Path)
			if err != nil {
				slog.Error("Full scan failed", "path", m.Path, "error", err)
				corruption = 0.5
			}
		} else {
			if duration == 0 {
				corruption = 0.5
			} else {
				scans := utils.CalculateSegments(duration, c.ChunkSize, gap)
				corruption = metadata.DecodeQuickScan(context.Background(), m.Path, scans, c.ChunkSize)

				if fullScanThreshold > 0 && corruption >= fullScanThreshold {
					slog.Info("Corruption threshold reached, performing full scan", "path", m.Path, "corruption", corruption)
					corruption, err = metadata.DecodeFullScan(context.Background(), m.Path)
					if err != nil {
						slog.Error("Full scan failed", "path", m.Path, "error", err)
					}
				}
			}
		}

		fmt.Printf("%.2f%%\t%s\n", corruption*100, m.Path)

		if deleteThreshold > 0 && corruption >= deleteThreshold {
			slog.Warn("Deleting corrupt file", "path", m.Path, "corruption", corruption)
			if !c.DryRun {
				if err := os.Remove(m.Path); err != nil {
					slog.Error("Failed to delete corrupt file", "path", m.Path, "error", err)
				} else {
					// Mark as deleted in DB
					sqlDB, err := db.Connect(m.DB)
					if err == nil {
						defer sqlDB.Close()
						queries := db.New(sqlDB)
						queries.MarkDeleted(context.Background(), db.MarkDeletedParams{
							Path:        m.Path,
							TimeDeleted: utils.ToNullInt64(time.Now().Unix()),
						})
					}
				}
			}
		}
	}

	return nil
}
