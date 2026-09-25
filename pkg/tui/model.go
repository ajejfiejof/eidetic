package tui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/ajejfiejof/eidetic/pkg/document"
	"github.com/ajejfiejof/eidetic/pkg/index"
)

var filterCycle = []document.SourceType{
	"",
	document.SourceShell,
	document.SourceClipboard,
	document.SourceFile,
	document.SourceNote,
}

// Model represents the state of the interactive Eidetic TUI.
type Model struct {
	engine        *index.Engine
	query         string
	results       []index.SearchResult
	selectedIndex int
	sourceFilter  document.SourceType
	filterIdx     int
	selectedDoc   *document.Document
	width         int
	height        int
	searchLatency time.Duration
	cancelled     bool
}

// NewModel creates an initial TUI model.
func NewModel(engine *index.Engine, initialQuery string) Model {
	m := Model{
		engine:       engine,
		query:        initialQuery,
		sourceFilter: "",
	}
	m.runSearch()
	return m
}

func (m *Model) runSearch() {
	start := time.Now()
	m.results = m.engine.Search(m.query, index.SearchFilter{
		Source: m.sourceFilter,
		Limit:  100,
	})
	m.searchLatency = time.Since(start)

	if m.selectedIndex >= len(m.results) {
		m.selectedIndex = max(0, len(m.results)-1)
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			m.cancelled = true
			return m, tea.Quit

		case tea.KeyEnter:
			if len(m.results) > 0 && m.selectedIndex < len(m.results) {
				doc := m.results[m.selectedIndex].Doc
				m.selectedDoc = &doc
			}
			return m, tea.Quit

		case tea.KeyUp, tea.KeyCtrlK, tea.KeyCtrlP:
			if m.selectedIndex > 0 {
				m.selectedIndex--
			}
			return m, nil

		case tea.KeyDown, tea.KeyCtrlJ, tea.KeyCtrlN:
			if m.selectedIndex < len(m.results)-1 {
				m.selectedIndex++
			}
			return m, nil

		case tea.KeyTab:
			m.filterIdx = (m.filterIdx + 1) % len(filterCycle)
			m.sourceFilter = filterCycle[m.filterIdx]
			m.runSearch()
			return m, nil

		case tea.KeyBackspace:
			if len(m.query) > 0 {
				m.query = m.query[:len(m.query)-1]
				m.runSearch()
			}
			return m, nil

		case tea.KeyRunes, tea.KeySpace:
			m.query += msg.String()
			m.selectedIndex = 0
			m.runSearch()
			return m, nil
		}
	}
	return m, nil
}

func formatRelativeTime(t time.Time) string {
	diff := time.Since(t)
	if diff < time.Minute {
		return "just now"
	}
	if diff < time.Hour {
		return fmt.Sprintf("%dm ago", int(diff.Minutes()))
	}
	if diff < 24*time.Hour {
		return fmt.Sprintf("%dh ago", int(diff.Hours()))
	}
	days := int(diff.Hours() / 24)
	if days < 30 {
		return fmt.Sprintf("%dd ago", days)
	}
	return t.Format("Jan 02")
}

func sourceBadge(src document.SourceType) string {
	switch src {
	case document.SourceShell:
		return StyleBadgeShell.Render("[SHELL]")
	case document.SourceClipboard:
		return StyleBadgeClip.Render("[CLIP ]")
	case document.SourceFile:
		return StyleBadgeFile.Render("[FILE ]")
	case document.SourceNote:
		return StyleBadgeNote.Render("[NOTE ]")
	default:
		return StyleMuted.Render("[OTHER]")
	}
}

func (m Model) View() string {
	if m.width == 0 {
		return "Initializing Eidetic..."
	}

	var sb strings.Builder

	// 1. Header / Search Bar
	filterText := "ALL SOURCES"
	if m.sourceFilter != "" {
		filterText = strings.ToUpper(string(m.sourceFilter))
	}
	badge := StyleFilterBadge.Render(filterText)
	prompt := StyleSearchPrompt.Render("❯")
	searchBar := lipgloss.JoinHorizontal(lipgloss.Center, badge, prompt, " "+m.query)
	sb.WriteString(searchBar + "\n\n")

	// 2. Split Screen View
	leftWidth := (m.width * 5) / 10
	if leftWidth < 40 {
		leftWidth = 40
	}
	rightWidth := m.width - leftWidth - 6
	if rightWidth < 20 {
		rightWidth = 20
	}

	availableHeight := m.height - 7
	if availableHeight < 5 {
		availableHeight = 5
	}

	// Build left results list
	var listLines []string
	startIdx := 0
	if m.selectedIndex >= availableHeight {
		startIdx = m.selectedIndex - availableHeight + 1
	}
	endIdx := min(len(m.results), startIdx+availableHeight)

	for i := startIdx; i < endIdx; i++ {
		res := m.results[i]
		badgeStr := sourceBadge(res.Doc.Source)
		timeStr := StyleMuted.Render(fmt.Sprintf("%8s", formatRelativeTime(res.Doc.Timestamp)))

		titleText := res.Doc.Title
		maxTitleLen := leftWidth - 22
		if maxTitleLen > 5 && len(titleText) > maxTitleLen {
			titleText = titleText[:maxTitleLen-3] + "..."
		}

		line := fmt.Sprintf("%s %s %s", badgeStr, timeStr, titleText)
		if i == m.selectedIndex {
			line = StyleSelectedRow.Width(leftWidth).Render("> " + line)
		} else {
			line = StyleNormalRow.Width(leftWidth).Render("  " + line)
		}
		listLines = append(listLines, line)
	}

	for len(listLines) < availableHeight {
		listLines = append(listLines, "")
	}
	leftPane := strings.Join(listLines, "\n")

	// Build right preview pane
	var rightPane string
	if len(m.results) > 0 && m.selectedIndex < len(m.results) {
		selected := m.results[m.selectedIndex].Doc
		metaLines := []string{
			lipgloss.NewStyle().Bold(true).Foreground(ColorWhite).Render(selected.Title),
			StyleMuted.Render(fmt.Sprintf("Source: %s | Time: %s", selected.Source, selected.Timestamp.Format("2006-01-02 15:04:05"))),
		}
		if p, ok := selected.Metadata["path"]; ok {
			metaLines = append(metaLines, StyleMuted.Render("Path: "+p))
		}
		metaLines = append(metaLines, strings.Repeat("─", rightWidth-4))

		// Content snippet
		content := selected.Content
		contentLines := strings.Split(content, "\n")
		for _, cl := range contentLines {
			if len(cl) > rightWidth-6 {
				cl = cl[:rightWidth-9] + "..."
			}
			metaLines = append(metaLines, cl)
		}

		rightPane = StylePreviewBox.
			Width(rightWidth).
			Height(availableHeight).
			Render(strings.Join(metaLines, "\n"))
	} else {
		rightPane = StylePreviewBox.
			Width(rightWidth).
			Height(availableHeight).
			Render(StyleMuted.Render("No documents matched your query.\nType to search or press Tab to cycle sources."))
	}

	mainView := lipgloss.JoinHorizontal(lipgloss.Top, leftPane, rightPane)
	sb.WriteString(mainView + "\n\n")

	// 3. Status Footer
	stats := fmt.Sprintf(" %d matches (%.2fms) | %d indexed | Tab: filter source | ↑/↓: navigate | Enter: copy/select | Esc: exit ",
		len(m.results), float64(m.searchLatency.Microseconds())/1000.0, m.engine.Count())
	sb.WriteString(StyleStatusBar.Width(m.width).Render(stats))

	return sb.String()
}

// RunTUI launches the interactive search interface and returns selected doc.
func RunTUI(engine *index.Engine, initialQuery string) (*document.Document, error) {
	p := tea.NewProgram(NewModel(engine, initialQuery), tea.WithAltScreen())
	finalModel, err := p.Run()
	if err != nil {
		return nil, err
	}

	m := finalModel.(Model)
	if m.cancelled {
		return nil, nil
	}
	return m.selectedDoc, nil
}
