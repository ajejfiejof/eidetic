package tui

import "github.com/charmbracelet/lipgloss"

type ThemeColors struct {
	Primary   lipgloss.Color
	Secondary lipgloss.Color
	Cyan      lipgloss.Color
	Yellow    lipgloss.Color
	Pink      lipgloss.Color
	Muted     lipgloss.Color
	BgDark    lipgloss.Color
	BgSelect  lipgloss.Color
	White     lipgloss.Color
}

var (
	// Trans Pride Palette
	ThemeTrans = ThemeColors{
		Primary:   lipgloss.Color("#F5A9B8"), // Pastel Pink
		Secondary: lipgloss.Color("#5BCEFA"), // Pastel Cyan
		Cyan:      lipgloss.Color("#5BCEFA"),
		Yellow:    lipgloss.Color("#F9E2AF"),
		Pink:      lipgloss.Color("#F5A9B8"),
		Muted:     lipgloss.Color("#7F849C"),
		BgDark:    lipgloss.Color("#11111B"),
		BgSelect:  lipgloss.Color("#313244"),
		White:     lipgloss.Color("#FFFFFF"),
	}

	// Catppuccin Mocha Palette
	ThemeCatppuccin = ThemeColors{
		Primary:   lipgloss.Color("#CBA6F7"), // Mauve
		Secondary: lipgloss.Color("#A6E3A1"), // Green
		Cyan:      lipgloss.Color("#89DCEB"), // Sky
		Yellow:    lipgloss.Color("#F9E2AF"), // Yellow
		Pink:      lipgloss.Color("#F5C2E7"), // Pink
		Muted:     lipgloss.Color("#6C7086"),
		BgDark:    lipgloss.Color("#181825"),
		BgSelect:  lipgloss.Color("#313244"),
		White:     lipgloss.Color("#CDD6F4"),
	}

	// Default Cyber Palette
	ThemeDefault = ThemeColors{
		Primary:   lipgloss.Color("#7D56F4"),
		Secondary: lipgloss.Color("#04B575"),
		Cyan:      lipgloss.Color("#00D7D7"),
		Yellow:    lipgloss.Color("#FFD700"),
		Pink:      lipgloss.Color("#FF5F87"),
		Muted:     lipgloss.Color("#626262"),
		BgDark:    lipgloss.Color("#1A1A24"),
		BgSelect:  lipgloss.Color("#353555"),
		White:     lipgloss.Color("#FAFAFA"),
	}

	CurrentTheme = ThemeTrans

	StyleTitle         lipgloss.Style
	StyleSearchPrompt  lipgloss.Style
	StyleFilterBadge   lipgloss.Style
	StyleSelectedRow   lipgloss.Style
	StyleNormalRow     lipgloss.Style
	StyleBadgeShell    lipgloss.Style
	StyleBadgeClip     lipgloss.Style
	StyleBadgeFile     lipgloss.Style
	StyleBadgeNote     lipgloss.Style
	StyleBadgeGit      lipgloss.Style
	StyleBadgePinned   lipgloss.Style
	StyleMuted         lipgloss.Style
	StylePreviewBox    lipgloss.Style
	StyleStatusBar     lipgloss.Style
	StyleHighlightTerm lipgloss.Style
)

func InitTheme(name string) {
	switch name {
	case "trans", "transgender", "pride":
		CurrentTheme = ThemeTrans
	case "catppuccin", "mocha":
		CurrentTheme = ThemeCatppuccin
	default:
		CurrentTheme = ThemeDefault
	}

	StyleTitle = lipgloss.NewStyle().
		Bold(true).
		Foreground(CurrentTheme.Primary).
		Padding(0, 1)

	StyleSearchPrompt = lipgloss.NewStyle().
		Bold(true).
		Foreground(CurrentTheme.Secondary).
		SetString("❯ ")

	StyleFilterBadge = lipgloss.NewStyle().
		Bold(true).
		Foreground(CurrentTheme.White).
		Background(CurrentTheme.Primary).
		Padding(0, 1).
		MarginRight(1)

	StyleSelectedRow = lipgloss.NewStyle().
		Bold(true).
		Foreground(CurrentTheme.White).
		Background(CurrentTheme.BgSelect).
		Padding(0, 1)

	StyleNormalRow = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#BAC2DE")).
		Padding(0, 1)

	StyleBadgeShell = lipgloss.NewStyle().
		Bold(true).
		Foreground(CurrentTheme.Secondary)

	StyleBadgeClip = lipgloss.NewStyle().
		Bold(true).
		Foreground(CurrentTheme.Cyan)

	StyleBadgeFile = lipgloss.NewStyle().
		Bold(true).
		Foreground(CurrentTheme.Yellow)

	StyleBadgeNote = lipgloss.NewStyle().
		Bold(true).
		Foreground(CurrentTheme.Primary)

	StyleBadgeGit = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#F38BA8"))

	StyleBadgePinned = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FFD700"))

	StyleMuted = lipgloss.NewStyle().
		Foreground(CurrentTheme.Muted)

	StylePreviewBox = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(CurrentTheme.Primary).
		Padding(1).
		MarginLeft(1)

	StyleStatusBar = lipgloss.NewStyle().
		Foreground(CurrentTheme.White).
		Background(CurrentTheme.BgDark).
		Padding(0, 1)

	StyleHighlightTerm = lipgloss.NewStyle().
		Bold(true).
		Foreground(CurrentTheme.Yellow).
		Underline(true)
}

func init() {
	InitTheme("trans")
}
