package servedetail

import (
	"sort"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/table"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/GangGreenTemperTatum/rayatouille/internal/ray"
	"github.com/GangGreenTemperTatum/rayatouille/internal/ui"
)

// Section identifies which sub-view is active.
type Section int

const (
	// SectionDeployments shows the deployments table.
	SectionDeployments Section = iota
	// SectionReplicas shows the replicas table for a selected deployment.
	SectionReplicas
)

// Model is the serve application detail view model.
type Model struct {
	appName string
	app     ray.ApplicationDetails
	section Section

	deploymentsTable       table.Model
	replicasTable          table.Model
	selectedDeployment     *ray.DeploymentDetails
	selectedDeploymentName string

	// sortedDeploymentNames maintains deterministic ordering for cursor-to-map lookup.
	sortedDeploymentNames []string

	width  int
	height int
	keyMap KeyMap
}

// New creates a new serve detail model for the given application.
func New(name string, app ray.ApplicationDetails) Model {
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

	// Build sorted deployment names for deterministic ordering.
	names := make([]string, 0, len(app.Deployments))
	for n := range app.Deployments {
		names = append(names, n)
	}
	sort.Strings(names)

	// Build deployment rows.
	rows := make([]table.Row, len(names))
	for i, n := range names {
		d := app.Deployments[n]
		rows[i] = DeploymentToRow(n, d)
	}

	dt := table.New(
		table.WithColumns(DeploymentColumns(80)),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithStyles(styles),
		table.WithKeyMap(km),
	)

	rt := table.New(
		table.WithColumns(ReplicaColumns(80)),
		table.WithFocused(true),
		table.WithStyles(styles),
		table.WithKeyMap(km),
	)

	return Model{
		appName:               name,
		app:                   app,
		section:               SectionDeployments,
		deploymentsTable:      dt,
		replicasTable:         rt,
		sortedDeploymentNames: names,
		keyMap:                Keys,
	}
}

// SetSize updates the view dimensions.
func (m *Model) SetSize(w, h int) {
	m.width = w
	m.height = h
	// Metadata takes ~8 lines, section header 2 lines.
	tableHeight := h - 10
	if tableHeight < 3 {
		tableHeight = 3
	}
	m.deploymentsTable.SetWidth(w)
	m.deploymentsTable.SetHeight(tableHeight)
	m.deploymentsTable.SetColumns(DeploymentColumns(w))
	m.replicasTable.SetWidth(w)
	m.replicasTable.SetHeight(tableHeight)
	m.replicasTable.SetColumns(ReplicaColumns(w))
}

// Update handles messages for the serve detail view.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if msg, ok := msg.(tea.KeyPressMsg); ok {
		switch m.section {
		case SectionDeployments:
			if key.Matches(msg, m.keyMap.Enter) {
				return m.drillIntoDeployment()
			}
			var cmd tea.Cmd
			m.deploymentsTable, cmd = m.deploymentsTable.Update(msg)
			return m, cmd

		case SectionReplicas:
			if key.Matches(msg, m.keyMap.Back) {
				m.section = SectionDeployments
				m.selectedDeployment = nil
				m.selectedDeploymentName = ""
				return m, nil
			}
			var cmd tea.Cmd
			m.replicasTable, cmd = m.replicasTable.Update(msg)
			return m, cmd
		}
	}

	return m, nil
}

// drillIntoDeployment switches to the replicas view for the selected deployment.
func (m Model) drillIntoDeployment() (Model, tea.Cmd) {
	if len(m.sortedDeploymentNames) == 0 {
		return m, nil
	}

	cursor := m.deploymentsTable.Cursor()
	if cursor < 0 || cursor >= len(m.sortedDeploymentNames) {
		return m, nil
	}

	name := m.sortedDeploymentNames[cursor]
	d, ok := m.app.Deployments[name]
	if !ok {
		return m, nil
	}

	m.selectedDeployment = &d
	m.selectedDeploymentName = name
	m.section = SectionReplicas

	// Populate replicas table.
	rows := make([]table.Row, len(d.Replicas))
	for i, r := range d.Replicas {
		rows[i] = ReplicaToRow(r)
	}
	m.replicasTable.SetRows(rows)
	m.replicasTable.SetCursor(0)

	return m, nil
}

// View renders the serve detail view.
func (m Model) View() string {
	// Always show metadata at top.
	meta := renderMetadata(m.app, m.appName, m.width)

	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(ui.ColorFg)

	var sectionContent string
	switch m.section {
	case SectionDeployments:
		if len(m.sortedDeploymentNames) == 0 {
			sectionContent = lipgloss.JoinVertical(lipgloss.Left,
				headerStyle.Render("Deployments"),
				"",
				lipgloss.NewStyle().Foreground(ui.ColorMuted).Render("No deployments"),
			)
		} else {
			sectionContent = lipgloss.JoinVertical(lipgloss.Left,
				headerStyle.Render("Deployments"),
				"",
				m.deploymentsTable.View(),
			)
		}
	case SectionReplicas:
		title := headerStyle.Render("Deployment: " + m.selectedDeploymentName)
		if m.selectedDeployment != nil && len(m.selectedDeployment.Replicas) == 0 {
			sectionContent = lipgloss.JoinVertical(lipgloss.Left,
				title,
				"",
				lipgloss.NewStyle().Foreground(ui.ColorMuted).Render("No replicas"),
			)
		} else {
			sectionContent = lipgloss.JoinVertical(lipgloss.Left,
				title,
				"",
				m.replicasTable.View(),
			)
		}
	}

	return lipgloss.JoinVertical(lipgloss.Left, meta, "", sectionContent)
}
