package serve

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

// SelectAppMsg is sent when the user presses Enter on an application in the list.
// The root app model handles this message to push the application detail view.
type SelectAppMsg struct {
	Name string
	App  ray.ApplicationDetails
}

// StatusFilter defines which application statuses to display.
type StatusFilter int

const (
	StatusAll StatusFilter = iota
	StatusRunning
	StatusDeploying
	StatusFailed
)

// statusFilterLabel returns the display label for a status filter.
func statusFilterLabel(sf StatusFilter) string {
	switch sf {
	case StatusRunning:
		return "RUNNING"
	case StatusDeploying:
		return "DEPLOYING"
	case StatusFailed:
		return "FAILED"
	default:
		return "ALL"
	}
}

// SortField defines which column to sort by.
type SortField int

const (
	SortByName SortField = iota
	SortByStatus
	SortByRoute
)

// sortFieldLabel returns the display label for a sort field.
func sortFieldLabel(sf SortField) string {
	switch sf {
	case SortByStatus:
		return "Status"
	case SortByRoute:
		return "Route"
	default:
		return "Name"
	}
}

// SortOrder defines the sort direction.
type SortOrder int

const (
	SortDesc SortOrder = iota
	SortAsc
)

// appEntry pairs an application name (map key) with its details.
type appEntry struct {
	Name string
	App  ray.ApplicationDetails
}

// Model is the serve applications list view model.
type Model struct {
	table        table.Model
	filter       ui.FilterModel
	keyMap       KeyMap
	allApps      []appEntry
	filteredApps []appEntry
	statusFilter StatusFilter
	sortField    SortField
	sortOrder    SortOrder
	width        int
	height       int
	ready        bool
}

// New creates a new serve applications list model.
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
		sortField: SortByName,
		sortOrder: SortAsc,
	}
}

// SetSize updates the serve view dimensions.
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

// SetApps updates the applications list from the Serve instance details.
// If details is nil (Serve not running), the list is cleared gracefully.
func (m *Model) SetApps(details *ray.ServeInstanceDetails) {
	if details == nil || details.Applications == nil {
		m.allApps = nil
		m.applyFilters()
		return
	}

	apps := make([]appEntry, 0, len(details.Applications))
	for name, app := range details.Applications {
		apps = append(apps, appEntry{Name: name, App: app})
	}
	// Sort by name for stable ordering before filters.
	sort.Slice(apps, func(i, j int) bool {
		return apps[i].Name < apps[j].Name
	})
	m.allApps = apps
	m.applyFilters()
}

// applyFilters applies status filter, text filter, and sorting, then updates table rows.
func (m *Model) applyFilters() {
	filtered := make([]appEntry, 0, len(m.allApps))
	for _, e := range m.allApps {
		// Status filter.
		if m.statusFilter != StatusAll {
			switch m.statusFilter {
			case StatusRunning:
				if e.App.Status != "RUNNING" {
					continue
				}
			case StatusDeploying:
				if e.App.Status != "DEPLOYING" && e.App.Status != "NOT_STARTED" {
					continue
				}
			case StatusFailed:
				if e.App.Status != "DEPLOY_FAILED" && e.App.Status != "UNHEALTHY" && e.App.Status != "DELETING" {
					continue
				}
			}
		}

		// Text filter (match on Name, Status, RoutePrefix).
		route := ""
		if e.App.RoutePrefix != nil {
			route = *e.App.RoutePrefix
		}
		if !m.filter.Matches(e.Name) && !m.filter.Matches(e.App.Status) && !m.filter.Matches(route) {
			continue
		}

		filtered = append(filtered, e)
	}

	// Sort.
	sort.Slice(filtered, func(i, j int) bool {
		cmp := m.compareApps(filtered[i], filtered[j])
		if m.sortOrder == SortDesc {
			return cmp > 0
		}
		return cmp < 0
	})

	m.filteredApps = filtered

	// Save cursor, update rows, restore cursor.
	cursor := m.table.Cursor()
	rows := make([]table.Row, len(filtered))
	for i, e := range filtered {
		rows[i] = AppToRow(e.Name, e.App)
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

// compareApps compares two app entries for sorting. Returns <0, 0, or >0.
func (m *Model) compareApps(a, b appEntry) int {
	switch m.sortField {
	case SortByStatus:
		return strings.Compare(a.App.Status, b.App.Status)
	case SortByRoute:
		ra, rb := "", ""
		if a.App.RoutePrefix != nil {
			ra = *a.App.RoutePrefix
		}
		if b.App.RoutePrefix != nil {
			rb = *b.App.RoutePrefix
		}
		return strings.Compare(ra, rb)
	default: // SortByName
		return strings.Compare(a.Name, b.Name)
	}
}

// Update handles messages for the serve applications list.
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
		case key.Matches(msg, m.keyMap.StatusRunning):
			m.statusFilter = StatusRunning
			m.applyFilters()
			return m, nil
		case key.Matches(msg, m.keyMap.StatusDeploying):
			m.statusFilter = StatusDeploying
			m.applyFilters()
			return m, nil
		case key.Matches(msg, m.keyMap.StatusFailed):
			m.statusFilter = StatusFailed
			m.applyFilters()
			return m, nil
		case key.Matches(msg, m.keyMap.Sort):
			m.sortField = (m.sortField + 1) % 3
			m.applyFilters()
			return m, nil
		case key.Matches(msg, m.keyMap.Filter):
			cmd := m.filter.Activate()
			return m, cmd
		case key.Matches(msg, m.keyMap.Enter):
			selected := m.selectedApp()
			if selected == nil {
				return m, nil
			}
			entry := *selected
			return m, func() tea.Msg { return SelectAppMsg(entry) }
		}
	}

	// Forward remaining messages to table.
	var cmd tea.Cmd
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

// View renders the serve applications list.
func (m Model) View() string {
	if !m.ready {
		return "Loading..."
	}

	// Empty state.
	if len(m.allApps) == 0 && !m.filter.Active() && m.filter.Value() == "" {
		return lipgloss.JoinVertical(lipgloss.Left,
			ui.TitleStyle.Render("Serve Applications"),
			"",
			lipgloss.NewStyle().Foreground(ui.ColorMuted).Render("No Serve applications deployed"),
		)
	}

	// Header: title + status filter indicator.
	title := ui.TitleStyle.Render("Serve Applications")
	filterLabel := fmt.Sprintf("[%s]", statusFilterLabel(m.statusFilter))
	header := title + " " + lipgloss.NewStyle().Foreground(ui.ColorMuted).Render(filterLabel)

	// Filter bar.
	filterView := m.filter.View()

	// Table.
	tableView := m.table.View()

	// Footer: row count + sort indicator.
	showing := fmt.Sprintf("Showing %d of %d applications", len(m.filteredApps), len(m.allApps))
	sortIndicator := fmt.Sprintf("Sort: %s", sortFieldLabel(m.sortField))
	footer := lipgloss.NewStyle().Foreground(ui.ColorMuted).Render(showing + "  " + sortIndicator)

	// Hotkey hints.
	hints := ui.RenderHints([]ui.HintPair{
		ui.BindingHint(Keys.StatusRunning), ui.BindingHint(Keys.StatusDeploying),
		ui.BindingHint(Keys.StatusFailed), ui.BindingHint(Keys.StatusAll),
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

// selectedApp returns the currently selected app entry, or nil if the list is empty.
func (m Model) selectedApp() *appEntry {
	if len(m.filteredApps) == 0 {
		return nil
	}
	cursor := m.table.Cursor()
	if cursor < 0 || cursor >= len(m.filteredApps) {
		return nil
	}
	e := m.filteredApps[cursor]
	return &e
}
