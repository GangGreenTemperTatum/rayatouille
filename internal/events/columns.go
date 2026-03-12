package events

import (
	"time"

	"charm.land/bubbles/v2/table"

	"github.com/GangGreenTemperTatum/rayatouille/internal/ray"
)

// Columns returns the table column definitions for the events table.
func Columns(width int) []table.Column {
	const (
		timeWidth     = 20
		severityWidth = 8
		sourceWidth   = 12
		padding       = 6
	)

	msgWidth := width - timeWidth - severityWidth - sourceWidth - padding
	if msgWidth < 10 {
		msgWidth = 10
	}

	return []table.Column{
		{Title: "Time", Width: timeWidth},
		{Title: "Severity", Width: severityWidth},
		{Title: "Source", Width: sourceWidth},
		{Title: "Message", Width: msgWidth},
	}
}

// EventToRow converts a ClusterEvent into a table row.
func EventToRow(e ray.ClusterEvent) table.Row {
	displayTime := e.Time
	// Try parsing as ISO timestamp for consistent formatting.
	if t, err := time.Parse(time.RFC3339, e.Time); err == nil {
		displayTime = t.Format("2006-01-02 15:04:05")
	} else if t, err := time.Parse(time.RFC3339Nano, e.Time); err == nil {
		displayTime = t.Format("2006-01-02 15:04:05")
	}

	msg := e.Message
	if len(msg) > 80 {
		msg = msg[:78] + ".."
	}

	return table.Row{displayTime, e.Severity, e.SourceType, msg}
}
