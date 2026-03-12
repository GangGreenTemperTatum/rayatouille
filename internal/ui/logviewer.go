package ui

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// LogViewer wraps a viewport for scrollable, searchable log content.
type LogViewer struct {
	viewport     viewport.Model
	searching    bool
	searchQuery  string
	content      string
	lines        []string
	matchCount   int
	currentMatch int
	width        int
	height       int
	hasContent   bool
	copied       bool
}

// NewLogViewer creates a log viewer with the given dimensions.
func NewLogViewer(width, height int) LogViewer {
	vp := viewport.New(
		viewport.WithWidth(width),
		viewport.WithHeight(height),
	)
	vp.MouseWheelEnabled = true

	vp.LeftGutterFunc = func(info viewport.GutterContext) string {
		if info.Soft {
			return "     | "
		}
		if info.Index >= info.TotalLines {
			return "   ~ | "
		}
		return fmt.Sprintf("%4d | ", info.Index+1)
	}

	vp.HighlightStyle = lipgloss.NewStyle().
		Background(ColorWarning).
		Foreground(lipgloss.Color("#000000"))
	vp.SelectedHighlightStyle = lipgloss.NewStyle().
		Background(ColorAccent).
		Foreground(lipgloss.Color("#000000"))

	return LogViewer{
		viewport: vp,
		width:    width,
		height:   height,
	}
}

// SetContent sets the log content in the viewport and scrolls to the bottom.
func (m *LogViewer) SetContent(content string) {
	m.content = content
	m.lines = strings.Split(content, "\n")
	m.viewport.SetContent(content)
	m.viewport.GotoBottom()
	m.hasContent = true
}

// SetSize updates the viewport dimensions.
func (m *LogViewer) SetSize(w, h int) {
	m.width = w
	m.height = h
	m.viewport.SetWidth(w)
	m.viewport.SetHeight(h)
}

// Update handles messages for the log viewer.
func (m LogViewer) Update(msg tea.Msg) (LogViewer, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if m.searching {
			return m.updateSearchMode(msg)
		}
		return m.updateNormalMode(msg)
	}

	// Forward non-key messages to viewport.
	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

// updateSearchMode handles key events while in search mode.
func (m LogViewer) updateSearchMode(msg tea.KeyPressMsg) (LogViewer, tea.Cmd) {
	switch msg.Code {
	case tea.KeyEnter:
		m.searching = false
		if m.searchQuery != "" {
			m.performSearch()
		}
		return m, nil
	case tea.KeyEscape:
		m.searching = false
		m.searchQuery = ""
		m.matchCount = 0
		m.currentMatch = 0
		m.viewport.ClearHighlights()
		return m, nil
	case tea.KeyBackspace:
		if len(m.searchQuery) > 0 {
			m.searchQuery = m.searchQuery[:len(m.searchQuery)-1]
		}
		return m, nil
	default:
		// Append printable characters.
		if msg.Text != "" {
			m.searchQuery += msg.Text
		}
		return m, nil
	}
}

// updateNormalMode handles key events in normal (non-search) mode.
func (m LogViewer) updateNormalMode(msg tea.KeyPressMsg) (LogViewer, tea.Cmd) {
	// Clear copy indicator on any key.
	m.copied = false
	switch msg.Text {
	case "/":
		m.searching = true
		m.searchQuery = ""
		return m, nil
	case "n":
		if m.matchCount > 0 {
			m.viewport.HighlightNext()
			m.currentMatch = (m.currentMatch + 1) % m.matchCount
		}
		return m, nil
	case "N":
		if m.matchCount > 0 {
			m.viewport.HighlightPrevious()
			m.currentMatch = (m.currentMatch - 1 + m.matchCount) % m.matchCount
		}
		return m, nil
	case "y":
		// Copy current line to clipboard.
		line := m.currentLine()
		if line != "" {
			m.copied = true
			return m, tea.SetClipboard(line)
		}
		return m, nil
	case "Y":
		// Copy all visible lines to clipboard.
		visible := m.visibleContent()
		if visible != "" {
			m.copied = true
			return m, tea.SetClipboard(visible)
		}
		return m, nil
	case "G":
		m.viewport.GotoBottom()
		return m, nil
	case "g":
		m.viewport.GotoTop()
		return m, nil
	}

	// Everything else (j/k, space/b, u/d, pgup/pgdn, arrows) goes to viewport.
	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

// performSearch finds all case-insensitive matches and highlights them.
func (m *LogViewer) performSearch() {
	if m.searchQuery == "" || m.content == "" {
		m.matchCount = 0
		m.currentMatch = 0
		m.viewport.ClearHighlights()
		return
	}

	lowerContent := strings.ToLower(m.content)
	lowerQuery := strings.ToLower(m.searchQuery)

	var highlights [][]int
	idx := 0
	for {
		pos := strings.Index(lowerContent[idx:], lowerQuery)
		if pos == -1 {
			break
		}
		start := idx + pos
		end := start + len(m.searchQuery)
		highlights = append(highlights, []int{start, end})
		idx = end
	}

	m.matchCount = len(highlights)
	m.currentMatch = 0

	if m.matchCount > 0 {
		m.viewport.SetHighlights(highlights)
		m.viewport.HighlightNext()
		m.currentMatch = 1
	} else {
		m.viewport.ClearHighlights()
	}
}

// currentLine returns the line at the current viewport offset.
func (m LogViewer) currentLine() string {
	offset := m.viewport.YOffset()
	if offset < 0 || offset >= len(m.lines) {
		return ""
	}
	return m.lines[offset]
}

// visibleContent returns all lines currently visible in the viewport.
func (m LogViewer) visibleContent() string {
	offset := m.viewport.YOffset()
	end := offset + m.viewport.Height()
	if offset < 0 {
		offset = 0
	}
	if end > len(m.lines) {
		end = len(m.lines)
	}
	if offset >= end {
		return ""
	}
	return strings.Join(m.lines[offset:end], "\n")
}

// View renders the log viewer.
func (m LogViewer) View() string {
	if !m.hasContent {
		return lipgloss.NewStyle().Foreground(ColorMuted).Render("No logs available")
	}

	view := m.viewport.View()

	// Status line below viewport.
	var statusLine string
	if m.searching {
		statusLine = lipgloss.NewStyle().Foreground(ColorWarning).Render("/") +
			m.searchQuery +
			lipgloss.NewStyle().Foreground(ColorMuted).Render("_")
	} else if m.copied {
		statusLine = lipgloss.NewStyle().Foreground(ColorSuccess).Render("Copied to clipboard!")
	} else if m.matchCount > 0 {
		statusLine = lipgloss.NewStyle().Foreground(ColorMuted).Render(
			fmt.Sprintf("Match %d/%d", m.currentMatch, m.matchCount),
		)
	}

	if statusLine != "" {
		return lipgloss.JoinVertical(lipgloss.Left, view, statusLine)
	}

	return view
}

// Searching returns whether the log viewer is in search mode.
func (m LogViewer) Searching() bool {
	return m.searching
}

// HasContent returns whether the log viewer has content set.
func (m LogViewer) HasContent() bool {
	return m.hasContent
}

// MatchCount returns the number of search matches.
func (m LogViewer) MatchCount() int {
	return m.matchCount
}
