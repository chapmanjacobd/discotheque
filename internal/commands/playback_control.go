package commands

import (
	"fmt"
	"os"
	"strings"

	"github.com/alecthomas/kong"
	"github.com/chapmanjacobd/discotheque/internal/models"
	"github.com/chapmanjacobd/discotheque/internal/utils"
)

type MpvControlBase struct {
	models.GlobalFlags
}

func (c MpvControlBase) IsPlaybackTrait() {}

type NowCmd struct {
	MpvControlBase
}

func (c *NowCmd) Run(ctx *kong.Context) error {
	socketPath := c.MpvSocket
	if socketPath == "" {
		socketPath = utils.GetMpvWatchSocket()
	}

	pathResp, err := utils.MpvCall(socketPath, "get_property", "path")
	if err != nil {
		return fmt.Errorf("mpv not running or socket not found: %w", err)
	}

	path := utils.GetString(pathResp.Data)
	fmt.Printf("Now Playing: %s\n", path)

	posResp, err := utils.MpvCall(socketPath, "get_property", "time-pos")
	if err == nil && posResp.Data != nil {
		pos := 0.0
		switch v := posResp.Data.(type) {
		case float64:
			pos = v
		}
		fmt.Printf("    Playhead: %s\n", utils.SecondsToHHMMSS(int64(pos)))
	}

	durResp, err := utils.MpvCall(socketPath, "get_property", "duration")
	if err == nil && durResp.Data != nil {
		dur := 0.0
		switch v := durResp.Data.(type) {
		case float64:
			dur = v
		}
		fmt.Printf("    Duration: %s\n", utils.SecondsToHHMMSS(int64(dur)))
	}

	return nil
}

type StopCmd struct {
	MpvControlBase
}

func (c *StopCmd) Run(ctx *kong.Context) error {
	socketPath := c.MpvSocket
	if socketPath == "" {
		socketPath = utils.GetMpvWatchSocket()
	}

	// Making mpv exit with code 3 like Python version (loadfile /dev/null)
	_, err := utils.MpvCall(socketPath, "loadfile", "/dev/null")
	if err != nil {
		return err
	}
	// Also delete the socket file as Python version does
	os.Remove(socketPath)
	return nil
}

type PauseCmd struct {
	MpvControlBase
}

func (c *PauseCmd) Run(ctx *kong.Context) error {
	socketPath := c.MpvSocket
	if socketPath == "" {
		socketPath = utils.GetMpvWatchSocket()
	}
	_, err := utils.MpvCall(socketPath, "cycle", "pause")
	return err
}

type NextCmd struct {
	MpvControlBase
}

func (c *NextCmd) Run(ctx *kong.Context) error {
	socketPath := c.MpvSocket
	if socketPath == "" {
		socketPath = utils.GetMpvWatchSocket()
	}
	_, err := utils.MpvCall(socketPath, "playlist_next", "force")
	return err
}

type SeekCmd struct {
	MpvControlBase
	Time string `arg:"" help:"Time to seek to (e.g. 10, +10, -10, 00:01:30)"`
}

func (c *SeekCmd) Run(ctx *kong.Context) error {
	socketPath := c.MpvSocket
	if socketPath == "" {
		socketPath = utils.GetMpvWatchSocket()
	}

	s := c.Time
	isRelative := false
	isNegative := false

	if strings.HasPrefix(s, "-") {
		isNegative = true
		isRelative = true
		s = s[1:]
	} else if strings.HasPrefix(s, "+") {
		isRelative = true
		s = s[1:]
	}

	var seconds float64
	if strings.Contains(s, ":") {
		seconds = utils.FromTimestampSeconds(s)
	} else {
		if f := utils.SafeFloat(s); f != nil {
			seconds = *f
		} else {
			return fmt.Errorf("invalid time format: %s", c.Time)
		}
	}

	if isNegative {
		seconds = -seconds
	}

	mode := "absolute"
	if isRelative {
		mode = "relative"
	} else if !strings.Contains(c.Time, ":") {
		// If it's just a number without +/- or :, Python logic implies it might be relative too?
		// Python: if ":" not in s: is_relative = True
		mode = "relative"
	}

	_, err := utils.MpvCall(socketPath, "seek", seconds, mode)
	return err
}
