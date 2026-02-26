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
	ctx := kong.Parse(cli,
		kong.Name("disco"),
		kong.Description("discotheque management tool"),
		kong.UsageOnError(),
	)

	// Configure logger
	logger := slog.New(&utils.PlainHandler{
		Level: models.LogLevel,
		Out:   os.Stderr,
	})
	slog.SetDefault(logger)

	err := ctx.Run(ctx)
	ctx.FatalIfErrorf(err)
}
