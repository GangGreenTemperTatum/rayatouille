package dashboard

import (
	"fmt"
	"image/color"
	"sort"
	"strings"
	"time"

	"charm.land/bubbles/v2/progress"
	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/GangGreenTemperTatum/rayatouille/internal/ray"
	"github.com/GangGreenTemperTatum/rayatouille/internal/ui"
)

// ClusterData holds fetched cluster state from a polling cycle.
type ClusterData struct {
	Nodes    []ray.Node
	Jobs     []ray.JobDetail
	FetchErr error
	Latency  time.Duration
}

// ClusterDataMsg carries fetched cluster data as a Bubble Tea message.
type ClusterDataMsg ClusterData

// maxRecentJobs is the number of recent jobs to display.
const maxRecentJobs = 5

// Model is the dashboard view model displaying cluster overview data.
type Model struct {
	width       int
	height      int
	health      ClusterHealth
	jobs        JobSummary
	nodes       []ray.Node
	jobDetails  []ray.JobDetail
	lastErr     error
	lastLatency time.Duration
	refreshing  bool
	lastUpdate  time.Time
	spinner     spinner.Model
	address     string

	// Progress bars for resource capacity display.
	cpuBar  progress.Model
	gpuBar  progress.Model
	memBar  progress.Model
	diskBar progress.Model

	// Heatmap state.
	heatmapResource HeatmapResource
}

// New creates a new dashboard Model with the cluster address for display.
func New(address string) Model {
	barOpts := []progress.Option{
		progress.WithWidth(30),
		progress.WithDefaultBlend(),
		progress.WithoutPercentage(),
	}
	return Model{
		spinner: spinner.New(spinner.WithSpinner(spinner.Dot)),
		cpuBar:  progress.New(barOpts...),
		gpuBar:  progress.New(barOpts...),
		memBar:  progress.New(barOpts...),
		diskBar: progress.New(barOpts...),
		address: address,
	}
}

// SetSize updates the dashboard dimensions.
func (m *Model) SetSize(w, h int) {
	m.width = w
	m.height = h
}

// Init returns the initial command for the dashboard (spinner tick).
func (m Model) Init() tea.Cmd {
	return m.spinner.Tick
}

// Update handles messages for the dashboard model.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case ClusterDataMsg:
		m.refreshing = false
		m.lastLatency = msg.Latency
		m.lastErr = msg.FetchErr
		if msg.FetchErr == nil {
			m.nodes = msg.Nodes
			m.jobDetails = msg.Jobs
			m.health = AggregateClusterHealth(m.nodes)
			m.jobs = AggregateJobSummary(m.jobDetails)
			m.lastUpdate = time.Now()
		}
		return m, nil

	case tea.KeyPressMsg:
		if msg.Text == "h" {
			m.heatmapResource = HeatmapResource((int(m.heatmapResource) + 1) % heatmapResourceCount)
			return m, nil
		}

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}
	return m, nil
}

// View renders the dashboard.
func (m Model) View() string {
	w := m.width
	if w == 0 {
		w = 80
	}

	header := m.renderHeader()
	errBanner := m.renderErrorBanner()
	donut := renderDonutChart(m.jobs, m.health)
	// Shrink health section to leave room for donut chart when present.
	donutWidth := 20 // donut grid (11) + margin (2) + legend padding
	healthWidth := w
	if donut != "" {
		healthWidth = w - donutWidth
	}
	health := m.renderHealthSection(healthWidth)
	resources := m.renderResourceBars(w)
	nodeDots := m.renderNodeDots(w)
	nodeStatus := m.renderNodeStatus(w)
	jobStatusBar := m.renderJobStatusBar(w)
	recentJobs := m.renderRecentJobs(w)
	heatmap := m.renderNodeHeatmap(w)

	nav := lipgloss.NewStyle().Foreground(ui.ColorMuted).Render(
		"[1] jobs  [2] nodes  [3] actors  [4] serve  [5] events  [tab] cycle  [:] command  [?] help")

	parts := []string{header}
	if errBanner != "" {
		parts = append(parts, errBanner)
	}

	// Place donut chart to the right of the health section.
	if donut != "" && w > 50 {
		donutStyled := lipgloss.NewStyle().MarginLeft(2).Render(donut)
		healthRow := lipgloss.JoinHorizontal(lipgloss.Top, health, donutStyled)
		parts = append(parts, healthRow)
	} else {
		parts = append(parts, health)
	}

	if nodeDots != "" {
		parts = append(parts, nodeDots)
	}
	parts = append(parts, resources)
	if nodeStatus != "" {
		parts = append(parts, nodeStatus)
	}
	if jobStatusBar != "" {
		parts = append(parts, jobStatusBar)
	}
	parts = append(parts, recentJobs)
	if heatmap != "" {
		parts = append(parts, heatmap)
	}
	parts = append(parts, "", nav)

	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}

// renderNodeHeatmap renders the node heatmap section if nodes are available.
func (m Model) renderNodeHeatmap(width int) string {
	if len(m.nodes) == 0 {
		return ""
	}
	return renderHeatmap(m.nodes, m.heatmapResource, width)
}

// renderHeader renders the dashboard title and cluster status indicator.
func (m Model) renderHeader() string {
	title := ui.TitleStyle.Render("Cluster Overview")

	// Show cluster address as a clickable OSC 8 hyperlink.
	var addressLine string
	if m.address != "" {
		// OSC 8 terminal hyperlink: \033]8;;URL\033\\LABEL\033]8;;\033\\
		addressLine = fmt.Sprintf("\033]8;;%s\033\\%s\033]8;;\033\\",
			m.address,
			lipgloss.NewStyle().Foreground(ui.ColorAccent).Underline(true).Render(m.address))
	}

	var indicator string
	switch m.health.Status {
	case "healthy":
		indicator = lipgloss.NewStyle().Foreground(ui.ColorSuccess).Render("●") + " healthy"
	case "degraded":
		indicator = lipgloss.NewStyle().Foreground(ui.ColorWarning).Render("●") + " degraded"
	case "unhealthy":
		indicator = lipgloss.NewStyle().Foreground(ui.ColorDanger).Render("●") + " unhealthy"
	default:
		indicator = lipgloss.NewStyle().Foreground(ui.ColorMuted).Render("●") + " unknown"
	}

	header := title + "\n" + indicator
	if addressLine != "" {
		header += "  " + addressLine
	}
	return header
}

// renderHealthSection renders the cluster health summary with node counts and resource totals.
func (m Model) renderHealthSection(width int) string {
	// If no data yet, show waiting message.
	if m.health.NodeCount == 0 && m.health.TotalCPU == 0 && m.lastErr == nil {
		content := "Waiting for data..."
		return ui.SectionStyle.Width(width - 4).Render("Cluster Health\n\n" + content)
	}

	// Node count with color coding.
	nodeColor := ui.ColorSuccess
	if m.health.AliveNodes < m.health.NodeCount {
		nodeColor = ui.ColorWarning
	}
	nodesStr := lipgloss.NewStyle().Foreground(nodeColor).Render(
		fmt.Sprintf("%d/%d", m.health.AliveNodes, m.health.NodeCount),
	)

	memGiB := m.health.TotalMemory / 1073741824.0

	lines := []string{
		fmt.Sprintf("Nodes: %s", nodesStr),
		fmt.Sprintf("CPU: %.0f cores | GPU: %.0f | Memory: %.1f GiB",
			m.health.TotalCPU, m.health.TotalGPU, memGiB),
	}

	content := strings.Join(lines, "\n")
	return ui.SectionStyle.Width(width - 4).Render("Cluster Health\n\n" + content)
}

// resourceRatio returns used/total as a float64 in [0.0, 1.0].
// Returns 0 if total is zero.
func resourceRatio(used, total float64) float64 {
	if total <= 0 {
		return 0
	}
	r := used / total
	if r < 0 {
		return 0
	}
	if r > 1 {
		return 1
	}
	return r
}

// renderResourceBars renders resource utilization bars using progress.ViewAs().
// When resources_available data is present, bars show used/total ratio.
// When unavailable, bars show 0% with "N/A" label.
func (m Model) renderResourceBars(width int) string {
	barWidth := width - 20
	if barWidth > 30 {
		barWidth = 30
	}
	if barWidth < 5 {
		barWidth = 5
	}

	m.cpuBar.SetWidth(barWidth)
	m.gpuBar.SetWidth(barWidth)
	m.memBar.SetWidth(barWidth)
	m.diskBar.SetWidth(barWidth)

	hasData := m.health.HasAvailableData

	var cpuLine string
	if hasData {
		ratio := resourceRatio(m.health.UsedCPU, m.health.TotalCPU)
		cpuLine = fmt.Sprintf("CPU  %s %.0f/%.0f cores", m.cpuBar.ViewAs(ratio), m.health.UsedCPU, m.health.TotalCPU)
	} else {
		cpuLine = fmt.Sprintf("CPU  %s %.0f cores", m.cpuBar.ViewAs(0), m.health.TotalCPU)
	}

	var gpuLine string
	if m.health.TotalGPU == 0 {
		gpuLine = "GPU  N/A"
	} else if hasData {
		ratio := resourceRatio(m.health.UsedGPU, m.health.TotalGPU)
		gpuLine = fmt.Sprintf("GPU  %s %.0f/%.0f units", m.gpuBar.ViewAs(ratio), m.health.UsedGPU, m.health.TotalGPU)
	} else {
		gpuLine = fmt.Sprintf("GPU  %s %.0f units", m.gpuBar.ViewAs(0), m.health.TotalGPU)
	}

	memGiB := m.health.TotalMemory / 1073741824.0
	usedMemGiB := m.health.UsedMemory / 1073741824.0
	var memLine string
	if hasData {
		ratio := resourceRatio(m.health.UsedMemory, m.health.TotalMemory)
		memLine = fmt.Sprintf("MEM  %s %.1f/%.1f GiB", m.memBar.ViewAs(ratio), usedMemGiB, memGiB)
	} else {
		memLine = fmt.Sprintf("MEM  %s %.1f GiB", m.memBar.ViewAs(0), memGiB)
	}

	var diskLine string
	objStoreGiB := m.health.TotalObjectStoreMemory / 1073741824.0
	usedObjStoreGiB := m.health.UsedObjectStoreMemory / 1073741824.0
	if m.health.TotalObjectStoreMemory == 0 {
		diskLine = "DISK N/A"
	} else if hasData {
		ratio := resourceRatio(m.health.UsedObjectStoreMemory, m.health.TotalObjectStoreMemory)
		diskLine = fmt.Sprintf("DISK %s %.1f/%.1f GiB", m.diskBar.ViewAs(ratio), usedObjStoreGiB, objStoreGiB)
	} else {
		diskLine = fmt.Sprintf("DISK %s %.1f GiB", m.diskBar.ViewAs(0), objStoreGiB)
	}

	lines := strings.Join([]string{cpuLine, gpuLine, memLine, diskLine}, "\n")
	return ui.SectionStyle.Width(width - 4).Render("Resource Utilization\n\n" + lines)
}

// renderJobsSummary renders job status counts.
func (m Model) renderJobsSummary(width int) string {
	if m.jobs.Total == 0 {
		return ui.SectionStyle.Width(width - 4).Render("Jobs\n\nNo jobs found")
	}

	running := lipgloss.NewStyle().Foreground(ui.ColorSuccess).Render(fmt.Sprintf("%d running", m.jobs.Running))
	pending := lipgloss.NewStyle().Foreground(ui.ColorWarning).Render(fmt.Sprintf("%d pending", m.jobs.Pending))
	failed := lipgloss.NewStyle().Foreground(ui.ColorDanger).Render(fmt.Sprintf("%d failed", m.jobs.Failed))
	succeeded := lipgloss.NewStyle().Foreground(ui.ColorMuted).Render(fmt.Sprintf("%d succeeded", m.jobs.Succeeded))

	counts := fmt.Sprintf("%s | %s | %s | %s", running, pending, failed, succeeded)
	total := fmt.Sprintf("%d total jobs", m.jobs.Total)

	content := counts + "\n" + total
	return ui.SectionStyle.Width(width - 4).Render("Jobs\n\n" + content)
}

// renderRecentJobs renders the last N jobs sorted by most recent activity.
func (m Model) renderRecentJobs(width int) string {
	if len(m.jobDetails) == 0 {
		return ui.SectionStyle.Width(width - 4).Render("Recent Jobs\n\nNo recent jobs")
	}

	// Copy and sort by most recent activity (descending).
	sorted := make([]ray.JobDetail, len(m.jobDetails))
	copy(sorted, m.jobDetails)
	sort.Slice(sorted, func(i, j int) bool {
		return jobTimestamp(sorted[i]) > jobTimestamp(sorted[j])
	})

	// Limit to maxRecentJobs.
	if len(sorted) > maxRecentJobs {
		sorted = sorted[:maxRecentJobs]
	}

	var lines []string
	for _, j := range sorted {
		statusStyled := colorizeStatus(j.Status)

		// Show entrypoint (truncated) instead of raw submission ID.
		entry := j.Entrypoint
		if entry == "" {
			entry = j.SubmissionID
		}
		maxEntry := 40
		if width > 0 {
			maxEntry = width/2 - 10
		}
		if maxEntry < 20 {
			maxEntry = 20
		}
		if len(entry) > maxEntry {
			entry = entry[:maxEntry-1] + "…"
		}
		entryStyled := lipgloss.NewStyle().Foreground(ui.ColorFg).Render(entry)

		// Duration or age.
		dur := formatJobDuration(j)

		lines = append(lines, fmt.Sprintf("%s  %s  %s", statusStyled, entryStyled, lipgloss.NewStyle().Foreground(ui.ColorMuted).Render(dur)))
	}

	hint := lipgloss.NewStyle().Foreground(ui.ColorMuted).Render("  press [1] to browse jobs")
	content := strings.Join(lines, "\n")
	return ui.SectionStyle.Width(width - 4).Render("Recent Jobs\n\n" + content + "\n" + hint)
}

// formatJobDuration returns a human-readable duration or age for a job.
func formatJobDuration(j ray.JobDetail) string {
	if j.Status == "RUNNING" && j.StartTime > 0 {
		d := time.Since(time.UnixMilli(j.StartTime))
		return "running " + formatDuration(d)
	}
	if j.EndTime > 0 && j.StartTime > 0 {
		d := time.Duration(j.EndTime-j.StartTime) * time.Millisecond
		return "took " + formatDuration(d)
	}
	return formatRelativeTime(jobTimestamp(j))
}

// formatDuration formats a duration as a concise human-readable string.
func formatDuration(d time.Duration) string {
	switch {
	case d < time.Minute:
		return fmt.Sprintf("%ds", int(d.Seconds()))
	case d < time.Hour:
		return fmt.Sprintf("%dm%ds", int(d.Minutes()), int(d.Seconds())%60)
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh%dm", int(d.Hours()), int(d.Minutes())%60)
	default:
		return fmt.Sprintf("%dd%dh", int(d.Hours())/24, int(d.Hours())%24)
	}
}

// renderNodeStatus renders individual node rows with status indicators.
func (m Model) renderNodeStatus(width int) string {
	if len(m.nodes) == 0 {
		return ""
	}

	var lines []string
	for _, n := range m.nodes {
		var dot string
		switch n.State {
		case "ALIVE":
			dot = lipgloss.NewStyle().Foreground(ui.ColorSuccess).Render("●")
		case "DEAD":
			dot = lipgloss.NewStyle().Foreground(ui.ColorDanger).Render("●")
		default:
			dot = lipgloss.NewStyle().Foreground(ui.ColorWarning).Render("●")
		}

		role := ""
		if n.IsHeadNode {
			role = lipgloss.NewStyle().Foreground(ui.ColorAccent).Render(" (head)")
		}

		name := n.NodeName
		if name == "" {
			name = n.NodeIP
		}

		cpu := n.ResourcesTotal["CPU"]
		memGiB := n.ResourcesTotal["memory"] / 1073741824.0

		info := lipgloss.NewStyle().Foreground(ui.ColorMuted).Render(
			fmt.Sprintf("  cpu:%.0f mem:%.0fG", cpu, memGiB))

		lines = append(lines, fmt.Sprintf("%s %s%s%s", dot, name, role, info))
	}

	hint := lipgloss.NewStyle().Foreground(ui.ColorMuted).Render("  press [2] to browse nodes")
	content := strings.Join(lines, "\n")
	return ui.SectionStyle.Width(width - 4).Render("Nodes\n\n" + content + "\n" + hint)
}

// renderErrorBanner renders an error banner if there is an error.
func (m Model) renderErrorBanner() string {
	if m.lastErr == nil {
		return ""
	}
	errStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ui.ColorDanger).
		Foreground(ui.ColorDanger).
		Padding(0, 1)
	return errStyle.Render(m.lastErr.Error())
}

// jobTimestamp returns the most relevant timestamp for a job (EndTime if > 0, else StartTime).
func jobTimestamp(j ray.JobDetail) int64 {
	if j.EndTime > 0 {
		return j.EndTime
	}
	return j.StartTime
}

// formatRelativeTime formats a Unix millisecond timestamp as a relative time string.
func formatRelativeTime(ms int64) string {
	if ms == 0 {
		return "unknown"
	}
	t := time.UnixMilli(ms)
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return fmt.Sprintf("%ds ago", int(d.Seconds()))
	case d < time.Hour:
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	default:
		return fmt.Sprintf("%dd ago", int(d.Hours()/24))
	}
}

// colorizeStatus returns a styled status string with appropriate color.
func colorizeStatus(status string) string {
	var c color.Color
	switch status {
	case "RUNNING":
		c = ui.ColorSuccess
	case "PENDING":
		c = ui.ColorWarning
	case "FAILED":
		c = ui.ColorDanger
	case "SUCCEEDED", "STOPPED":
		c = ui.ColorMuted
	default:
		c = ui.ColorFg
	}
	return lipgloss.NewStyle().Foreground(c).Render(status)
}
