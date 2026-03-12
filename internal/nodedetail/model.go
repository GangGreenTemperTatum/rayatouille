package nodedetail

import (
	"context"
	"strings"
	"time"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/GangGreenTemperTatum/rayatouille/internal/ray"
	"github.com/GangGreenTemperTatum/rayatouille/internal/ui"
)

// Section identifies the active tab in the node detail view.
type Section int

const (
	// SectionInfo shows node metadata and resource usage.
	SectionInfo Section = iota
	// SectionActors shows actors running on this node.
	SectionActors
	// SectionLogs shows node log file browser and viewer.
	SectionLogs
)

// sectionCount is the number of sections for tab cycling.
const sectionCount = 3

// NodeActorsMsg carries the result of an async actors fetch for this node.
type NodeActorsMsg struct {
	Actors []ray.Actor
	Err    error
}

// NodeLogListMsg carries the result of an async node log listing fetch.
type NodeLogListMsg struct {
	Listing *ray.NodeLogListing
	Err     error
}

// NodeLogContentMsg carries the result of an async node log file content fetch.
type NodeLogContentMsg struct {
	Content  string
	Filename string
	Err      error
}

// LogFile represents a flattened log file entry with its category.
type LogFile struct {
	Category string
	Filename string
}

// Model is the node detail view model.
type Model struct {
	node   ray.Node
	client ray.Client

	section Section

	// Actors tab state.
	actors        []ray.Actor
	actorsErr     error
	actorsLoading bool

	// Logs tab state.
	logListing        *ray.NodeLogListing
	logListErr        error
	logListLoading    bool
	logFiles          []LogFile
	selectedLogFile   int
	logContent        string
	logContentErr     error
	logContentLoading bool
	logViewer         ui.LogViewer
	viewingLog        bool

	width  int
	height int
	ready  bool
	keyMap KeyMap
}

// New creates a new node detail model for the given node.
func New(client ray.Client, node ray.Node) Model {
	return Model{
		node:           node,
		client:         client,
		section:        SectionInfo,
		actorsLoading:  true,
		logListLoading: strings.EqualFold(node.State, "ALIVE"),
		logViewer:      ui.NewLogViewer(80, 20),
		keyMap:         Keys,
	}
}

// Init returns commands that fetch actors and log listing for this node.
func (m Model) Init() tea.Cmd {
	cmds := []tea.Cmd{m.fetchActors()}

	if strings.EqualFold(m.node.State, "ALIVE") {
		cmds = append(cmds, m.fetchLogList())
	}

	return tea.Batch(cmds...)
}

// fetchActors returns a command that fetches actors filtered by this node.
func (m Model) fetchActors() tea.Cmd {
	client := m.client
	nodeID := m.node.NodeID
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		allActors, err := client.ListActors(ctx)
		if err != nil {
			return NodeActorsMsg{Actors: nil, Err: err}
		}
		var nodeActors []ray.Actor
		for _, a := range allActors {
			if a.NodeID == nodeID {
				nodeActors = append(nodeActors, a)
			}
		}
		return NodeActorsMsg{Actors: nodeActors, Err: nil}
	}
}

// fetchLogList returns a command that fetches the node's log file listing.
func (m Model) fetchLogList() tea.Cmd {
	client := m.client
	nodeID := m.node.NodeID
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		listing, err := client.ListNodeLogs(ctx, nodeID)
		return NodeLogListMsg{Listing: listing, Err: err}
	}
}

// fetchLogContent returns a command that fetches a specific log file's content.
func (m Model) fetchLogContent(filename string) tea.Cmd {
	client := m.client
	nodeID := m.node.NodeID
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		content, err := client.GetNodeLogFile(ctx, nodeID, filename)
		return NodeLogContentMsg{Content: content, Filename: filename, Err: err}
	}
}

// Update handles messages for the node detail view.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case NodeActorsMsg:
		m.actorsLoading = false
		m.actors = msg.Actors
		m.actorsErr = msg.Err
		return m, nil

	case NodeLogListMsg:
		m.logListLoading = false
		if msg.Err != nil {
			m.logListErr = msg.Err
		} else if msg.Listing != nil {
			m.logListing = msg.Listing
			m.logFiles = flattenLogFiles(msg.Listing.Categories)
		}
		return m, nil

	case NodeLogContentMsg:
		m.logContentLoading = false
		m.logContent = msg.Content
		m.logContentErr = msg.Err
		if msg.Err == nil && msg.Content != "" {
			m.logViewer.SetContent(msg.Content)
		}
		return m, nil

	case tea.KeyPressMsg:
		// When viewing a log file and log viewer is searching, forward all keys to it.
		if m.viewingLog && m.section == SectionLogs && m.logViewer.Searching() {
			var cmd tea.Cmd
			m.logViewer, cmd = m.logViewer.Update(msg)
			return m, cmd
		}

		switch {
		case key.Matches(msg, m.keyMap.Tab):
			m.section = Section((int(m.section) + 1) % sectionCount)
			return m, nil

		case key.Matches(msg, m.keyMap.Refresh):
			m.actorsLoading = true
			if strings.EqualFold(m.node.State, "ALIVE") {
				m.logListLoading = true
			}
			m.viewingLog = false
			return m, m.Init()

		case key.Matches(msg, m.keyMap.SelectFile):
			if m.section == SectionLogs && !m.viewingLog && len(m.logFiles) > 0 {
				// Open selected log file.
				lf := m.logFiles[m.selectedLogFile]
				m.viewingLog = true
				m.logContentLoading = true
				m.logViewer = ui.NewLogViewer(m.width, m.height-4)
				return m, m.fetchLogContent(lf.Filename)
			}
			return m, nil
		}

		// Esc when viewing log content goes back to file list.
		if msg.Code == tea.KeyEscape && m.viewingLog && m.section == SectionLogs {
			m.viewingLog = false
			return m, nil
		}

		// Navigate log file list with j/k when in logs tab and not viewing a file.
		if m.section == SectionLogs && !m.viewingLog {
			switch msg.Text {
			case "j":
				if m.selectedLogFile < len(m.logFiles)-1 {
					m.selectedLogFile++
				}
				return m, nil
			case "k":
				if m.selectedLogFile > 0 {
					m.selectedLogFile--
				}
				return m, nil
			}
		}

		// Forward keys to log viewer when viewing log content.
		if m.viewingLog && m.section == SectionLogs {
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

// View renders the node detail view.
func (m Model) View() string {
	// Tab bar.
	infoLabel := " Info "
	actorsLabel := " Actors "
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
		renderTab(actorsLabel, SectionActors) + " " +
		renderTab(logsLabel, SectionLogs)

	// Section content.
	var content string
	switch m.section {
	case SectionInfo:
		content = renderInfo(m.node, m.width)
	case SectionActors:
		content = renderActors(m.actors, m.actorsLoading, m.actorsErr, m.width)
	case SectionLogs:
		content = renderLogContent(m)
	}

	// Hotkey hints for detail view.
	hints := ui.RenderHints([]ui.HintPair{
		{Key: "tab", Desc: "section"}, {Key: "j/k", Desc: "scroll"}, {Key: "space/b", Desc: "page"},
		{Key: "d/u", Desc: "half page"}, {Key: "G/g", Desc: "bottom/top"}, {Key: "/", Desc: "search"},
		{Key: "y", Desc: "copy line"}, {Key: "Y", Desc: "copy visible"}, {Key: "r", Desc: "refresh"},
	})

	return lipgloss.JoinVertical(lipgloss.Left, tabBar, "", content, hints)
}
