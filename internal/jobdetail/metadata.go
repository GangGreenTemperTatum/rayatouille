package jobdetail

import (
	"fmt"
	"image/color"
	"time"

	"charm.land/lipgloss/v2"

	"github.com/GangGreenTemperTatum/rayatouille/internal/ray"
	"github.com/GangGreenTemperTatum/rayatouille/internal/ui"
)

// renderMetadata renders the job metadata panel as key-value pairs.
func renderMetadata(job ray.JobDetail, width int) string {
	labelStyle := ui.LabelStyle.Width(16)
	valueStyle := ui.ValueStyle

	rows := []string{
		renderRow(labelStyle, valueStyle, "Status", statusStyled(job.Status)),
		renderRow(labelStyle, valueStyle, "Submission ID", job.SubmissionID),
		renderRow(labelStyle, valueStyle, "Job ID", ptrOrNA(job.JobID)),
		renderRow(labelStyle, valueStyle, "Entrypoint", job.Entrypoint),
		renderRow(labelStyle, valueStyle, "Started", formatTimestamp(job.StartTime)),
		renderRow(labelStyle, valueStyle, "Duration", formatDuration(job.StartTime, job.EndTime)),
		renderRow(labelStyle, valueStyle, "Driver Node", truncateNode(job.DriverNodeID)),
		renderRow(labelStyle, valueStyle, "Exit Code", formatExitCode(job.DriverExitCode)),
	}

	if job.Message != "" {
		rows = append(rows, renderRow(labelStyle, valueStyle, "Message", job.Message))
	}

	content := lipgloss.JoinVertical(lipgloss.Left, rows...)

	sectionWidth := width - 6 // account for section padding/border
	if sectionWidth < 40 {
		sectionWidth = 40
	}

	return ui.SectionStyle.Width(sectionWidth).Render(content)
}

// renderRow renders a single label: value row.
func renderRow(labelStyle, valueStyle lipgloss.Style, label, value string) string {
	return labelStyle.Render(label+":") + " " + valueStyle.Render(value)
}

// statusStyled returns the status string with color coding.
func statusStyled(status string) string {
	c := statusColor(status)
	return lipgloss.NewStyle().Foreground(c).Bold(true).Render(status)
}

// statusColor maps a job status to a theme color.
func statusColor(status string) color.Color {
	switch status {
	case "SUCCEEDED":
		return ui.ColorSuccess
	case "FAILED":
		return ui.ColorDanger
	case "RUNNING":
		return ui.ColorWarning
	case "PENDING":
		return ui.ColorMuted
	default:
		return ui.ColorFg
	}
}

// ptrOrNA returns the dereferenced string pointer or "N/A" if nil.
func ptrOrNA(s *string) string {
	if s == nil {
		return "N/A"
	}
	return *s
}

// formatTimestamp formats a Unix millisecond timestamp as a human-readable datetime.
func formatTimestamp(ms int64) string {
	if ms <= 0 {
		return "-"
	}
	t := time.UnixMilli(ms)
	return t.Format("2006-01-02 15:04:05")
}

// formatDuration returns a human-readable duration between start and end timestamps.
func formatDuration(startMs, endMs int64) string {
	if startMs <= 0 {
		return "-"
	}
	start := time.UnixMilli(startMs)
	var d time.Duration
	if endMs > 0 {
		end := time.UnixMilli(endMs)
		d = end.Sub(start)
	} else {
		d = time.Since(start)
	}
	if d < 0 {
		d = 0
	}
	switch {
	case d < time.Minute:
		return fmt.Sprintf("%ds", int(d.Seconds()))
	case d < time.Hour:
		return fmt.Sprintf("%dm%ds", int(d.Minutes()), int(d.Seconds())%60)
	default:
		return fmt.Sprintf("%dh%dm", int(d.Hours()), int(d.Minutes())%60)
	}
}

// truncateNode truncates a node ID to 12 chars + "..." if longer.
func truncateNode(nodeID string) string {
	if nodeID == "" {
		return "N/A"
	}
	if len(nodeID) > 12 {
		return nodeID[:12] + "..."
	}
	return nodeID
}

// formatExitCode formats an optional exit code.
func formatExitCode(code *int) string {
	if code == nil {
		return "N/A"
	}
	return fmt.Sprintf("%d", *code)
}
