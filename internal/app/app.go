package app

import (
	"context"
	"fmt"
	"strings"
	"time"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/GangGreenTemperTatum/rayatouille/internal/actordetail"
	"github.com/GangGreenTemperTatum/rayatouille/internal/actors"
	"github.com/GangGreenTemperTatum/rayatouille/internal/config"
	"github.com/GangGreenTemperTatum/rayatouille/internal/dashboard"
	"github.com/GangGreenTemperTatum/rayatouille/internal/events"
	"github.com/GangGreenTemperTatum/rayatouille/internal/jobdetail"
	"github.com/GangGreenTemperTatum/rayatouille/internal/jobs"
	"github.com/GangGreenTemperTatum/rayatouille/internal/nodedetail"
	"github.com/GangGreenTemperTatum/rayatouille/internal/nodes"
	"github.com/GangGreenTemperTatum/rayatouille/internal/ray"
	"github.com/GangGreenTemperTatum/rayatouille/internal/serve"
	"github.com/GangGreenTemperTatum/rayatouille/internal/servedetail"
	"github.com/GangGreenTemperTatum/rayatouille/internal/ui"
)

// Model is the root Bubble Tea model for rayatouille.
type Model struct {
	client          ray.Client
	config          *config.Config
	version         *ray.VersionInfo
	width           int
	height          int
	nav             NavStack
	showHelp        bool
	helpModel       help.Model
	dashboard       dashboard.Model
	jobsView        jobs.Model
	jobDetailView   jobdetail.Model
	nodesView       nodes.Model
	nodeDetailView  nodedetail.Model
	actorsView      actors.Model
	actorDetailView actordetail.Model
	serveView       serve.Model
	serveDetailView servedetail.Model
	eventsView      events.Model
	command         ui.CommandModel

	// Shared data cache -- root owns polling, child views read from here.
	cachedJobs   []ray.JobDetail
	cachedNodes  []ray.Node
	cachedActors []ray.Actor
	cachedServe  *ray.ServeInstanceDetails
	cachedEvents []ray.ClusterEvent

	// Polling state.
	refreshing      bool
	refreshInterval time.Duration

	// Status bar state (global, shared across views).
	lastErr     error
	lastLatency time.Duration
	lastUpdate  time.Time

	// Active profile name (shown in status bar).
	activeProfile string
}

// New creates a new root Model with the given client, config, and version info.
func New(client ray.Client, cfg *config.Config, version *ray.VersionInfo) Model {
	// Try to detect active profile name for status bar display.
	var activeProfile string
	if pc, err := config.LoadProfileConfig(); err == nil && pc.ActiveProfile != "" {
		activeProfile = pc.ActiveProfile
	}

	// Load and apply custom keybindings.
	kb := config.LoadKeybindings()
	ApplyGlobalBindings(kb.Global)
	jobs.ApplyBindings(kb.Jobs)
	nodes.ApplyBindings(kb.Nodes)
	actors.ApplyBindings(kb.Actors)
	serve.ApplyBindings(kb.Serve)
	events.ApplyBindings(kb.Events)
	jobdetail.ApplyBindings(kb.Detail, kb.Logging)

	return Model{
		client:          client,
		config:          cfg,
		version:         version,
		nav:             NewNavStack(),
		helpModel:       help.New(),
		dashboard:       dashboard.New(cfg.Address),
		jobsView:        jobs.New(),
		nodesView:       nodes.New(),
		actorsView:      actors.New(),
		serveView:       serve.New(),
		eventsView:      events.New(),
		command:         ui.NewCommand(),
		refreshInterval: cfg.RefreshInterval,
		activeProfile:   activeProfile,
	}
}

// fetchClusterData returns a command that fetches cluster data from the Ray API.
func fetchClusterData(client ray.Client) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		start := time.Now()
		nodes, nodesErr := client.ListNodes(ctx)
		jobs, jobsErr := client.ListJobDetails(ctx)
		actors, actorsErr := client.ListActors(ctx)
		serveDetails, _ := client.GetServeApplications(ctx) // Ignore error -- nil means not running
		clusterEvents, _ := client.ListClusterEvents(ctx)   // Ignore error -- nil means no events support
		latency := time.Since(start)

		var fetchErr error
		if nodesErr != nil {
			fetchErr = nodesErr
		} else if jobsErr != nil {
			fetchErr = jobsErr
		} else if actorsErr != nil {
			fetchErr = actorsErr
		}

		return ClusterDataMsg{
			Nodes:    nodes,
			Jobs:     jobs,
			Actors:   actors,
			Serve:    serveDetails,
			Events:   clusterEvents,
			FetchErr: fetchErr,
			Latency:  latency,
		}
	}
}

// doTick returns a command that sends a TickMsg after the given interval.
func doTick(interval time.Duration) tea.Cmd {
	return tea.Tick(interval, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}

// switchProfile loads the named profile and creates a new client connected to it.
func switchProfile(name string) tea.Cmd {
	return func() tea.Msg {
		pc, err := config.LoadProfileConfig()
		if err != nil {
			return ProfileSwitchedMsg{Name: name, Err: fmt.Errorf("loading config: %w", err)}
		}

		profile, ok := pc.Profiles[name]
		if !ok {
			return ProfileSwitchedMsg{Name: name, Err: fmt.Errorf("profile %q not found", name)}
		}

		client := ray.NewClient(strings.TrimRight(profile.Address, "/"), profile.TimeoutDuration())

		ctx, cancel := context.WithTimeout(context.Background(), profile.TimeoutDuration())
		defer cancel()
		version, err := client.Ping(ctx)
		if err != nil {
			return ProfileSwitchedMsg{Name: name, Err: fmt.Errorf("connecting to %s: %w", profile.Address, err)}
		}

		return ProfileSwitchedMsg{
			Name:    name,
			Client:  client,
			Version: version,
		}
	}
}

// doUITick returns a command that sends a UITickMsg every second.
func doUITick() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return UITickMsg(t)
	})
}

// Init returns the initial command for the model.
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		fetchClusterData(m.client),
		doTick(m.refreshInterval),
		doUITick(),
		m.dashboard.Init(),
	)
}

// Update handles messages and returns the updated model and any commands.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case UITickMsg:
		// Re-render every second so "Updated Xs ago" stays current.
		return m, doUITick()

	case TickMsg:
		if m.refreshing {
			return m, doTick(m.refreshInterval)
		}
		m.refreshing = true
		return m, tea.Batch(fetchClusterData(m.client), doTick(m.refreshInterval))

	case ClusterDataMsg:
		m.refreshing = false
		m.lastLatency = msg.Latency
		m.lastErr = msg.FetchErr
		// Store in shared cache.
		if msg.FetchErr == nil {
			m.cachedJobs = msg.Jobs
			m.cachedNodes = msg.Nodes
			m.cachedActors = msg.Actors
			m.lastUpdate = time.Now()
		}
		// Forward to dashboard.
		var dashCmd tea.Cmd
		m.dashboard, dashCmd = m.dashboard.Update(dashboard.ClusterDataMsg{
			Nodes:    msg.Nodes,
			Jobs:     msg.Jobs,
			FetchErr: msg.FetchErr,
			Latency:  msg.Latency,
		})
		// Forward to jobs view.
		if msg.FetchErr == nil {
			m.jobsView.SetJobs(msg.Jobs)
			m.nodesView.SetNodes(msg.Nodes)
			m.actorsView.SetActors(msg.Actors)
		}
		// Serve and events update regardless of main fetch error
		// (they handle nil data gracefully).
		m.cachedServe = msg.Serve
		m.cachedEvents = msg.Events
		m.serveView.SetApps(msg.Serve)
		m.eventsView.SetEvents(msg.Events)
		return m, dashCmd

	case ProfileSwitchedMsg:
		if msg.Err != nil {
			m.lastErr = fmt.Errorf("profile switch failed: %w", msg.Err)
			return m, nil
		}
		// Replace client and version with new connection.
		m.client = msg.Client
		m.version = msg.Version
		m.activeProfile = msg.Name

		// Clear all cached data.
		m.cachedJobs = nil
		m.cachedNodes = nil
		m.cachedActors = nil
		m.cachedServe = nil
		m.cachedEvents = nil

		// Reset child views.
		m.jobsView.SetJobs(nil)
		m.nodesView.SetNodes(nil)
		m.actorsView.SetActors(nil)
		m.serveView.SetApps(nil)
		m.eventsView.SetEvents(nil)

		// Reset navigation to dashboard.
		m.nav = NewNavStack()

		// Reset status bar state.
		m.lastUpdate = time.Time{}
		m.lastErr = nil
		m.refreshing = false

		// Trigger immediate data fetch from new cluster.
		return m, fetchClusterData(m.client)

	case jobs.SelectJobMsg:
		m.nav.Push(ViewJobDetail)
		m.jobDetailView = jobdetail.New(m.client, msg.Job)
		m.jobDetailView.SetSize(m.width, m.height-2)
		return m, m.jobDetailView.Init()

	case nodes.SelectNodeMsg:
		m.nav.Push(ViewNodeDetail)
		m.nodeDetailView = nodedetail.New(m.client, msg.Node)
		m.nodeDetailView.SetSize(m.width, m.height-2)
		return m, m.nodeDetailView.Init()

	case actors.SelectActorMsg:
		m.nav.Push(ViewActorDetail)
		m.actorDetailView = actordetail.New(m.client, msg.Actor)
		m.actorDetailView.SetSize(m.width, m.height-2)
		return m, m.actorDetailView.Init()

	case serve.SelectAppMsg:
		m.nav.Push(ViewServeDetail)
		m.serveDetailView = servedetail.New(msg.Name, msg.App)
		m.serveDetailView.SetSize(m.width, m.height-2)
		return m, nil

	case actordetail.GoToJobMsg:
		for _, j := range m.cachedJobs {
			if j.JobID != nil && *j.JobID == msg.JobID {
				m.nav.Push(ViewJobDetail)
				m.jobDetailView = jobdetail.New(m.client, j)
				m.jobDetailView.SetSize(m.width, m.height-2)
				return m, m.jobDetailView.Init()
			}
		}
		// Job not found in cache (e.g., driver job) -- stay on current view.
		return m, nil

	case actordetail.GoToNodeMsg:
		for _, n := range m.cachedNodes {
			if n.NodeID == msg.NodeID {
				m.nav.Push(ViewNodeDetail)
				m.nodeDetailView = nodedetail.New(m.client, n)
				m.nodeDetailView.SetSize(m.width, m.height-2)
				return m, m.nodeDetailView.Init()
			}
		}
		return m, nil

	case tea.KeyPressMsg:
		// If command input is active, route to it first.
		if m.command.Active() {
			var cmd tea.Cmd
			var result *ui.CommandResult
			m.command, cmd, result = m.command.Update(msg)
			if result != nil {
				cmdResult := m.handleCommandResult(*result)
				if cmdResult != nil {
					return m, tea.Batch(cmd, cmdResult)
				}
			}
			return m, cmd
		}

		// If the active child view has an active filter, let it handle keys
		// (don't intercept esc/q/? etc at the global level).
		if m.isChildFilterActive() {
			return m.routeToActiveView(msg)
		}

		// Global keybindings handled before child views.
		switch {
		case key.Matches(msg, GlobalKeys.Quit):
			return m, tea.Quit
		case key.Matches(msg, GlobalKeys.Help):
			m.showHelp = !m.showHelp
			return m, nil
		case key.Matches(msg, GlobalKeys.Back):
			if m.showHelp {
				m.showHelp = false
				return m, nil
			}
			if m.nav.Depth() > 1 {
				m.nav.Pop()
				return m, nil
			}
			return m, nil
		case key.Matches(msg, GlobalKeys.Tab):
			// Only cycle top-level views from top-level views.
			// Detail views handle Tab themselves (e.g., section cycling).
			if m.isTopLevelView() {
				return m.cycleView(), nil
			}
			// Fall through to child view.
		case msg.String() == "1" && m.isTopLevelView():
			return m.switchToView(ViewJobs), nil
		case msg.String() == "2" && m.isTopLevelView():
			return m.switchToView(ViewNodes), nil
		case msg.String() == "3" && m.isTopLevelView():
			return m.switchToView(ViewActors), nil
		case msg.String() == "4" && m.isTopLevelView():
			return m.switchToView(ViewServe), nil
		case msg.String() == "5" && m.isTopLevelView():
			return m.switchToView(ViewEvents), nil
		case msg.String() == ":":
			cmd := m.command.Activate()
			return m, cmd
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.helpModel.SetWidth(msg.Width)
		m.dashboard.SetSize(m.width, m.height)
		// Reserve space for banner (1 line) + breadcrumb (1 line) + status bar (1 line).
		m.jobsView.SetSize(m.width, m.height-3)
		m.jobDetailView.SetSize(m.width, m.height-3)
		m.nodesView.SetSize(m.width, m.height-3)
		m.nodeDetailView.SetSize(m.width, m.height-3)
		m.actorsView.SetSize(m.width, m.height-3)
		m.actorDetailView.SetSize(m.width, m.height-3)
		m.serveView.SetSize(m.width, m.height-3)
		m.serveDetailView.SetSize(m.width, m.height-3)
		m.eventsView.SetSize(m.width, m.height-3)
	}

	return m.routeToActiveView(msg)
}

// handleCommandResult processes a colon command result.
// Returns a tea.Cmd if the command triggers an async operation (e.g., profile switch).
func (m *Model) handleCommandResult(result ui.CommandResult) tea.Cmd {
	switch result.Command {
	case ui.CmdJobs:
		m.nav.Push(ViewJobs)
		m.jobsView.SetSize(m.width, m.height-3)
		if m.cachedJobs != nil {
			m.jobsView.SetJobs(m.cachedJobs)
		}
	case ui.CmdDashboard:
		// Pop back to dashboard.
		for m.nav.Current() != ViewDashboard && m.nav.Depth() > 1 {
			m.nav.Pop()
		}
	case ui.CmdNodes:
		m.nav.Push(ViewNodes)
		m.nodesView.SetSize(m.width, m.height-3)
		if m.cachedNodes != nil {
			m.nodesView.SetNodes(m.cachedNodes)
		}
	case ui.CmdActors:
		m.nav.Push(ViewActors)
		m.actorsView.SetSize(m.width, m.height-3)
		if m.cachedActors != nil {
			m.actorsView.SetActors(m.cachedActors)
		}
	case ui.CmdServe:
		m.nav.Push(ViewServe)
		m.serveView.SetSize(m.width, m.height-3)
		if m.cachedServe != nil {
			m.serveView.SetApps(m.cachedServe)
		}
	case ui.CmdEvents:
		m.nav.Push(ViewEvents)
		m.eventsView.SetSize(m.width, m.height-3)
		if m.cachedEvents != nil {
			m.eventsView.SetEvents(m.cachedEvents)
		}
	case ui.CmdProfile:
		if result.Arg == "" {
			return nil
		}
		return switchProfile(result.Arg)
	}
	return nil
}

// isChildFilterActive returns true if the active child view has an active text filter.
func (m *Model) isChildFilterActive() bool {
	switch m.nav.Current() {
	case ViewJobs:
		return m.jobsView.FilterActive()
	case ViewNodes:
		return m.nodesView.FilterActive()
	case ViewActors:
		return m.actorsView.FilterActive()
	case ViewServe:
		return m.serveView.FilterActive()
	case ViewEvents:
		return m.eventsView.FilterActive()
	}
	return false
}

// isTopLevelView returns true if the current view is a top-level (list) view.
func (m *Model) isTopLevelView() bool {
	switch m.nav.Current() {
	case ViewDashboard, ViewJobs, ViewNodes, ViewActors, ViewServe, ViewEvents:
		return true
	}
	return false
}

// topLevelViews defines the Tab cycle order for top-level views.
var topLevelViews = []View{ViewDashboard, ViewJobs, ViewNodes, ViewActors, ViewServe, ViewEvents}

// cycleView advances to the next top-level view via Tab.
func (m Model) cycleView() Model {
	cur := m.nav.Current()
	next := topLevelViews[0]
	for i, v := range topLevelViews {
		if v == cur {
			next = topLevelViews[(i+1)%len(topLevelViews)]
			break
		}
	}
	return m.switchToView(next)
}

// switchToView resets the nav stack and pushes the target top-level view.
// Re-populates the target view from cached data to ensure visibility.
func (m Model) switchToView(v View) Model {
	m.nav = NewNavStack()
	if v != ViewDashboard {
		m.nav.Push(v)
	}
	// Re-populate target view from cache so data is always visible,
	// regardless of timing between WindowSizeMsg and ClusterDataMsg.
	switch v {
	case ViewJobs:
		m.jobsView.SetSize(m.width, m.height-3)
		if m.cachedJobs != nil {
			m.jobsView.SetJobs(m.cachedJobs)
		}
	case ViewNodes:
		m.nodesView.SetSize(m.width, m.height-3)
		if m.cachedNodes != nil {
			m.nodesView.SetNodes(m.cachedNodes)
		}
	case ViewActors:
		m.actorsView.SetSize(m.width, m.height-3)
		if m.cachedActors != nil {
			m.actorsView.SetActors(m.cachedActors)
		}
	case ViewServe:
		m.serveView.SetSize(m.width, m.height-3)
		if m.cachedServe != nil {
			m.serveView.SetApps(m.cachedServe)
		}
	case ViewEvents:
		m.eventsView.SetSize(m.width, m.height-3)
		if m.cachedEvents != nil {
			m.eventsView.SetEvents(m.cachedEvents)
		}
	}
	return m
}

// routeToActiveView routes a message to the currently active view.
func (m Model) routeToActiveView(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch m.nav.Current() {
	case ViewDashboard:
		m.dashboard, cmd = m.dashboard.Update(msg)
	case ViewJobs:
		m.jobsView, cmd = m.jobsView.Update(msg)
	case ViewJobDetail:
		m.jobDetailView, cmd = m.jobDetailView.Update(msg)
	case ViewNodes:
		m.nodesView, cmd = m.nodesView.Update(msg)
	case ViewNodeDetail:
		m.nodeDetailView, cmd = m.nodeDetailView.Update(msg)
	case ViewActors:
		m.actorsView, cmd = m.actorsView.Update(msg)
	case ViewActorDetail:
		m.actorDetailView, cmd = m.actorDetailView.Update(msg)
	case ViewServe:
		m.serveView, cmd = m.serveView.Update(msg)
	case ViewServeDetail:
		m.serveDetailView, cmd = m.serveDetailView.Update(msg)
	case ViewEvents:
		m.eventsView, cmd = m.eventsView.Update(msg)
	}
	return m, cmd
}

// View renders the current state of the model.
func (m Model) View() tea.View {
	var parts []string

	// Banner with power bar on the right.
	bannerLeft := lipgloss.NewStyle().Bold(true).Foreground(ui.ColorAccent).Render("🐀 rayatouille") +
		lipgloss.NewStyle().Foreground(ui.ColorMuted).Render(" — terminal ray tracing, lightly sautéed with chaos where anyone can cook")
	powerBar := m.renderPowerBar()
	if powerBar != "" && m.width > 0 {
		leftWidth := lipgloss.Width(bannerLeft)
		rightWidth := lipgloss.Width(powerBar)
		gap := m.width - leftWidth - rightWidth
		if gap < 2 {
			gap = 2
		}
		parts = append(parts, bannerLeft+strings.Repeat(" ", gap)+powerBar)
	} else {
		parts = append(parts, bannerLeft)
	}

	// Breadcrumbs shown only when navigated beyond root.
	if m.nav.Depth() > 1 {
		crumbs := lipgloss.NewStyle().Foreground(ui.ColorMuted).Render(m.nav.Breadcrumbs())
		parts = append(parts, crumbs)
	}

	// Active view content.
	switch m.nav.Current() {
	case ViewDashboard:
		parts = append(parts, m.dashboard.View())
	case ViewJobs:
		parts = append(parts, m.jobsView.View())
	case ViewJobDetail:
		parts = append(parts, m.jobDetailView.View())
	case ViewNodes:
		parts = append(parts, m.nodesView.View())
	case ViewNodeDetail:
		parts = append(parts, m.nodeDetailView.View())
	case ViewActors:
		parts = append(parts, m.actorsView.View())
	case ViewActorDetail:
		parts = append(parts, m.actorDetailView.View())
	case ViewServe:
		parts = append(parts, m.serveView.View())
	case ViewServeDetail:
		parts = append(parts, m.serveDetailView.View())
	case ViewEvents:
		parts = append(parts, m.eventsView.View())
	}

	// Help overlay -- context-sensitive: show view-specific keys alongside global keys.
	if m.showHelp {
		parts = append(parts, m.helpModel.View(m.activeHelpKeys()))
	}

	// Command bar.
	if m.command.Active() {
		parts = append(parts, m.command.View())
	}

	body := lipgloss.JoinVertical(lipgloss.Left, parts...)

	// Status bar at bottom.
	statusBar := m.renderStatusBar(m.width)

	bodyHeight := lipgloss.Height(body)
	statusBarHeight := lipgloss.Height(statusBar)
	spacerHeight := m.height - bodyHeight - statusBarHeight
	if spacerHeight < 0 {
		spacerHeight = 0
	}

	content := body + strings.Repeat("\n", spacerHeight) + statusBar

	v := tea.NewView(content)
	v.AltScreen = true
	return v
}

// activeHelpKeys returns the appropriate help.KeyMap for the current view.
// When on a view with its own keybindings, returns a CombinedKeyMap that shows
// both the view-specific and global keybindings.
func (m Model) activeHelpKeys() help.KeyMap {
	switch m.nav.Current() {
	case ViewJobs:
		return CombinedKeyMap{Global: GlobalKeys, View: jobs.Keys}
	case ViewJobDetail:
		return CombinedKeyMap{Global: GlobalKeys, View: jobdetail.Keys}
	case ViewNodes:
		return CombinedKeyMap{Global: GlobalKeys, View: nodes.Keys}
	case ViewNodeDetail:
		return CombinedKeyMap{Global: GlobalKeys, View: nodedetail.Keys}
	case ViewActors:
		return CombinedKeyMap{Global: GlobalKeys, View: actors.Keys}
	case ViewActorDetail:
		return CombinedKeyMap{Global: GlobalKeys, View: actordetail.Keys}
	case ViewServe:
		return CombinedKeyMap{Global: GlobalKeys, View: serve.Keys}
	case ViewServeDetail:
		return CombinedKeyMap{Global: GlobalKeys, View: servedetail.Keys}
	case ViewEvents:
		return CombinedKeyMap{Global: GlobalKeys, View: events.Keys}
	default:
		return GlobalKeys
	}
}

// renderStatusBar renders the global status bar with connection health and refresh indicator.
func (m Model) renderStatusBar(width int) string {
	if width == 0 {
		width = 80
	}

	// Left side: profile name (if any) + connection status + latency.
	var left string

	// Profile prefix for status bar.
	profilePrefix := ""
	if m.activeProfile != "" {
		profilePrefix = lipgloss.NewStyle().Bold(true).Foreground(ui.ColorAccent).Render(m.activeProfile) + " | "
	}

	if m.lastErr != nil {
		dot := lipgloss.NewStyle().Foreground(ui.ColorDanger).Render("●")
		errMsg := m.lastErr.Error()
		if len(errMsg) > 30 {
			errMsg = errMsg[:30] + "..."
		}
		left = fmt.Sprintf("%s%s Disconnected | Error: %s", profilePrefix, dot, errMsg)
	} else if !m.lastUpdate.IsZero() {
		dot := lipgloss.NewStyle().Foreground(ui.ColorSuccess).Render("●")
		left = fmt.Sprintf("%s%s Connected | %dms", profilePrefix, dot, m.lastLatency.Milliseconds())
	} else {
		dot := lipgloss.NewStyle().Foreground(ui.ColorMuted).Render("●")
		left = fmt.Sprintf("%s%s Connecting...", profilePrefix, dot)
	}

	// Right side: refresh indicator.
	var right string
	if m.refreshing {
		right = "Refreshing..."
	} else if !m.lastUpdate.IsZero() {
		ago := time.Since(m.lastUpdate).Truncate(time.Second)
		right = fmt.Sprintf("Updated %s ago", ago)
	} else {
		right = "Waiting..."
	}

	leftWidth := lipgloss.Width(left)
	rightWidth := lipgloss.Width(right)
	gap := width - leftWidth - rightWidth - 2
	if gap < 1 {
		gap = 1
	}

	bar := left + strings.Repeat(" ", gap) + right
	return ui.StatusBarStyle.Width(width).Render(bar)
}

// renderPowerBar renders a compact cluster health indicator for the banner.
// Shows a mini segmented bar: node health + running job count.
// Example: ▐████░░▌ 3N 2J
func (m Model) renderPowerBar() string {
	if len(m.cachedNodes) == 0 && len(m.cachedJobs) == 0 {
		return ""
	}

	var segments []string

	// Node segments: green for alive, red for dead.
	for _, n := range m.cachedNodes {
		if n.State == "ALIVE" {
			segments = append(segments, lipgloss.NewStyle().Foreground(ui.ColorSuccess).Render("█"))
		} else {
			segments = append(segments, lipgloss.NewStyle().Foreground(ui.ColorDanger).Render("█"))
		}
	}

	// Job activity segments.
	running := 0
	failed := 0
	for _, j := range m.cachedJobs {
		switch j.Status {
		case "RUNNING":
			running++
		case "FAILED":
			failed++
		}
	}

	bar := lipgloss.NewStyle().Foreground(ui.ColorMuted).Render("▐") +
		strings.Join(segments, "") +
		lipgloss.NewStyle().Foreground(ui.ColorMuted).Render("▌")

	// Compact counts.
	var counts []string
	if running > 0 {
		counts = append(counts, lipgloss.NewStyle().Foreground(ui.ColorSuccess).Render(fmt.Sprintf("%dJ", running)))
	}
	if failed > 0 {
		counts = append(counts, lipgloss.NewStyle().Foreground(ui.ColorDanger).Render(fmt.Sprintf("%dF", failed)))
	}

	if len(counts) > 0 {
		return bar + " " + strings.Join(counts, " ")
	}
	return bar
}
