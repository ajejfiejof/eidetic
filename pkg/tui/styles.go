package tui

import "github.com/charmbracelet/lipgloss"

var (
	// Colors
	ColorPrimary   = lipgloss.Color("#7D56F4") // Purple
	ColorSecondary = lipgloss.Color("#04B575") // Green
	ColorCyan      = lipgloss.Color("#00D7D7") // Cyan
	ColorYellow    = lipgloss.Color("#FFD700") // Yellow
	ColorMuted     = lipgloss.Color("#626262") // Gray
	ColorHighlight = lipgloss.Color("#FF5F87") // Coral/Pink
	ColorBgDark    = lipgloss.Color("#1A1A24")
	ColorWhite     = lipgloss.Color("#FAFAFA")

	// Component Styles
	StyleTitle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorPrimary).
			Padding(0, 1)

	StyleSearchPrompt = lipgloss.NewStyle().
				Bold(true).
				Foreground(ColorSecondary).
				SetString("❯ ")

	StyleFilterBadge = lipgloss.NewStyle().
				Bold(true).
				Foreground(ColorWhite).
				Background(ColorPrimary).
				Padding(0, 1).
				MarginRight(1)

	StyleSelectedRow = lipgloss.NewStyle().
				Bold(true).
				Foreground(ColorWhite).
				Background(lipgloss.Color("#353555")).
				Padding(0, 1)

	StyleNormalRow = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#D0D0D0")).
			Padding(0, 1)

	StyleBadgeShell = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorSecondary)

	StyleBadgeClip = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorCyan)

	StyleBadgeFile = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorYellow)

	StyleBadgeNote = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorPrimary)

	StyleMuted = lipgloss.NewStyle().
			Foreground(ColorMuted)

	StylePreviewBox = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorPrimary).
			Padding(1).
			MarginLeft(1)

	StyleStatusBar = lipgloss.NewStyle().
			Foreground(ColorWhite).
			Background(ColorBgDark).
			Padding(0, 1)
)
