package tui

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/chapmanjacobd/discotheque/internal/models"
	"github.com/chapmanjacobd/discotheque/internal/query"
	"github.com/chapmanjacobd/discotheque/internal/utils"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	sizeBarStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	barFullStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
)

type duItem struct {
	stats models.FolderStats
	isDir bool
}

func (i duItem) Title() string {
	if i.isDir {
		return "📁 " + filepath.Base(i.stats.Path)
	}
	return "📄 " + filepath.Base(i.stats.Path)
}

func (i duItem) Description() string {
	return fmt.Sprintf("%s • %d files • %s",
		utils.FormatSize(i.stats.TotalSize),
		i.stats.Count,
		utils.FormatDuration(int(i.stats.TotalDuration)))
}

func (i duItem) FilterValue() string {
	return i.stats.Path
}

type DUModel struct {
	list        list.Model
	allMedia    []models.MediaWithDB
	currentPath string
	history     []string
	totalSize   int64
	quitting    bool
	flags       models.GlobalFlags
}

func NewDUModel(media []models.MediaWithDB, flags models.GlobalFlags) DUModel {
	m := DUModel{
		allMedia: media,
		flags:    flags,
	}
	m.updateList()
	return m
}

func (m *DUModel) updateList() {
	var currentMedia []models.MediaWithDB
	if m.currentPath == "" {
		currentMedia = m.allMedia
	} else {
		for _, med := range m.allMedia {
			if strings.HasPrefix(med.Path, m.currentPath) {
				currentMedia = append(currentMedia, med)
			}
		}
	}

	// Determine next depth
	depth := 1
	if m.currentPath != "" {
		depth = strings.Count(filepath.Clean(m.currentPath), string(filepath.Separator)) + 1
	}

	tempFlags := m.flags
	tempFlags.Depth = depth
	tempFlags.Parents = false

	stats := query.AggregateMedia(currentMedia, tempFlags)

	items := make([]list.Item, len(stats))
	maxSize := int64(0)
	for i, s := range stats {
		items[i] = duItem{stats: s, isDir: true} // Logic to distinguish files vs dirs could be added
		if s.TotalSize > maxSize {
			maxSize = s.TotalSize
		}
	}
	m.totalSize = maxSize

	l := list.New(items, duDelegate{maxSize: maxSize}, 0, 0)
	l.Title = "Disk Usage: " + m.currentPath
	if m.currentPath == "" {
		l.Title = "Disk Usage: Root"
	}
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(true)
	l.Styles.Title = titleStyle
	m.list = l
}

type duDelegate struct {
	maxSize int64
}

func (d duDelegate) Height() int                               { return 2 }
func (d duDelegate) Spacing() int                              { return 1 }
func (d duDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }
func (d duDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(duItem)
	if !ok {
		return
	}

	styles := list.NewDefaultItemStyles()
	title := i.Title()
	desc := i.Description()

	if index == m.Index() {
		title = styles.SelectedTitle.Render(title)
		desc = styles.SelectedDesc.Render(desc)
	} else {
		title = styles.NormalTitle.Render(title)
		desc = styles.NormalDesc.Render(desc)
	}

	barWidth := 20
	filled := 0
	if d.maxSize > 0 {
		filled = int(float64(i.stats.TotalSize) / float64(d.maxSize) * float64(barWidth))
	}
	bar := "[" + barFullStyle.Render(strings.Repeat("#", filled)) + sizeBarStyle.Render(strings.Repeat("-", barWidth-filled)) + "]"

	fmt.Fprintf(w, "%s %s\n%s", title, bar, desc)
}

func (m DUModel) Init() tea.Cmd {
	return nil
}

func (m DUModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.list.FilterState() == list.Filtering {
			break
		}

		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit
		case "enter", "right":
			i, ok := m.list.SelectedItem().(duItem)
			if ok && i.isDir {
				m.history = append(m.history, m.currentPath)
				m.currentPath = i.stats.Path
				m.updateList()
				return m, nil
			}
		case "backspace", "left":
			if len(m.history) > 0 {
				m.currentPath = m.history[len(m.history)-1]
				m.history = m.history[:len(m.history)-1]
				m.updateList()
				return m, nil
			}
		}
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m DUModel) View() string {
	if m.quitting {
		return ""
	}
	return docStyle.Render(m.list.View())
}
