package dashboard

import (
	"fmt"
	"image/color"
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/GangGreenTemperTatum/rayatouille/internal/ui"
)

// ringCell defines one cell of the donut ring perimeter.
type ringCell struct {
	row, col int
	char     string
}

// 22-segment donut ring, 5 lines × 11 cols.
// Ordered clockwise from top-left so proportional coloring sweeps naturally.
//
//	   ▄▄▄▄▄
//	 ██     ██
//	██  N/T  ██
//	 ██     ██
//	   ▀▀▀▀▀
var ring = []ringCell{
	// Top arc (left to right)
	{0, 3, "▄"}, {0, 4, "▄"}, {0, 5, "▄"}, {0, 6, "▄"}, {0, 7, "▄"},
	// Right side (top to bottom)
	{1, 8, "█"}, {1, 9, "█"},
	{2, 9, "█"}, {2, 10, "█"},
	// Right bottom
	{3, 8, "█"}, {3, 9, "█"},
	// Bottom arc (right to left)
	{4, 7, "▀"}, {4, 6, "▀"}, {4, 5, "▀"}, {4, 4, "▀"}, {4, 3, "▀"},
	// Left bottom
	{3, 1, "█"}, {3, 2, "█"},
	// Left side (bottom to top)
	{2, 0, "█"}, {2, 1, "█"},
	// Left top
	{1, 1, "█"}, {1, 2, "█"},
}

// donutSlice represents one proportional segment of the donut.
type donutSlice struct {
	count int
	color color.Color
	label string
}

// renderDonutChart renders a colored donut chart showing cluster health.
// The ring shows job status distribution; the center shows node health.
// Returns empty string if no data exists.
func renderDonutChart(jobs JobSummary, health ClusterHealth) string {
	if jobs.Total == 0 && health.NodeCount == 0 {
		return ""
	}

	// Build ring slices from jobs.
	var slices []donutSlice
	if jobs.Total > 0 {
		slices = []donutSlice{
			{jobs.Running, ui.ColorSuccess, "run"},
			{jobs.Pending, ui.ColorWarning, "pend"},
			{jobs.Failed, ui.ColorDanger, "fail"},
			{jobs.Succeeded + jobs.Stopped, ui.ColorMuted, "done"},
		}
	} else {
		// No jobs — show a muted ring.
		slices = []donutSlice{
			{1, ui.ColorMuted, "none"},
		}
	}

	// Assign a color to each ring segment proportionally.
	total := 0
	for _, sl := range slices {
		total += sl.count
	}
	totalSegments := len(ring)
	segColors := make([]color.Color, totalSegments)
	idx := 0
	for _, sl := range slices {
		if sl.count == 0 {
			continue
		}
		n := totalSegments * sl.count / total
		if n == 0 && sl.count > 0 {
			n = 1 // At least 1 segment for any non-zero slice.
		}
		for i := 0; i < n && idx < totalSegments; i++ {
			segColors[idx] = sl.color
			idx++
		}
	}
	// Fill remaining with last color (handles rounding).
	lastColor := ui.ColorMuted
	for _, sl := range slices {
		if sl.count > 0 {
			lastColor = sl.color
		}
	}
	for idx < totalSegments {
		segColors[idx] = lastColor
		idx++
	}

	// Build a 5×11 grid initialized with spaces.
	const rows, cols = 5, 11
	grid := make([][]string, rows)
	for r := range grid {
		grid[r] = make([]string, cols)
		for c := range grid[r] {
			grid[r][c] = " "
		}
	}

	// Place colored ring segments.
	for i, cell := range ring {
		styled := lipgloss.NewStyle().Foreground(segColors[i]).Render(cell.char)
		grid[cell.row][cell.col] = styled
	}

	// Center text: show node health as colored fraction.
	var centerText string
	var centerColor color.Color
	if health.NodeCount > 0 {
		centerText = fmt.Sprintf("%d/%dN", health.AliveNodes, health.NodeCount)
		switch health.Status {
		case "healthy":
			centerColor = ui.ColorSuccess
		case "degraded":
			centerColor = ui.ColorWarning
		default:
			centerColor = ui.ColorDanger
		}
	} else {
		centerText = fmt.Sprintf("%dJ", jobs.Total)
		centerColor = ui.ColorFg
	}
	if len(centerText) > 5 {
		centerText = centerText[:5]
	}
	pad := 5 - len(centerText)
	left := pad / 2
	padded := strings.Repeat(" ", left) + centerText + strings.Repeat(" ", 5-left-len(centerText))
	styledCenter := lipgloss.NewStyle().Bold(true).Foreground(centerColor).Render(padded)
	grid[2][3] = styledCenter
	grid[2][4] = ""
	grid[2][5] = ""
	grid[2][6] = ""
	grid[2][7] = ""

	// Render grid to string.
	var lines []string
	for _, row := range grid {
		lines = append(lines, strings.Join(row, ""))
	}

	// Legend: job status counts + node count.
	var legendParts []string
	type legendEntry struct {
		label string
		color color.Color
		count int
	}
	entries := []legendEntry{
		{"run", ui.ColorSuccess, jobs.Running},
		{"pend", ui.ColorWarning, jobs.Pending},
		{"fail", ui.ColorDanger, jobs.Failed},
		{"done", ui.ColorMuted, jobs.Succeeded + jobs.Stopped},
	}
	for _, e := range entries {
		if e.count == 0 {
			continue
		}
		dot := lipgloss.NewStyle().Foreground(e.color).Render("●")
		legendParts = append(legendParts, fmt.Sprintf("%s%d%s", dot, e.count, e.label))
	}
	legend := strings.Join(legendParts, " ")

	lines = append(lines, legend)

	return strings.Join(lines, "\n")
}
