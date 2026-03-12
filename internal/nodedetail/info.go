package nodedetail

import (
	"fmt"
	image_color "image/color"
	"time"

	"charm.land/lipgloss/v2"

	"github.com/GangGreenTemperTatum/rayatouille/internal/ray"
	"github.com/GangGreenTemperTatum/rayatouille/internal/ui"
)

// renderInfo renders the node info panel showing metadata and resources.
func renderInfo(node ray.Node, width int) string {
	labelStyle := ui.LabelStyle.Width(16)
	valueStyle := ui.ValueStyle

	// Role string.
	role := "worker"
	if node.IsHeadNode {
		role = "head"
	}

	// State with color.
	stateStr := stateStyled(node.State)

	rows := []string{
		renderRow(labelStyle, valueStyle, "Node ID", node.NodeID),
		renderRow(labelStyle, valueStyle, "IP", node.NodeIP),
		renderRow(labelStyle, valueStyle, "State", stateStr),
		renderRow(labelStyle, valueStyle, "Role", role),
		renderRow(labelStyle, valueStyle, "Start Time", formatTimestamp(node.StartTimeMs)),
	}

	// State message for dead nodes.
	if node.StateMessage != nil && *node.StateMessage != "" {
		rows = append(rows, renderRow(labelStyle, valueStyle, "State Message", *node.StateMessage))
	}

	content := lipgloss.JoinVertical(lipgloss.Left, rows...)
	sectionWidth := width - 6
	if sectionWidth < 40 {
		sectionWidth = 40
	}
	metaSection := ui.SectionStyle.Width(sectionWidth).Render(content)

	// Resources section.
	resourceRows := []string{
		renderResourceRow(labelStyle, "CPU", node.ResourcesTotal, node.ResourcesAvailable, "CPU", false),
		renderResourceRow(labelStyle, "GPU", node.ResourcesTotal, node.ResourcesAvailable, "GPU", true),
		renderResourceRow(labelStyle, "Memory", node.ResourcesTotal, node.ResourcesAvailable, "memory", false),
		renderResourceRow(labelStyle, "Object Store", node.ResourcesTotal, node.ResourcesAvailable, "object_store_memory", false),
	}
	resourceContent := lipgloss.JoinVertical(lipgloss.Left, resourceRows...)
	resourceSection := ui.SectionStyle.Width(sectionWidth).Render(resourceContent)

	parts := []string{metaSection, "", resourceSection}

	// Labels section.
	if len(node.Labels) > 0 {
		var labelRows []string
		for k, v := range node.Labels {
			labelRows = append(labelRows, renderRow(labelStyle, valueStyle, k, v))
		}
		labelContent := lipgloss.JoinVertical(lipgloss.Left, labelRows...)
		labelSection := ui.SectionStyle.Width(sectionWidth).Render(labelContent)
		parts = append(parts, "", labelSection)
	}

	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}

// renderRow renders a single label: value row.
func renderRow(labelStyle, valueStyle lipgloss.Style, label, value string) string {
	return labelStyle.Render(label+":") + " " + valueStyle.Render(value)
}

// renderResourceRow renders a resource row with total and used/available info.
func renderResourceRow(labelStyle lipgloss.Style, label string, total, available map[string]float64, key string, isGPU bool) string {
	totalVal, hasTotal := total[key]

	if !hasTotal || (isGPU && totalVal == 0) {
		if isGPU {
			return labelStyle.Render(label+":") + " " + ui.ValueStyle.Render("-")
		}
		return labelStyle.Render(label+":") + " " + ui.ValueStyle.Render("N/A")
	}

	totalStr := formatResourceValue(key, totalVal)

	if available == nil {
		return labelStyle.Render(label+":") + " " + ui.ValueStyle.Render(totalStr+" (available: N/A)")
	}

	availVal, hasAvail := available[key]
	if !hasAvail {
		return labelStyle.Render(label+":") + " " + ui.ValueStyle.Render(totalStr+" (available: N/A)")
	}

	usedVal := totalVal - availVal
	usedStr := formatResourceValue(key, usedVal)
	return labelStyle.Render(label+":") + " " + ui.ValueStyle.Render(fmt.Sprintf("%s / %s", usedStr, totalStr))
}

// formatResourceValue formats a resource value based on its key.
func formatResourceValue(key string, val float64) string {
	switch key {
	case "memory", "object_store_memory":
		gib := val / (1024 * 1024 * 1024)
		return fmt.Sprintf("%.1f GiB", gib)
	case "CPU", "GPU":
		return fmt.Sprintf("%d", int(val))
	default:
		return fmt.Sprintf("%.0f", val)
	}
}

// stateStyled returns the state string with color coding.
func stateStyled(state string) string {
	var c image_color.Color
	switch state {
	case "ALIVE":
		c = ui.ColorSuccess
	case "DEAD":
		c = ui.ColorDanger
	default:
		c = ui.ColorFg
	}
	return lipgloss.NewStyle().Foreground(c).Bold(true).Render(state)
}

// formatTimestamp formats a Unix millisecond timestamp as a human-readable datetime.
func formatTimestamp(ms int64) string {
	if ms <= 0 {
		return "-"
	}
	t := time.UnixMilli(ms)
	return t.Format("2006-01-02 15:04:05")
}
