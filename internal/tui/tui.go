package tui

import (
	"fmt"

	"github.com/chapmanjacobd/discotheque/internal/models"
	"github.com/chapmanjacobd/discotheque/internal/utils"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	docStyle   = lipgloss.NewStyle().Margin(1, 2)
	titleStyle = lipgloss.NewStyle().MarginLeft(2)
)

type item struct {
	media models.MediaWithDB
}

func (i item) Title() string {
	if i.media.Title != nil && *i.media.Title != "" {
		return *i.media.Title
	}
	return i.media.Path
}

func (i item) Description() string {
	dur := "-"
	if i.media.Duration != nil {
		dur = utils.FormatDuration(int(*i.media.Duration))
	}
	size := "-"
	if i.media.Size != nil {
		size = utils.FormatSize(*i.media.Size)
	}
	return fmt.Sprintf("%s • %s • %s", dur, size, i.media.DB)
}

func (i item) FilterValue() string {
	val := i.media.Path
	if i.media.Title != nil {
		val += " " + *i.media.Title
	}
	return val
}

type Model struct {
	list     list.Model
	choice   *models.MediaWithDB
	quitting bool
}

func NewModel(media []models.MediaWithDB) Model {
	items := make([]list.Item, len(media))
	for i, m := range media {
		items[i] = item{media: m}
	}

	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Discotheque Media Picker"
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(true)
	l.Styles.Title = titleStyle

	return Model{list: l}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			m.quitting = true
			return m, tea.Quit
		}
		if msg.String() == "enter" {
			i, ok := m.list.SelectedItem().(item)
			if ok {
				m.choice = &i.media
			}
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m Model) View() string {
	if m.quitting {
		return ""
	}
	return docStyle.Render(m.list.View())
}

func (m Model) GetChoice() *models.MediaWithDB {
	return m.choice
}
