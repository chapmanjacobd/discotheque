package tui

import (
	"fmt"
	"io"

	"github.com/chapmanjacobd/discotheque/internal/models"
	"github.com/chapmanjacobd/discotheque/internal/utils"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

type itemDelegate struct{}

func (d itemDelegate) Height() int                               { return 2 }
func (d itemDelegate) Spacing() int                              { return 1 }
func (d itemDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(item)
	if !ok {
		return
	}

	title := i.Title()
	desc := i.Description()

	if index == m.Index() {
		title = StyleSelected.Render(title)
		desc = StyleMuted.Render("  " + desc)
	} else {
		title = StyleNormal.Render(title)
		desc = StyleMuted.Render("  " + desc)
	}

	fmt.Fprintf(w, "%s\n%s", title, desc)
}

type item struct {
	media models.MediaWithDB
}

func (i item) Title() string {
	icon := "❓"
	if i.media.Type != nil {
		switch *i.media.Type {
		case "audio":
			icon = "🎵"
		case "video":
			icon = "🎬"
		case "text":
			icon = "📄"
		case "image":
			icon = "🖼️"
		}
	}

	title := i.media.Path
	if i.media.Title != nil && *i.media.Title != "" {
		title = *i.media.Title
	}
	return icon + " " + title
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

	tags := ""
	if i.media.Categories != nil && *i.media.Categories != "" {
		tags = " • " + *i.media.Categories
	}

	return fmt.Sprintf("%s • %s • %s%s", dur, size, i.media.DB, tags)
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

	l := list.New(items, itemDelegate{}, 0, 0)
	l.Title = "🪩  " + StyleLogoPrefix.Render("Disco") + StyleLogoSuffix.Render("theque") + " Media Picker"
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(true)
	l.Styles.Title = StyleTitle

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
		h, v := StyleDoc.GetFrameSize()
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
	return StyleDoc.Render(m.list.View())
}

func (m Model) GetChoice() *models.MediaWithDB {
	return m.choice
}
