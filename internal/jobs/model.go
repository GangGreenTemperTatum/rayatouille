package jobs

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

// SelectJobMsg is sent when the user presses Enter on a job in the list.
// The root app model handles this message to push the job detail view.
type SelectJobMsg struct {
	Job ray.JobDetail
}

// StatusFilter defines which job statuses to display.
type StatusFilter int

const (
	StatusAll StatusFilter = iota
	StatusRunning
	StatusFailed
	StatusPending
	StatusSucceeded
)

// statusFilterLabel returns the display label for a status filter.
func statusFilterLabel(sf StatusFilter) string {
	switch sf {
	case StatusRunning:
		return "RUNNING"
	case StatusFailed:
		return "FAILED"
	case StatusPending:
		return "PENDING"
	case StatusSucceeded:
		return "SUCCEEDED"
	default:
		return "ALL"
	}
}

// SortField defines which column to sort by.
type SortField int

const (
	SortByAge SortField = iota
	SortByStatus
	SortByEntrypoint
	SortByDuration
)

// sortFieldLabel returns the display label for a sort field.
func sortFieldLabel(sf SortField) string {
	switch sf {
	case SortByStatus:
		return "Status"
	case SortByEntrypoint:
		return "Entrypoint"
	case SortByDuration:
		return "Duration"
	default:
		return "Age"
	}
}

// SortOrder defines the sort direction.
type SortOrder int

const (
	SortDesc SortOrder = iota
	SortAsc
)

// Model is the jobs list view model.
type Model struct {
	table        table.Model
	filter       ui.FilterModel
	keyMap       KeyMap
	allJobs      []ray.JobDetail
	filteredJobs []ray.JobDetail
	statusFilter StatusFilter
	sortField    SortField
	sortOrder    SortOrder
	width        int
	height       int
	ready        bool
}

// New creates a new jobs list model.
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
		table:        t,
		filter:       ui.NewFilter(),
		keyMap:       Keys,
		statusFilter: StatusRunning,
		sortField:    SortByAge,
		sortOrder:    SortDesc,
	}
}

// SetSize updates the jobs view dimensions.
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

// SetJobs updates the job list from the root model's shared data.
func (m *Model) SetJobs(jobs []ray.JobDetail) {
	m.allJobs = jobs
	m.applyFilters()
}

// applyFilters applies status filter, text filter, and sorting, then updates table rows.
func (m *Model) applyFilters() {
	// Start with all jobs.
	filtered := make([]ray.JobDetail, 0, len(m.allJobs))
	for _, j := range m.allJobs {
		// Status filter.
		if m.statusFilter != StatusAll {
			want := statusFilterLabel(m.statusFilter)
			if !strings.EqualFold(j.Status, want) {
				continue
			}
		}
		// Text filter (match on SubmissionID or Entrypoint).
		if !m.filter.Matches(j.SubmissionID) && !m.filter.Matches(j.Entrypoint) {
			continue
		}
		filtered = append(filtered, j)
	}

	// Sort.
	sort.Slice(filtered, func(i, j int) bool {
		cmp := m.compareJobs(filtered[i], filtered[j])
		if m.sortOrder == SortDesc {
			return cmp > 0
		}
		return cmp < 0
	})

	m.filteredJobs = filtered

	// Save cursor, update rows, restore cursor.
	cursor := m.table.Cursor()
	rows := make([]table.Row, len(filtered))
	for i, j := range filtered {
		rows[i] = JobToRow(j)
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

// compareJobs compares two jobs for sorting. Returns <0, 0, or >0.
func (m *Model) compareJobs(a, b ray.JobDetail) int {
	switch m.sortField {
	case SortByStatus:
		return strings.Compare(a.Status, b.Status)
	case SortByEntrypoint:
		return strings.Compare(a.Entrypoint, b.Entrypoint)
	case SortByDuration:
		da := jobDurationMs(a)
		db := jobDurationMs(b)
		if da < db {
			return -1
		}
		if da > db {
			return 1
		}
		return 0
	default: // SortByAge -- by StartTime.
		if a.StartTime < b.StartTime {
			return -1
		}
		if a.StartTime > b.StartTime {
			return 1
		}
		return 0
	}
}

// jobDurationMs returns the duration of a job in milliseconds.
func jobDurationMs(j ray.JobDetail) int64 {
	if j.StartTime <= 0 {
		return 0
	}
	if j.EndTime > 0 {
		return j.EndTime - j.StartTime
	}
	return 0
}

// Update handles messages for the jobs list.
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
		case key.Matches(msg, m.keyMap.StatusFailed):
			m.statusFilter = StatusFailed
			m.applyFilters()
			return m, nil
		case key.Matches(msg, m.keyMap.StatusPending):
			m.statusFilter = StatusPending
			m.applyFilters()
			return m, nil
		case key.Matches(msg, m.keyMap.StatusSucceeded):
			m.statusFilter = StatusSucceeded
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
			selected := m.SelectedJob()
			if selected == nil {
				return m, nil
			}
			job := *selected
			return m, func() tea.Msg { return SelectJobMsg{Job: job} }
		}
	}

	// Forward remaining messages to table.
	var cmd tea.Cmd
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

// View renders the jobs list.
func (m Model) View() string {
	if !m.ready {
		return "Loading..."
	}

	if len(m.allJobs) == 0 {
		title := ui.TitleStyle.Render("Jobs")
		empty := lipgloss.NewStyle().Foreground(ui.ColorMuted).Render("No jobs found")
		return lipgloss.JoinVertical(lipgloss.Left, title, "", empty)
	}

	// Header: title + status filter indicator.
	title := ui.TitleStyle.Render("Jobs")
	filterLabel := fmt.Sprintf("[%s]", statusFilterLabel(m.statusFilter))
	header := title + " " + lipgloss.NewStyle().Foreground(ui.ColorMuted).Render(filterLabel)

	// Filter bar.
	filterView := m.filter.View()

	// Table.
	tableView := m.table.View()

	// Footer: row count + sort indicator.
	showing := fmt.Sprintf("Showing %d of %d jobs", len(m.filteredJobs), len(m.allJobs))
	sortIndicator := fmt.Sprintf("Sort: %s", sortFieldLabel(m.sortField))
	footer := lipgloss.NewStyle().Foreground(ui.ColorMuted).Render(showing + "  " + sortIndicator)

	// Hotkey hints (derived from keybindings so custom keys are reflected).
	hints := ui.RenderHints([]ui.HintPair{
		ui.BindingHint(Keys.StatusRunning), ui.BindingHint(Keys.StatusFailed),
		ui.BindingHint(Keys.StatusSucceeded), ui.BindingHint(Keys.StatusPending), ui.BindingHint(Keys.StatusAll),
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

// SelectedJob returns the currently selected job, or nil if the list is empty.
func (m Model) SelectedJob() *ray.JobDetail {
	if len(m.filteredJobs) == 0 {
		return nil
	}
	cursor := m.table.Cursor()
	if cursor < 0 || cursor >= len(m.filteredJobs) {
		return nil
	}
	j := m.filteredJobs[cursor]
	return &j
}
