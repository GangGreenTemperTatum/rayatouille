package actors

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

// SelectActorMsg is sent when the user presses Enter on an actor in the list.
// The root app model handles this message to push the actor detail view.
type SelectActorMsg struct {
	Actor ray.Actor
}

// StatusFilter defines which actor statuses to display.
type StatusFilter int

const (
	StatusAll StatusFilter = iota
	StatusAlive
	StatusDead
	StatusPending
)

// statusFilterLabel returns the display label for a status filter.
func statusFilterLabel(sf StatusFilter) string {
	switch sf {
	case StatusAlive:
		return "ALIVE"
	case StatusDead:
		return "DEAD"
	case StatusPending:
		return "PENDING"
	default:
		return "ALL"
	}
}

// SortField defines which column to sort by.
type SortField int

const (
	SortByClass SortField = iota
	SortByState
	SortByPID
	SortByJobID
)

// sortFieldLabel returns the display label for a sort field.
func sortFieldLabel(sf SortField) string {
	switch sf {
	case SortByState:
		return "State"
	case SortByPID:
		return "PID"
	case SortByJobID:
		return "Job ID"
	default:
		return "Class"
	}
}

// SortOrder defines the sort direction.
type SortOrder int

const (
	SortDesc SortOrder = iota
	SortAsc
)

// Model is the actors list view model.
type Model struct {
	table          table.Model
	filter         ui.FilterModel
	keyMap         KeyMap
	allActors      []ray.Actor
	filteredActors []ray.Actor
	statusFilter   StatusFilter
	sortField      SortField
	sortOrder      SortOrder
	width          int
	height         int
	ready          bool
}

// New creates a new actors list model.
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
		sortField: SortByClass,
		sortOrder: SortAsc,
	}
}

// SetSize updates the actors view dimensions.
func (m *Model) SetSize(w, h int) {
	m.width = w
	m.height = h
	m.ready = true
	// Reserve space: 1 breadcrumb, 1 header/filter bar, 1 footer, 1 padding.
	tableHeight := h - 5
	if tableHeight < 3 {
		tableHeight = 3
	}
	m.table.SetWidth(w)
	m.table.SetHeight(tableHeight)
	m.table.SetColumns(Columns(w))
}

// SetActors updates the actor list from the root model's shared data.
func (m *Model) SetActors(actors []ray.Actor) {
	m.allActors = actors
	m.applyFilters()
}

// applyFilters applies status filter, text filter, and sorting, then updates table rows.
func (m *Model) applyFilters() {
	// Start with all actors.
	filtered := make([]ray.Actor, 0, len(m.allActors))
	for _, a := range m.allActors {
		// Status filter.
		if m.statusFilter != StatusAll {
			switch m.statusFilter {
			case StatusAlive:
				if !strings.EqualFold(a.State, "ALIVE") {
					continue
				}
			case StatusDead:
				if !strings.EqualFold(a.State, "DEAD") {
					continue
				}
			case StatusPending:
				// Match any state starting with "PENDING" (covers PENDING_CREATION, etc.)
				if !strings.HasPrefix(strings.ToUpper(a.State), "PENDING") {
					continue
				}
			}
		}
		// Text filter (match on ActorID, ClassName, Name).
		if !m.filter.Matches(a.ActorID) && !m.filter.Matches(a.ClassName) && !m.filter.Matches(a.Name) {
			continue
		}
		filtered = append(filtered, a)
	}

	// Sort.
	sort.Slice(filtered, func(i, j int) bool {
		cmp := m.compareActors(filtered[i], filtered[j])
		if m.sortOrder == SortDesc {
			return cmp > 0
		}
		return cmp < 0
	})

	m.filteredActors = filtered

	// Save cursor, update rows, restore cursor.
	cursor := m.table.Cursor()
	rows := make([]table.Row, len(filtered))
	for i, a := range filtered {
		rows[i] = ActorToRow(a)
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

// compareActors compares two actors for sorting. Returns <0, 0, or >0.
func (m *Model) compareActors(a, b ray.Actor) int {
	switch m.sortField {
	case SortByState:
		return strings.Compare(a.State, b.State)
	case SortByPID:
		if a.PID < b.PID {
			return -1
		}
		if a.PID > b.PID {
			return 1
		}
		return 0
	case SortByJobID:
		return strings.Compare(a.JobID, b.JobID)
	default: // SortByClass
		return strings.Compare(a.ClassName, b.ClassName)
	}
}

// Update handles messages for the actors list.
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
		case key.Matches(msg, m.keyMap.StatusAll):
			m.statusFilter = StatusAll
			m.applyFilters()
			return m, nil
		case key.Matches(msg, m.keyMap.StatusAlive):
			m.statusFilter = StatusAlive
			m.applyFilters()
			return m, nil
		case key.Matches(msg, m.keyMap.StatusDead):
			m.statusFilter = StatusDead
			m.applyFilters()
			return m, nil
		case key.Matches(msg, m.keyMap.StatusPending):
			m.statusFilter = StatusPending
			m.applyFilters()
			return m, nil
		case key.Matches(msg, m.keyMap.Sort):
			m.sortField = (m.sortField + 1) % 4
			m.applyFilters()
			return m, nil
		case key.Matches(msg, m.keyMap.Filter):
			cmd := m.filter.Activate()
			return m, cmd
		case key.Matches(msg, m.keyMap.Enter):
			selected := m.SelectedActor()
			if selected == nil {
				return m, nil
			}
			actor := *selected
			return m, func() tea.Msg { return SelectActorMsg{Actor: actor} }
		}
	}

	// Forward remaining messages to table.
	var cmd tea.Cmd
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

// View renders the actors list.
func (m Model) View() string {
	if !m.ready {
		return "Loading..."
	}

	if len(m.allActors) == 0 {
		title := ui.TitleStyle.Render("Actors")
		empty := lipgloss.NewStyle().Foreground(ui.ColorMuted).Render("No actors found")
		return lipgloss.JoinVertical(lipgloss.Left, title, "", empty)
	}

	// Header: title + status filter indicator.
	title := ui.TitleStyle.Render("Actors")
	filterLabel := fmt.Sprintf("[%s]", statusFilterLabel(m.statusFilter))
	header := title + " " + lipgloss.NewStyle().Foreground(ui.ColorMuted).Render(filterLabel)

	// Filter bar.
	filterView := m.filter.View()

	// Table.
	tableView := m.table.View()

	// Footer: row count + sort indicator.
	showing := fmt.Sprintf("Showing %d of %d actors", len(m.filteredActors), len(m.allActors))
	sortIndicator := fmt.Sprintf("Sort: %s", sortFieldLabel(m.sortField))
	footer := lipgloss.NewStyle().Foreground(ui.ColorMuted).Render(showing + "  " + sortIndicator)

	// Hotkey hints.
	hints := ui.RenderHints([]ui.HintPair{
		ui.BindingHint(Keys.StatusAlive), ui.BindingHint(Keys.StatusDead),
		ui.BindingHint(Keys.StatusPending), ui.BindingHint(Keys.StatusAll),
		ui.BindingHint(Keys.Sort), ui.BindingHint(Keys.Filter), ui.BindingHint(Keys.Enter),
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

// SelectedActor returns the currently selected actor, or nil if the list is empty.
func (m Model) SelectedActor() *ray.Actor {
	if len(m.filteredActors) == 0 {
		return nil
	}
	cursor := m.table.Cursor()
	if cursor < 0 || cursor >= len(m.filteredActors) {
		return nil
	}
	a := m.filteredActors[cursor]
	return &a
}
