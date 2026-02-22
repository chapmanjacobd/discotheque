package tui

import (
	"fmt"
	"io"
	"strings"

	"github.com/chapmanjacobd/discotheque/internal/models"
	"github.com/chapmanjacobd/discotheque/internal/utils"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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
	list        list.Model
	choice      *models.MediaWithDB
	showDetails bool
	quitting    bool
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
	l.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{
			key.NewBinding(
				key.WithKeys("d"),
				key.WithHelp("d", "toggle details"),
			),
		}
	}
	l.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{
			key.NewBinding(
				key.WithKeys("d"),
				key.WithHelp("d", "details"),
			),
		}
	}

	return Model{list: l}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.showDetails {
			if msg.String() == "d" || msg.String() == "esc" || msg.String() == "q" {
				m.showDetails = false
				return m, nil
			}
			return m, nil
		}

		if msg.String() == "ctrl+c" {
			m.quitting = true
			return m, tea.Quit
		}
		if msg.String() == "d" {
			m.showDetails = true
			return m, nil
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
	if m.showDetails {
		return StyleDoc.Render(m.renderDetails())
	}
	return StyleDoc.Render(m.list.View())
}

func (m Model) renderDetails() string {
	i, ok := m.list.SelectedItem().(item)
	if !ok {
		return "No item selected"
	}
	media := i.media.Media

	var sb strings.Builder
	sb.WriteString(StyleHeader.Render("Media Details") + "\n\n")

	addField := func(label, value string) {
		if value != "" && value != "0" && value != "<nil>" {
			sb.WriteString(lipgloss.NewStyle().Foreground(ColorAccent).Bold(true).Render(label+": ") + value + "\n")
		}
	}

	addField("Path", media.Path)
	if media.Title != nil {
		addField("Title", *media.Title)
	}
	if media.Type != nil {
		addField("Type", *media.Type)
	}
	if media.Duration != nil {
		addField("Duration", utils.FormatDuration(int(*media.Duration)))
	}
	if media.Size != nil {
		addField("Size", utils.FormatSize(*media.Size))
	}
	if media.VideoCodecs != nil {
		addField("Video Codec", *media.VideoCodecs)
	}
	if media.AudioCodecs != nil {
		addField("Audio Codec", *media.AudioCodecs)
	}
	if media.Width != nil && media.Height != nil {
		addField("Resolution", fmt.Sprintf("%dx%d", *media.Width, *media.Height))
	}
	if media.Album != nil {
		addField("Album", *media.Album)
	}
	if media.Artist != nil {
		addField("Artist", *media.Artist)
	}
	if media.Genre != nil {
		addField("Genre", *media.Genre)
	}
	if media.Categories != nil {
		addField("Categories", *media.Categories)
	}
	if media.Score != nil {
		addField("Score", fmt.Sprintf("%.1f", *media.Score))
	}

	sb.WriteString("\n" + StyleMuted.Render("Press 'd', 'esc', or 'q' to return"))

	return sb.String()
}

func (m Model) GetChoice() *models.MediaWithDB {
	return m.choice
}
