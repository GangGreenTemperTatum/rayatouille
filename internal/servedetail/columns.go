package servedetail

import (
	"fmt"
	"time"

	"charm.land/bubbles/v2/table"

	"github.com/GangGreenTemperTatum/rayatouille/internal/ray"
)

// DeploymentColumns returns the table column definitions for the deployments table.
func DeploymentColumns(width int) []table.Column {
	const (
		nameWidth    = 20
		statusWidth  = 12
		targetWidth  = 8
		runningWidth = 8
		padding      = 6
	)

	msgWidth := width - nameWidth - statusWidth - targetWidth - runningWidth - padding
	if msgWidth < 10 {
		msgWidth = 10
	}

	return []table.Column{
		{Title: "Name", Width: nameWidth},
		{Title: "Status", Width: statusWidth},
		{Title: "Target", Width: targetWidth},
		{Title: "Running", Width: runningWidth},
		{Title: "Message", Width: msgWidth},
	}
}

// DeploymentToRow converts a deployment name and details into a table row.
func DeploymentToRow(name string, d ray.DeploymentDetails) table.Row {
	displayName := name
	if len(displayName) > 20 {
		displayName = displayName[:18] + ".."
	}

	running := 0
	for _, r := range d.Replicas {
		if r.State == "RUNNING" {
			running++
		}
	}

	msg := d.Message
	if msg == "" {
		msg = "-"
	} else if len(msg) > 50 {
		msg = msg[:48] + ".."
	}

	return table.Row{
		displayName,
		d.Status,
		fmt.Sprintf("%d", d.TargetNumReplicas),
		fmt.Sprintf("%d", running),
		msg,
	}
}

// ReplicaColumns returns the table column definitions for the replicas table.
func ReplicaColumns(width int) []table.Column {
	const (
		replicaIDWidth = 14
		stateWidth     = 10
		pidWidth       = 8
		nodeIPWidth    = 16
		startTimeWidth = 20
		padding        = 6
	)

	return []table.Column{
		{Title: "Replica ID", Width: replicaIDWidth},
		{Title: "State", Width: stateWidth},
		{Title: "PID", Width: pidWidth},
		{Title: "Node IP", Width: nodeIPWidth},
		{Title: "Start Time", Width: startTimeWidth},
	}
}

// ReplicaToRow converts a ReplicaDetails into a table row.
func ReplicaToRow(r ray.ReplicaDetails) table.Row {
	replicaID := r.ReplicaID
	if len(replicaID) > 12 {
		replicaID = replicaID[:12] + ".."
	}

	pid := "-"
	if r.PID != nil {
		pid = fmt.Sprintf("%d", *r.PID)
	}

	nodeIP := "-"
	if r.NodeIP != nil {
		nodeIP = *r.NodeIP
	}

	startTime := "-"
	if r.StartTimeS > 0 {
		startTime = formatRelativeTime(r.StartTimeS)
	}

	return table.Row{replicaID, r.State, pid, nodeIP, startTime}
}

// formatRelativeTime formats a Unix timestamp (float64 seconds) as a relative time string.
func formatRelativeTime(ts float64) string {
	t := time.Unix(int64(ts), 0)
	d := time.Since(t)
	if d < 0 {
		return "just now"
	}
	switch {
	case d < time.Minute:
		return fmt.Sprintf("%ds ago", int(d.Seconds()))
	case d < time.Hour:
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	default:
		days := int(d.Hours() / 24)
		return fmt.Sprintf("%dd ago", days)
	}
}
