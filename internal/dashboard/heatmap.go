package dashboard

import (
	"fmt"
	"image/color"
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/GangGreenTemperTatum/rayatouille/internal/ray"
	"github.com/GangGreenTemperTatum/rayatouille/internal/ui"
)

// HeatmapResource represents the resource type displayed in the heatmap.
type HeatmapResource int

const (
	// HeatCPU displays CPU utilization in the heatmap.
	HeatCPU HeatmapResource = iota
	// HeatMemory displays memory utilization in the heatmap.
	HeatMemory
	// HeatGPU displays GPU utilization in the heatmap.
	HeatGPU
)

// heatmapResourceCount is the total number of heatmap resource types for cycling.
const heatmapResourceCount = 3

// heatmapResourceLabel returns a human-readable label for the resource.
func heatmapResourceLabel(r HeatmapResource) string {
	switch r {
	case HeatCPU:
		return "CPU"
	case HeatMemory:
		return "Memory"
	case HeatGPU:
		return "GPU"
	default:
		return "CPU"
	}
}

// heatColor returns a discrete color for the given utilization ratio.
// Negative ratios indicate unknown/N/A data and return gray.
func heatColor(ratio float64) color.Color {
	switch {
	case ratio < 0:
		return lipgloss.Color("#626262") // gray -- unknown/N/A
	case ratio < 0.3:
		return lipgloss.Color("#04B575") // green -- low usage
	case ratio < 0.6:
		return lipgloss.Color("#6FBF40") // light green -- moderate
	case ratio < 0.8:
		return lipgloss.Color("#FFCC00") // yellow -- elevated
	default:
		return lipgloss.Color("#FF4444") // red -- high usage
	}
}

// nodeResourceRatio returns the utilization ratio (0.0-1.0) for a node's resource.
// Returns -1 if the resource data is unavailable (nil resources_available, zero total, or missing key).
func nodeResourceRatio(n ray.Node, resource HeatmapResource) float64 {
	var key string
	switch resource {
	case HeatCPU:
		key = "CPU"
	case HeatMemory:
		key = "memory"
	case HeatGPU:
		key = "GPU"
	default:
		key = "CPU"
	}

	total, ok := n.ResourcesTotal[key]
	if !ok || total == 0 {
		return -1
	}

	if n.ResourcesAvailable == nil {
		return -1
	}

	avail, ok := n.ResourcesAvailable[key]
	if !ok {
		return -1
	}

	ratio := (total - avail) / total
	if ratio < 0 {
		return 0
	}
	if ratio > 1 {
		return 1
	}
	return ratio
}

// renderHeatmapCell renders a single heatmap cell for a node.
func renderHeatmapCell(n ray.Node, resource HeatmapResource) string {
	// Build label: last segment of IP or full IP if short.
	label := n.NodeIP
	if idx := strings.LastIndex(label, "."); idx >= 0 && idx < len(label)-1 {
		label = label[idx+1:]
	}
	if len(label) > 12 {
		label = label[:12]
	}

	// Head node indicator.
	if n.IsHeadNode {
		label += "*"
	}

	// Dead node styling.
	if n.State != "ALIVE" {
		label += " X"
		style := lipgloss.NewStyle().
			Background(lipgloss.Color("#333333")).
			Foreground(lipgloss.Color("#888888")).
			Padding(0, 1)
		return style.Render(label)
	}

	// Alive node: color by resource ratio.
	ratio := nodeResourceRatio(n, resource)
	bg := heatColor(ratio)
	style := lipgloss.NewStyle().
		Background(bg).
		Foreground(lipgloss.Color("#000000")).
		Padding(0, 1)
	return style.Render(label)
}

// renderHeatmap renders the full heatmap grid for the given nodes.
func renderHeatmap(nodes []ray.Node, resource HeatmapResource, width int) string {
	if len(nodes) == 0 {
		return "No nodes"
	}

	// Build cells.
	cells := make([]string, len(nodes))
	for i, n := range nodes {
		cells[i] = renderHeatmapCell(n, resource)
	}

	// Compute layout.
	cellWidth := lipgloss.Width(cells[0])
	if cellWidth == 0 {
		cellWidth = 6
	}
	cols := (width - 4) / cellWidth
	if cols < 1 {
		cols = 1
	}

	// Arrange into rows.
	var rows []string
	for i := 0; i < len(cells); i += cols {
		end := i + cols
		if end > len(cells) {
			end = len(cells)
		}
		rows = append(rows, strings.Join(cells[i:end], " "))
	}

	grid := strings.Join(rows, "\n")

	// Header and legend.
	header := fmt.Sprintf("%s Usage  [h to cycle]", heatmapResourceLabel(resource))
	legend := fmt.Sprintf("Low %s Med %s High %s N/A %s",
		lipgloss.NewStyle().Background(lipgloss.Color("#04B575")).Foreground(lipgloss.Color("#000000")).Padding(0, 1).Render(" "),
		lipgloss.NewStyle().Background(lipgloss.Color("#FFCC00")).Foreground(lipgloss.Color("#000000")).Padding(0, 1).Render(" "),
		lipgloss.NewStyle().Background(lipgloss.Color("#FF4444")).Foreground(lipgloss.Color("#000000")).Padding(0, 1).Render(" "),
		lipgloss.NewStyle().Background(lipgloss.Color("#626262")).Foreground(lipgloss.Color("#000000")).Padding(0, 1).Render(" "),
	)

	content := header + "\n" + legend + "\n\n" + grid
	return ui.SectionStyle.Width(width - 4).Render(content)
}
