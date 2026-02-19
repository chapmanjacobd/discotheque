package commands

import (
	"github.com/alecthomas/kong"
	"github.com/chapmanjacobd/discotheque/internal/models"
)

type BigDirsCmd struct {
	models.GlobalFlags
	Databases []string `arg:"" required:"" help:"SQLite database files" type:"existingfile"`
}

func (c *BigDirsCmd) Run(ctx *kong.Context) error {
	// Bigdirs is Essentially Print with BigDirs enabled by default
	c.BigDirs = true
	printCmd := PrintCmd{GlobalFlags: c.GlobalFlags, Databases: c.Databases}
	return printCmd.Run(ctx)
}
