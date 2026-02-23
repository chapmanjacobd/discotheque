package tui

import "github.com/charmbracelet/lipgloss"

var (
	// Colors inspired by the Web UI
	ColorAccent = lipgloss.Color("#77b3ff")
	ColorText   = lipgloss.Color("#e6e6e6")
	ColorMuted  = lipgloss.Color("#888888")
	ColorBG     = lipgloss.Color("#0f111a")
	ColorLogo1  = lipgloss.Color("#ff00ff")
	ColorLogo2  = lipgloss.Color("#00ffff")
	ColorBorder = lipgloss.Color("#32394d")

	// Styles
	StyleDoc = lipgloss.NewStyle().Margin(1, 2)

	StyleTitle = lipgloss.NewStyle().
			Foreground(ColorAccent).
			Bold(true).
			Padding(0, 1)

	StyleLogoSuffix = lipgloss.NewStyle().
			Foreground(ColorLogo2).
			Italic(true).
			Bold(true)

	StyleLogoPrefix = lipgloss.NewStyle().
			Foreground(ColorLogo1).
			Italic(false).
			Bold(true)

	StyleMuted = lipgloss.NewStyle().
			Foreground(ColorMuted)

	StyleSelected = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), false, false, false, true).
			BorderForeground(ColorAccent).
			PaddingLeft(1).
			Foreground(ColorAccent)

	StyleNormal = lipgloss.NewStyle().
			PaddingLeft(2)

	StyleHeader = lipgloss.NewStyle().
			Foreground(ColorLogo2).
			Bold(true).
			BorderStyle(lipgloss.NormalBorder()).
			BorderBottom(true).
			BorderForeground(ColorBorder).
			MarginBottom(1)

	StyleSidebar = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), false, true, false, false).
			BorderForeground(ColorBorder).
			Padding(0, 1).
			MarginRight(1)

	StyleActivePane = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(ColorAccent)

	StyleInactivePane = lipgloss.NewStyle().
				Border(lipgloss.NormalBorder()).
				BorderForeground(ColorBorder)
)
