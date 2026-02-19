package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/alecthomas/kong"
	"github.com/chapmanjacobd/discotheque/internal/metadata"
	"github.com/chapmanjacobd/discotheque/internal/models"
)

type FilesInfoCmd struct {
	models.GlobalFlags
	Paths []string `arg:"" required:"" help:"Files or directories to scan"`
}

func (c *FilesInfoCmd) Run(ctx *kong.Context) error {
	models.SetupLogging(c.Verbose)

	var allMeta []*metadata.MediaMetadata
	for _, root := range c.Paths {
		err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				return nil
			}

			meta, err := metadata.Extract(context.Background(), path, c.ScanSubtitles)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error extracting metadata for %s: %v\n", path, err)
				return nil
			}
			allMeta = append(allMeta, meta)
			return nil
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error walking %s: %v\n", root, err)
		}
	}

	if c.JSON {
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(allMeta)
	}


	// Basic print if not JSON
	for _, m := range allMeta {
		fmt.Printf("Path: %s\n", m.Media.Path)
		if m.Media.Title.Valid {
			fmt.Printf("  Title: %s\n", m.Media.Title.String)
		}
		if m.Media.Type.Valid {
			fmt.Printf("  Type: %s\n", m.Media.Type.String)
		}
		if m.Media.Size.Valid {
			fmt.Printf("  Size: %d\n", m.Media.Size.Int64)
		}
		if m.Media.Duration.Valid {
			fmt.Printf("  Duration: %d\n", m.Media.Duration.Int64)
		}
		fmt.Println()
	}

	return nil
}
