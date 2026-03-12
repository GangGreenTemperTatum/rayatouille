package jobdetail

import (
	"context"
	"fmt"
	"time"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/GangGreenTemperTatum/rayatouille/internal/ray"
	"github.com/GangGreenTemperTatum/rayatouille/internal/ui"
)

// Section identifies the active tab in the job detail view.
type Section int

const (
	// SectionMetadata shows job metadata key-value pairs.
	SectionMetadata Section = iota
	// SectionTasks shows the task summary table.
	SectionTasks
	// SectionLogs shows the log viewer with search support.
	SectionLogs
)

// sectionCount is the number of sections for tab cycling.
const sectionCount = 3

// TaskSummaryMsg carries the result of an async task summary fetch.
type TaskSummaryMsg struct {
	Summary *ray.TaskSummaryResponse
	Err     error
}

// JobLogsMsg carries the result of an async job logs fetch.
type JobLogsMsg struct {
	Logs string
	Err  error
}

// Model is the job detail view model.
type Model struct {
	job         ray.JobDetail
	client      ray.Client
	section     Section
	taskSummary *ray.TaskSummaryResponse
	taskErr     error
	loading     bool
	logs        string
	logsErr     error
	logsLoading bool
	logViewer   ui.LogViewer
	width       int
	height      int
	ready       bool
	keyMap      KeyMap
}

// New creates a new job detail model for the given job.
func New(client ray.Client, job ray.JobDetail) Model {
	return Model{
		job:         job,
		client:      client,
		section:     SectionMetadata,
		loading:     true,
		logsLoading: true,
		logViewer:   ui.NewLogViewer(80, 20),
		keyMap:      Keys,
	}
}

// Init returns a command that fetches the task summary and logs for this job.
func (m Model) Init() tea.Cmd {
	cmds := []tea.Cmd{m.fetchLogs()}

	if m.job.JobID == nil {
		cmds = append(cmds, func() tea.Msg {
			return TaskSummaryMsg{Summary: nil, Err: nil}
		})
	} else {
		jobID := *m.job.JobID
		client := m.client
		cmds = append(cmds, func() tea.Msg {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			summary, err := client.GetTaskSummary(ctx, jobID)
			return TaskSummaryMsg{Summary: summary, Err: err}
		})
	}

	return tea.Batch(cmds...)
}

// fetchLogs returns a command that fetches job logs.
func (m Model) fetchLogs() tea.Cmd {
	submissionID := m.job.SubmissionID
	client := m.client
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		logs, err := client.GetJobLogs(ctx, submissionID)
		return JobLogsMsg{Logs: logs, Err: err}
	}
}

// Update handles messages for the job detail view.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case TaskSummaryMsg:
		m.loading = false
		m.taskSummary = msg.Summary
		m.taskErr = msg.Err
		return m, nil

	case JobLogsMsg:
		m.logsLoading = false
		m.logs = msg.Logs
		m.logsErr = msg.Err
		if msg.Err == nil && msg.Logs != "" {
			m.logViewer.SetContent(msg.Logs)
		}
		return m, nil

	case tea.KeyPressMsg:
		// When in logs section and log viewer is searching, forward all keys to it.
		if m.section == SectionLogs && m.logViewer.Searching() {
			var cmd tea.Cmd
			m.logViewer, cmd = m.logViewer.Update(msg)
			return m, cmd
		}

		switch {
		case key.Matches(msg, m.keyMap.Tab):
			m.section = Section((int(m.section) + 1) % sectionCount)
			return m, nil
		case key.Matches(msg, m.keyMap.Refresh):
			m.loading = true
			m.logsLoading = true
			return m, m.Init()
		}

		// Forward keys to log viewer when on logs tab.
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

// View renders the job detail view.
func (m Model) View() string {
	// Tab bar.
	metaLabel := " Metadata "
	tasksLabel := " Tasks "
	logsLabel := " Logs "

	activeStyle := lipgloss.NewStyle().Bold(true).Foreground(ui.ColorPrimary)
	inactiveStyle := lipgloss.NewStyle().Foreground(ui.ColorMuted)

	renderTab := func(label string, section Section) string {
		if m.section == section {
			return activeStyle.Render("[" + label + "]")
		}
		return inactiveStyle.Render("[" + label + "]")
	}

	tabBar := renderTab(metaLabel, SectionMetadata) + " " +
		renderTab(tasksLabel, SectionTasks) + " " +
		renderTab(logsLabel, SectionLogs)

	// Section content.
	var content string
	switch m.section {
	case SectionMetadata:
		content = renderMetadata(m.job, m.width)
	case SectionTasks:
		content = renderTasks(m.taskSummary, m.loading, m.taskErr, m.width)
	case SectionLogs:
		content = m.renderLogs()
	}

	// Hotkey hints for detail view.
	hints := ui.RenderHints([]ui.HintPair{
		ui.BindingHint(Keys.Tab), {Key: "j/k", Desc: "scroll"}, ui.BindingHint(Keys.PageDown),
		ui.BindingHint(Keys.HalfDown), ui.BindingHint(Keys.Bottom), ui.BindingHint(Keys.Search),
		{Key: "y", Desc: "copy line"}, {Key: "Y", Desc: "copy visible"}, ui.BindingHint(Keys.Refresh),
	})

	return lipgloss.JoinVertical(lipgloss.Left, tabBar, "", content, hints)
}

// renderLogs renders the logs section content.
func (m Model) renderLogs() string {
	if m.logsLoading {
		return lipgloss.NewStyle().Foreground(ui.ColorMuted).Render("Fetching logs...")
	}

	if m.logsErr != nil {
		return lipgloss.NewStyle().Foreground(ui.ColorDanger).Render(
			fmt.Sprintf("Error fetching logs: %s", m.logsErr.Error()),
		)
	}

	return m.logViewer.View()
}
