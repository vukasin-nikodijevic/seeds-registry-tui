package ui

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	// Colors
	Primary   = lipgloss.Color("#7C3AED")
	Secondary = lipgloss.Color("#10B981")
	Danger    = lipgloss.Color("#EF4444")
	Warning   = lipgloss.Color("#F59E0B")
	Muted     = lipgloss.Color("#6B7280")
	White     = lipgloss.Color("#FFFFFF")
	BgDark    = lipgloss.Color("#1F2937")
	BgLight   = lipgloss.Color("#374151")
	Highlight = lipgloss.Color("#FBBF24")

	// Styles
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(White).
			Background(Primary).
			Padding(0, 2)

	SubtitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(Primary)

	SelectedStyle = lipgloss.NewStyle().
			Foreground(White).
			Background(Primary).
			Bold(true).
			Padding(0, 1)

	NormalStyle = lipgloss.NewStyle().
			Foreground(White).
			Padding(0, 1)

	MutedStyle = lipgloss.NewStyle().
			Foreground(Muted)

	InputStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Primary).
			Padding(0, 1)

	ActiveInputStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(Secondary).
				Padding(0, 1)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(Danger).
			Bold(true)

	SuccessStyle = lipgloss.NewStyle().
			Foreground(Secondary).
			Bold(true)

	BoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Primary).
			Padding(1, 2)

	// Matrix cell styles
	EmptyCellStyle = lipgloss.NewStyle().
			Foreground(Muted).
			Width(12).
			Align(lipgloss.Center)

	FilledCellStyle = lipgloss.NewStyle().
			Foreground(White).
			Background(BgLight).
			Width(12).
			Align(lipgloss.Center)

	HighlightCellStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#000000")).
				Background(Highlight).
				Bold(true).
				Width(12).
				Align(lipgloss.Center)

	HeaderCellStyle = lipgloss.NewStyle().
			Foreground(Primary).
			Bold(true).
			Width(12).
			Align(lipgloss.Center)

	HelpStyle = lipgloss.NewStyle().
			Foreground(Muted).
			Italic(true)

	TabActiveStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(White).
			Background(Primary).
			Padding(0, 2)

	TabInactiveStyle = lipgloss.NewStyle().
				Foreground(Muted).
				Padding(0, 2)
)
