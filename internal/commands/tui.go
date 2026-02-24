package commands

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/alecthomas/kong"
	"github.com/chapmanjacobd/discotheque/internal/db"
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

	var customCats []string
	for _, dbPath := range c.Databases {
		sqlDB, err := db.Connect(dbPath)
		if err == nil {
			queries := db.New(sqlDB)
			cats, err := queries.GetCustomCategories(context.Background())
			if err == nil {
				customCats = append(customCats, cats...)
			}
			sqlDB.Close()
		}
	}

	m := tui.NewModel(media, c.Databases, c.GlobalFlags, customCats)
	p := tea.NewProgram(m, tea.WithAltScreen())

	finalModel, err := p.Run()
	if err != nil {
		return err
	}

	choice := finalModel.(*tui.Model).GetChoice()
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
