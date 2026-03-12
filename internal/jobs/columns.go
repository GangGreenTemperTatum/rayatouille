package jobs

import (
	"fmt"
	"time"

	"charm.land/bubbles/v2/table"

	"github.com/GangGreenTemperTatum/rayatouille/internal/ray"
)

// Columns returns the table column definitions for the given width.
func Columns(width int) []table.Column {
	const (
		idWidth       = 14
		statusWidth   = 10
		durationWidth = 12
		ageWidth      = 12
		padding       = 6 // cell padding
	)

	entrypointWidth := width - idWidth - statusWidth - durationWidth - ageWidth - padding
	if entrypointWidth < 20 {
		entrypointWidth = 20
	}

	return []table.Column{
		{Title: "Submission ID", Width: idWidth},
		{Title: "Status", Width: statusWidth},
		{Title: "Entrypoint", Width: entrypointWidth},
		{Title: "Duration", Width: durationWidth},
		{Title: "Age", Width: ageWidth},
	}
}

// JobToRow converts a JobDetail into a table row.
func JobToRow(j ray.JobDetail) table.Row {
	subID := j.SubmissionID
	if len(subID) > 12 {
		subID = subID[:12] + ".."
	}

	entrypoint := j.Entrypoint
	const maxEntrypoint = 30
	if len(entrypoint) > maxEntrypoint {
		entrypoint = entrypoint[:maxEntrypoint-3] + "..."
	}

	duration := formatDuration(j.StartTime, j.EndTime)
	age := formatRelativeTime(j.StartTime)

	return table.Row{subID, j.Status, entrypoint, duration, age}
}

// formatDuration returns a human-readable duration string.
func formatDuration(startMs, endMs int64) string {
	if startMs <= 0 {
		return "-"
	}
	start := time.UnixMilli(startMs)
	if endMs > 0 {
		end := time.UnixMilli(endMs)
		d := end.Sub(start)
		return formatDurationValue(d)
	}
	// Still running.
	d := time.Since(start)
	return formatDurationValue(d)
}

// formatDurationValue formats a duration to a compact string.
func formatDurationValue(d time.Duration) string {
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

// formatRelativeTime formats a Unix millisecond timestamp as a relative time string.
func formatRelativeTime(ms int64) string {
	if ms == 0 {
		return "-"
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
