package actordetail

import (
	"context"
	"time"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/GangGreenTemperTatum/rayatouille/internal/ray"
	"github.com/GangGreenTemperTatum/rayatouille/internal/ui"
)

// Section identifies the active tab in the actor detail view.
type Section int

const (
	// SectionInfo shows actor metadata and death cause.
	SectionInfo Section = iota
	// SectionLogs shows actor logs via the shared LogViewer.
	SectionLogs
)

// sectionCount is the number of sections for tab cycling.
const sectionCount = 2

// GoToJobMsg is sent when the user requests navigation to the actor's parent job.
type GoToJobMsg struct {
	JobID string
}

// GoToNodeMsg is sent when the user requests navigation to the actor's host node.
type GoToNodeMsg struct {
	NodeID string
}

// ActorLogMsg carries the result of an async actor log fetch.
type ActorLogMsg struct {
	Content string
	Err     error
}

// Model is the actor detail view model.
type Model struct {
	actor  ray.Actor
	client ray.Client

	section Section

	// Logs tab state.
	logContent string
	logErr     error
	logLoading bool
	logViewer  ui.LogViewer

	width  int
	height int
	ready  bool
	keyMap KeyMap
}

// New creates a new actor detail model for the given actor.
func New(client ray.Client, actor ray.Actor) Model {
	return Model{
		actor:      actor,
		client:     client,
		section:    SectionInfo,
		logLoading: true,
		logViewer:  ui.NewLogViewer(80, 20),
		keyMap:     Keys,
	}
}

// Init returns a command that fetches actor logs.
func (m Model) Init() tea.Cmd {
	return m.fetchActorLogs()
}

// fetchActorLogs returns a command that fetches the actor's stdout logs.
func (m Model) fetchActorLogs() tea.Cmd {
	client := m.client
	actorID := m.actor.ActorID
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		content, err := client.GetActorLogs(ctx, actorID)
		return ActorLogMsg{Content: content, Err: err}
	}
}

// Update handles messages for the actor detail view.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case ActorLogMsg:
		m.logLoading = false
		m.logContent = msg.Content
		m.logErr = msg.Err
		if msg.Err == nil && msg.Content != "" {
			m.logViewer.SetContent(msg.Content)
		}
		return m, nil

	case tea.KeyPressMsg:
		// When log viewer is searching on Logs section, forward all keys to it.
		if m.logViewer.Searching() && m.section == SectionLogs {
			var cmd tea.Cmd
			m.logViewer, cmd = m.logViewer.Update(msg)
			return m, cmd
		}

		switch {
		case key.Matches(msg, m.keyMap.Tab):
			m.section = Section((int(m.section) + 1) % sectionCount)
			return m, nil

		case key.Matches(msg, m.keyMap.Refresh):
			m.logLoading = true
			return m, m.fetchActorLogs()

		case key.Matches(msg, m.keyMap.GoToJob):
			return m, func() tea.Msg {
				return GoToJobMsg{JobID: m.actor.JobID}
			}
		}

		// GoToNode only on Info section.
		if m.section == SectionInfo && key.Matches(msg, m.keyMap.GoToNode) {
			return m, func() tea.Msg {
				return GoToNodeMsg{NodeID: m.actor.NodeID}
			}
		}

		// Forward keys to log viewer when on Logs section.
		if m.section == SectionLogs {
			var cmd tea.Cmd
			m.logViewer, cmd = m.logViewer.Update(msg)
			return m, cmd
		}
	}

	return m, nil
}

// SetSize updates the view dimensions.
func (m *Model) SetSize(w, h int) {
	m.width = w
	m.height = h
	m.ready = true
	// Reserve space for tab bar (1 line) + empty line separator (1 line).
	m.logViewer.SetSize(w, h-2)
}

// View renders the actor detail view.
func (m Model) View() string {
	// Tab bar.
	infoLabel := " Info "
	logsLabel := " Logs "

	activeStyle := lipgloss.NewStyle().Bold(true).Foreground(ui.ColorPrimary)
	inactiveStyle := lipgloss.NewStyle().Foreground(ui.ColorMuted)

	renderTab := func(label string, section Section) string {
		if m.section == section {
			return activeStyle.Render("[" + label + "]")
		}
		return inactiveStyle.Render("[" + label + "]")
	}

	tabBar := renderTab(infoLabel, SectionInfo) + " " +
		renderTab(logsLabel, SectionLogs)

	// Section content.
	var content string
	switch m.section {
	case SectionInfo:
		content = renderInfo(m.actor, m.width)
	case SectionLogs:
		content = renderLogs(m)
	}

	// Hotkey hints for detail view.
	hints := ui.RenderHints([]ui.HintPair{
		{Key: "tab", Desc: "section"}, {Key: "j/k", Desc: "scroll"}, {Key: "space/b", Desc: "page"},
		{Key: "d/u", Desc: "half page"}, {Key: "G/g", Desc: "bottom/top"}, {Key: "/", Desc: "search"},
		{Key: "y", Desc: "copy line"}, {Key: "Y", Desc: "copy visible"}, {Key: "O", Desc: "go to node"},
	})

	return lipgloss.JoinVertical(lipgloss.Left, tabBar, "", content, hints)
}
