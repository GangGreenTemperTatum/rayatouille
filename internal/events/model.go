package events

import (
	"fmt"
	"sort"
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/table"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/GangGreenTemperTatum/rayatouille/internal/ray"
	"github.com/GangGreenTemperTatum/rayatouille/internal/ui"
)

// SeverityFilter defines which event severities to display.
type SeverityFilter int

const (
	SeverityAll SeverityFilter = iota
	SeverityError
	SeverityWarning
	SeverityInfo
)

// severityFilterLabel returns the display label for a severity filter.
func severityFilterLabel(sf SeverityFilter) string {
	switch sf {
	case SeverityError:
		return "ERROR"
	case SeverityWarning:
		return "WARNING"
	case SeverityInfo:
		return "INFO"
	default:
		return "ALL"
	}
}

// SortField defines which column to sort by.
type SortField int

const (
	SortByTime SortField = iota
	SortBySeverity
	SortBySource
)

// sortFieldLabel returns the display label for a sort field.
func sortFieldLabel(sf SortField) string {
	switch sf {
	case SortBySeverity:
		return "Severity"
	case SortBySource:
		return "Source"
	default:
		return "Time"
	}
}

// SortOrder defines the sort direction.
type SortOrder int

const (
	SortDesc SortOrder = iota
	SortAsc
)

// Model is the events timeline view model.
type Model struct {
	table          table.Model
	filter         ui.FilterModel
	keyMap         KeyMap
	allEvents      []ray.ClusterEvent
	filteredEvents []ray.ClusterEvent
	severityFilter SeverityFilter
	sortField      SortField
	sortOrder      SortOrder
	width          int
	height         int
	ready          bool
}

// New creates a new events list model.
func New() Model {
	km := table.DefaultKeyMap()
	km.LineUp.SetKeys("k", "up")
	km.LineDown.SetKeys("j", "down")

	styles := table.DefaultStyles()
	styles.Selected = lipgloss.NewStyle().Bold(true).Foreground(ui.ColorPrimary)
	styles.Header = lipgloss.NewStyle().Bold(true).Foreground(ui.ColorFg).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(ui.ColorMuted).
		BorderBottom(true).
		Padding(0, 1)
	styles.Cell = lipgloss.NewStyle().Padding(0, 1)

	t := table.New(
		table.WithColumns(Columns(80)),
		table.WithFocused(true),
		table.WithStyles(styles),
		table.WithKeyMap(km),
	)

	return Model{
		table:     t,
		filter:    ui.NewFilter(),
		keyMap:    Keys,
		sortField: SortByTime,
		sortOrder: SortDesc,
	}
}

// SetSize updates the events view dimensions.
func (m *Model) SetSize(w, h int) {
	m.width = w
	m.height = h
	m.ready = true
	// Reserve space: 1 header, 1 filter bar, 1 footer, 1 padding.
	tableHeight := h - 5
	if tableHeight < 3 {
		tableHeight = 3
	}
	m.table.SetWidth(w)
	m.table.SetHeight(tableHeight)
	m.table.SetColumns(Columns(w))
}

// SetEvents updates the event list.
func (m *Model) SetEvents(events []ray.ClusterEvent) {
	m.allEvents = events
	m.applyFilters()
}

// applyFilters applies severity filter, text filter, and sorting, then updates table rows.
func (m *Model) applyFilters() {
	filtered := make([]ray.ClusterEvent, 0, len(m.allEvents))
	for _, e := range m.allEvents {
		// Severity filter.
		if m.severityFilter != SeverityAll {
			label := severityFilterLabel(m.severityFilter)
			if !strings.EqualFold(e.Severity, label) {
				continue
			}
		}
		// Text filter (match on Severity, SourceType, Message, Time).
		if !m.filter.Matches(e.Severity) && !m.filter.Matches(e.SourceType) &&
			!m.filter.Matches(e.Message) && !m.filter.Matches(e.Time) {
			continue
		}
		filtered = append(filtered, e)
	}

	// Sort.
	sort.Slice(filtered, func(i, j int) bool {
		cmp := m.compareEvents(filtered[i], filtered[j])
		if m.sortOrder == SortDesc {
			return cmp > 0
		}
		return cmp < 0
	})

	m.filteredEvents = filtered

	// Save cursor, update rows, restore cursor.
	cursor := m.table.Cursor()
	rows := make([]table.Row, len(filtered))
	for i, e := range filtered {
		rows[i] = EventToRow(e)
	}
	m.table.SetRows(rows)
	if cursor >= len(rows) {
		cursor = len(rows) - 1
	}
	if cursor < 0 {
		cursor = 0
	}
	m.table.SetCursor(cursor)
}

// compareEvents compares two events for sorting. Returns <0, 0, or >0.
func (m *Model) compareEvents(a, b ray.ClusterEvent) int {
	switch m.sortField {
	case SortBySeverity:
		return severityWeight(a.Severity) - severityWeight(b.Severity)
	case SortBySource:
		return strings.Compare(a.SourceType, b.SourceType)
	default: // SortByTime
		// ISO timestamps sort correctly as strings.
		return strings.Compare(a.Time, b.Time)
	}
}

// severityWeight returns a numeric weight for severity ordering (higher = more severe).
func severityWeight(s string) int {
	switch strings.ToUpper(s) {
	case "ERROR":
		return 3
	case "WARNING":
		return 2
	case "INFO":
		return 1
	default:
		return 0
	}
}

// Update handles messages for the events list.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	// If filter is active, route key events to it first.
	if m.filter.Active() {
		if _, ok := msg.(tea.KeyPressMsg); ok {
			wasActive := m.filter.Active()
			var cmd tea.Cmd
			m.filter, cmd = m.filter.Update(msg)
			// If filter just deactivated, re-apply filters.
			if wasActive && !m.filter.Active() {
				m.applyFilters()
			}
			return m, cmd
		}
	}

	if msg, ok := msg.(tea.KeyPressMsg); ok {
		switch {
		case key.Matches(msg, m.keyMap.SeverityAll):
			m.severityFilter = SeverityAll
			m.applyFilters()
			return m, nil
		case key.Matches(msg, m.keyMap.SeverityError):
			m.severityFilter = SeverityError
			m.applyFilters()
			return m, nil
		case key.Matches(msg, m.keyMap.SeverityWarning):
			m.severityFilter = SeverityWarning
			m.applyFilters()
			return m, nil
		case key.Matches(msg, m.keyMap.SeverityInfo):
			m.severityFilter = SeverityInfo
			m.applyFilters()
			return m, nil
		case key.Matches(msg, m.keyMap.Sort):
			m.sortField = (m.sortField + 1) % 3
			m.applyFilters()
			return m, nil
		case key.Matches(msg, m.keyMap.Filter):
			cmd := m.filter.Activate()
			return m, cmd
		}
	}

	// Forward remaining messages to table.
	var cmd tea.Cmd
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

// View renders the events list.
func (m Model) View() string {
	if !m.ready {
		return "Loading..."
	}

	// Header: title + severity filter indicator.
	title := ui.TitleStyle.Render("Cluster Events")
	filterLabel := fmt.Sprintf("[%s]", severityFilterLabel(m.severityFilter))
	header := title + " " + lipgloss.NewStyle().Foreground(ui.ColorMuted).Render(filterLabel)

	// Filter bar.
	filterView := m.filter.View()

	// Table.
	tableView := m.table.View()

	// Footer: row count + sort indicator.
	showing := fmt.Sprintf("Showing %d of %d events", len(m.filteredEvents), len(m.allEvents))
	sortIndicator := fmt.Sprintf("Sort: %s", sortFieldLabel(m.sortField))
	footer := lipgloss.NewStyle().Foreground(ui.ColorMuted).Render(showing + "  " + sortIndicator)

	if len(m.allEvents) == 0 {
		return lipgloss.JoinVertical(lipgloss.Left,
			header,
			"",
			lipgloss.NewStyle().Foreground(ui.ColorMuted).Render("No cluster events"),
		)
	}

	// Hotkey hints.
	hints := ui.RenderHints([]ui.HintPair{
		ui.BindingHint(Keys.SeverityError), ui.BindingHint(Keys.SeverityWarning),
		ui.BindingHint(Keys.SeverityInfo), ui.BindingHint(Keys.SeverityAll),
		ui.BindingHint(Keys.Sort), ui.BindingHint(Keys.Filter),
	})

	parts := []string{header}
	if filterView != "" {
		parts = append(parts, filterView)
	}
	parts = append(parts, tableView, footer, hints)

	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}

// FilterActive returns whether the text filter input is currently active.
func (m Model) FilterActive() bool {
	return m.filter.Active()
}
