package nodes

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

// SelectNodeMsg is sent when the user presses Enter on a node in the list.
// The root app model handles this message to push the node detail view.
type SelectNodeMsg struct {
	Node ray.Node
}

// StatusFilter defines which node statuses to display.
type StatusFilter int

const (
	StatusAll StatusFilter = iota
	StatusAlive
	StatusDead
)

// statusFilterLabel returns the display label for a status filter.
func statusFilterLabel(sf StatusFilter) string {
	switch sf {
	case StatusAlive:
		return "ALIVE"
	case StatusDead:
		return "DEAD"
	default:
		return "ALL"
	}
}

// SortField defines which column to sort by.
type SortField int

const (
	SortByIP SortField = iota
	SortByStatus
	SortByCPU
	SortByMemory
)

// sortFieldLabel returns the display label for a sort field.
func sortFieldLabel(sf SortField) string {
	switch sf {
	case SortByStatus:
		return "Status"
	case SortByCPU:
		return "CPU"
	case SortByMemory:
		return "Memory"
	default:
		return "IP"
	}
}

// SortOrder defines the sort direction.
type SortOrder int

const (
	SortDesc SortOrder = iota
	SortAsc
)

// Model is the nodes list view model.
type Model struct {
	table         table.Model
	filter        ui.FilterModel
	keyMap        KeyMap
	allNodes      []ray.Node
	filteredNodes []ray.Node
	statusFilter  StatusFilter
	sortField     SortField
	sortOrder     SortOrder
	width         int
	height        int
	ready         bool
}

// New creates a new nodes list model.
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
		sortField: SortByIP,
		sortOrder: SortAsc,
	}
}

// SetSize updates the nodes view dimensions.
func (m *Model) SetSize(w, h int) {
	m.width = w
	m.height = h
	m.ready = true
	// Reserve space: 1 header/filter bar, 1 footer, 1 hints bar, 1 padding.
	tableHeight := h - 5
	if tableHeight < 3 {
		tableHeight = 3
	}
	m.table.SetWidth(w)
	m.table.SetHeight(tableHeight)
	m.table.SetColumns(Columns(w))
}

// SetNodes updates the node list from the root model's shared data.
func (m *Model) SetNodes(nodes []ray.Node) {
	m.allNodes = nodes
	m.applyFilters()
}

// applyFilters applies status filter, text filter, and sorting, then updates table rows.
func (m *Model) applyFilters() {
	// Start with all nodes.
	filtered := make([]ray.Node, 0, len(m.allNodes))
	for _, n := range m.allNodes {
		// Status filter.
		if m.statusFilter != StatusAll {
			want := statusFilterLabel(m.statusFilter)
			if !strings.EqualFold(n.State, want) {
				continue
			}
		}
		// Text filter (match on NodeID and NodeIP).
		if !m.filter.Matches(n.NodeID) && !m.filter.Matches(n.NodeIP) {
			continue
		}
		filtered = append(filtered, n)
	}

	// Sort.
	sort.Slice(filtered, func(i, j int) bool {
		cmp := m.compareNodes(filtered[i], filtered[j])
		if m.sortOrder == SortDesc {
			return cmp > 0
		}
		return cmp < 0
	})

	m.filteredNodes = filtered

	// Save cursor, update rows, restore cursor.
	cursor := m.table.Cursor()
	rows := make([]table.Row, len(filtered))
	for i, n := range filtered {
		rows[i] = NodeToRow(n)
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

// compareNodes compares two nodes for sorting. Returns <0, 0, or >0.
func (m *Model) compareNodes(a, b ray.Node) int {
	switch m.sortField {
	case SortByStatus:
		return strings.Compare(a.State, b.State)
	case SortByCPU:
		ac := a.ResourcesTotal["CPU"]
		bc := b.ResourcesTotal["CPU"]
		if ac < bc {
			return -1
		}
		if ac > bc {
			return 1
		}
		return 0
	case SortByMemory:
		am := a.ResourcesTotal["memory"]
		bm := b.ResourcesTotal["memory"]
		if am < bm {
			return -1
		}
		if am > bm {
			return 1
		}
		return 0
	default: // SortByIP
		return strings.Compare(a.NodeIP, b.NodeIP)
	}
}

// Update handles messages for the nodes list.
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
		case key.Matches(msg, m.keyMap.Sort):
			m.sortField = (m.sortField + 1) % 4
			m.applyFilters()
			return m, nil
		case key.Matches(msg, m.keyMap.Filter):
			cmd := m.filter.Activate()
			return m, cmd
		case key.Matches(msg, m.keyMap.Enter):
			selected := m.SelectedNode()
			if selected == nil {
				return m, nil
			}
			node := *selected
			return m, func() tea.Msg { return SelectNodeMsg{Node: node} }
		}
	}

	// Forward remaining messages to table.
	var cmd tea.Cmd
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

// View renders the nodes list.
func (m Model) View() string {
	if !m.ready {
		return "Loading..."
	}

	if len(m.allNodes) == 0 {
		title := ui.TitleStyle.Render("Nodes")
		empty := lipgloss.NewStyle().Foreground(ui.ColorMuted).Render("No nodes found")
		return lipgloss.JoinVertical(lipgloss.Left, title, "", empty)
	}

	// Header: title + status filter indicator.
	title := ui.TitleStyle.Render("Nodes")
	filterLabel := fmt.Sprintf("[%s]", statusFilterLabel(m.statusFilter))
	header := title + " " + lipgloss.NewStyle().Foreground(ui.ColorMuted).Render(filterLabel)

	// Filter bar.
	filterView := m.filter.View()

	// Table.
	tableView := m.table.View()

	// Footer: row count + sort indicator.
	showing := fmt.Sprintf("Showing %d of %d nodes", len(m.filteredNodes), len(m.allNodes))
	sortIndicator := fmt.Sprintf("Sort: %s", sortFieldLabel(m.sortField))
	footer := lipgloss.NewStyle().Foreground(ui.ColorMuted).Render(showing + "  " + sortIndicator)

	// Hotkey hints.
	hints := ui.RenderHints([]ui.HintPair{
		ui.BindingHint(Keys.StatusAlive), ui.BindingHint(Keys.StatusDead), ui.BindingHint(Keys.StatusAll),
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

// SelectedNode returns the currently selected node, or nil if the list is empty.
func (m Model) SelectedNode() *ray.Node {
	if len(m.filteredNodes) == 0 {
		return nil
	}
	cursor := m.table.Cursor()
	if cursor < 0 || cursor >= len(m.filteredNodes) {
		return nil
	}
	n := m.filteredNodes[cursor]
	return &n
}
