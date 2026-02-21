package commands

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/alecthomas/kong"
	"github.com/chapmanjacobd/discotheque/internal/models"
	"github.com/chapmanjacobd/discotheque/internal/query"
	"github.com/chapmanjacobd/discotheque/internal/tui"
	tea "github.com/charmbracelet/bubbletea"
)

type TuiCmd struct {
	models.GlobalFlags
	Databases []string `arg:"" required:"" help:"SQLite database files" type:"existingfile"`
}

func (c TuiCmd) IsFilterTrait() {}
func (c TuiCmd) IsSortTrait()   {}

func (c *TuiCmd) Run(ctx *kong.Context) error {
	models.SetupLogging(c.Verbose)

	media, err := query.MediaQuery(context.Background(), c.Databases, c.GlobalFlags)
	if err != nil {
		return err
	}

	if len(media) == 0 {
		return fmt.Errorf("no media found")
	}

	query.SortMedia(media, c.GlobalFlags)

	m := tui.NewModel(media)
	p := tea.NewProgram(m, tea.WithAltScreen())

	finalModel, err := p.Run()
	if err != nil {
		return err
	}

	choice := finalModel.(tui.Model).GetChoice()
	if choice != nil {
		// Play the chosen media
		fmt.Printf("Playing: %s\n", choice.Path)

		args := []string{"mpv", choice.Path}
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin

		return cmd.Run()
	}

	return nil
}
