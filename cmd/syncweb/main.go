//go:build syncweb

package main

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"syscall"

	"github.com/alecthomas/kong"
	"github.com/chapmanjacobd/discotheque/internal/commands"
	"github.com/chapmanjacobd/discotheque/internal/models"
	"github.com/chapmanjacobd/discotheque/internal/syncweb"
	"github.com/chapmanjacobd/discotheque/internal/utils"
	"github.com/sevlyar/go-daemon"
)

type CLI struct {
	models.CoreFlags
	models.SyncwebFlags
	Home string `help:"Base directory for syncweb metadata" env:"SYNCWEB_HOME" name:"home"`

	Create    commands.SyncwebCreateCmd    `cmd:"" help:"Create a syncweb folder" aliases:"init,in,share"`
	Join      commands.SyncwebJoinCmd      `cmd:"" help:"Join syncweb folders/devices" aliases:"import,clone"`
	Accept    commands.SyncwebAcceptCmd    `cmd:"" help:"Add a device to syncweb" aliases:"add"`
	Drop      commands.SyncwebDropCmd      `cmd:"" help:"Remove a device from syncweb" aliases:"remove,reject"`
	Folders   commands.SyncwebFoldersCmd   `cmd:"" help:"List Syncthing folders" aliases:"list-folders,lsf"`
	Devices   commands.SyncwebDevicesCmd   `cmd:"" help:"List Syncthing devices" aliases:"list-devices,lsd"`
	Ls        commands.SyncwebLsCmd        `cmd:"" help:"List files at the current directory level" aliases:"list"`
	Find      commands.SyncwebFindCmd      `cmd:"" help:"Search for files by filename, size, and modified date" aliases:"fd,search"`
	Stat      commands.SyncwebStatCmd      `cmd:"" help:"Display detailed file status information from Syncthing"`
	Sort      commands.SyncwebSortCmd      `cmd:"" help:"Sort Syncthing files by multiple criteria"`
	Download  commands.SyncwebDownloadCmd  `cmd:"" help:"Mark file paths for download/sync" aliases:"dl,upload,unignore,sync"`
	Automatic commands.SyncwebAutomaticCmd `cmd:"" help:"Start syncweb-automatic daemon"`
	Serve     ServeCmd                     `cmd:"" help:"Run Syncweb in foreground"`
	Start     StartCmd                     `cmd:"" help:"Start Syncweb daemon"`
	Stop      StopCmd                      `cmd:"" help:"Stop Syncweb daemon" aliases:"shutdown,quit"`
	Version   commands.SyncwebVersionCmd   `cmd:"" help:"Show Syncweb version"`
}

type ServeCmd struct{}

func (c *ServeCmd) Run(g *commands.SyncwebCmd) error {
	return g.WithSyncweb(func(s *syncweb.Syncweb) error {
		slog.Info("Syncweb serving in foreground", "myID", s.Node.MyID())
		return s.Node.Serve()
	})
}

type StartCmd struct{}

func (c *StartCmd) Run(g *commands.SyncwebCmd) error {
	home := g.SyncwebHome
	if home == "" {
		home = filepath.Join(os.Getenv("HOME"), ".config", "syncweb")
	}

	cntxt := &daemon.Context{
		PidFileName: filepath.Join(home, "syncweb.pid"),
		PidFilePerm: 0o644,
		LogFileName: filepath.Join(home, "syncweb.log"),
		LogFilePerm: 0o640,
		WorkDir:     home,
		Umask:       0o27,
		Args:        []string{"syncweb", "serve", "--home", home},
	}

	d, err := cntxt.Reborn()
	if err != nil {
		return fmt.Errorf("unable to run: %w", err)
	}
	if d != nil {
		slog.Info("Syncweb daemon started", "pid", d.Pid)
		return nil
	}
	defer cntxt.Release()

	slog.Info("Syncweb daemon process starting")
	// The child process continues from here
	// But kong will try to run the command again if we are not careful.
	// Actually, theArgs in daemon.Context is the new command line.
	// So we should just exit here if we are the parent.
	return nil
}

type StopCmd struct{}

func (c *StopCmd) Run(g *commands.SyncwebCmd) error {
	home := g.SyncwebHome
	if home == "" {
		home = filepath.Join(os.Getenv("HOME"), ".config", "syncweb")
	}

	pidFile := filepath.Join(home, "syncweb.pid")
	if _, err := os.Stat(pidFile); os.IsNotExist(err) {
		return fmt.Errorf("syncweb daemon is not running (PID file not found)")
	}

	cntxt := &daemon.Context{
		PidFileName: pidFile,
	}

	d, err := cntxt.Search()
	if err != nil {
		return fmt.Errorf("unable to find daemon process: %w", err)
	}

	if err := d.Signal(syscall.SIGTERM); err != nil {
		return fmt.Errorf("unable to send signal to daemon: %w", err)
	}

	slog.Info("Syncweb daemon stop signal sent")
	return nil
}

func main() {
	cli := &CLI{}
	syncwebCmd := &commands.SyncwebCmd{}

	parser, err := kong.New(cli,
		kong.Name("syncweb"),
		kong.Description("Syncweb: an offline-first distributed web"),
		kong.UsageOnError(),
		kong.Bind(syncwebCmd),
	)
	if err != nil {
		panic(err)
	}

	ctx, err := parser.Parse(os.Args[1:])
	if err != nil {
		parser.FatalIfErrorf(err)
	}

	// Update syncwebCmd with global flags from cli
	syncwebCmd.CoreFlags = cli.CoreFlags
	syncwebCmd.SyncwebFlags = cli.SyncwebFlags
	if cli.Home != "" {
		syncwebCmd.SyncwebHome = cli.Home
	}

	// Configure logger
	models.SetupLogging(cli.Verbose)
	logger := slog.New(&utils.PlainHandler{
		Level: models.LogLevel,
		Out:   os.Stderr,
	})
	slog.SetDefault(logger)

	err = ctx.Run()
	ctx.FatalIfErrorf(err)
}
