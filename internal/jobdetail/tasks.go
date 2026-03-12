package jobdetail

import (
	"fmt"
	"image/color"
	"sort"
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/GangGreenTemperTatum/rayatouille/internal/ray"
	"github.com/GangGreenTemperTatum/rayatouille/internal/ui"
)

// renderTasks renders the task summary table from the TaskSummaryResponse.
func renderTasks(summary *ray.TaskSummaryResponse, loading bool, taskErr error, width int) string {
	if loading {
		return lipgloss.NewStyle().Foreground(ui.ColorMuted).Render("Fetching task summary...")
	}

	if taskErr != nil {
		return lipgloss.NewStyle().Foreground(ui.ColorDanger).Render(
			fmt.Sprintf("Error fetching tasks: %s", taskErr.Error()),
		)
	}

	if summary == nil {
		return lipgloss.NewStyle().Foreground(ui.ColorMuted).Render("No tasks found for this job")
	}

	cluster, ok := summary.NodeIDToSummary["cluster"]
	if !ok {
		return lipgloss.NewStyle().Foreground(ui.ColorMuted).Render("No tasks found for this job")
	}

	if len(cluster.Summary) == 0 {
		return lipgloss.NewStyle().Foreground(ui.ColorMuted).Render("No tasks found for this job")
	}

	// Sort function names for deterministic display.
	funcNames := make([]string, 0, len(cluster.Summary))
	for name := range cluster.Summary {
		funcNames = append(funcNames, name)
	}
	sort.Strings(funcNames)

	// Header.
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(ui.ColorFg)
	header := fmt.Sprintf("  %-40s  %-24s  %s", "Function Name", "Type", "State Counts")
	rows := []string{headerStyle.Render(header)}
	rows = append(rows, lipgloss.NewStyle().Foreground(ui.ColorMuted).Render(strings.Repeat("-", min(width-8, 100))))

	// Rows.
	for _, name := range funcNames {
		entry := cluster.Summary[name]
		funcName := entry.FuncOrClassName
		if len(funcName) > 40 {
			funcName = funcName[:37] + "..."
		}

		counts := formatStateCounts(entry.StateCounts)
		row := fmt.Sprintf("  %-40s  %-24s  %s", funcName, entry.Type, counts)
		rows = append(rows, row)
	}

	// Totals.
	rows = append(rows, "")
	totals := lipgloss.NewStyle().Foreground(ui.ColorMuted).Render(
		fmt.Sprintf("  Total Tasks: %d | Actor Tasks: %d | Actor Scheduled: %d",
			cluster.TotalTasks, cluster.TotalActorTasks, cluster.TotalActorScheduled),
	)
	rows = append(rows, totals)

	content := lipgloss.JoinVertical(lipgloss.Left, rows...)

	sectionWidth := width - 6
	if sectionWidth < 40 {
		sectionWidth = 40
	}

	return ui.SectionStyle.Width(sectionWidth).Render(content)
}

// formatStateCounts formats state counts as colored "STATE: N" pairs.
func formatStateCounts(counts map[string]int) string {
	if len(counts) == 0 {
		return "-"
	}

	// Sort states for deterministic output.
	states := make([]string, 0, len(counts))
	for state := range counts {
		states = append(states, state)
	}
	sort.Strings(states)

	parts := make([]string, 0, len(states))
	for _, state := range states {
		count := counts[state]
		c := stateColor(state)
		styled := lipgloss.NewStyle().Foreground(c).Render(fmt.Sprintf("%s: %d", state, count))
		parts = append(parts, styled)
	}

	return strings.Join(parts, ", ")
}

// stateColor maps a task state to a theme color.
func stateColor(state string) color.Color {
	switch state {
	case "FINISHED":
		return ui.ColorSuccess
	case "RUNNING":
		return ui.ColorWarning
	case "FAILED":
		return ui.ColorDanger
	case "PENDING", "PENDING_ARGS_AVAIL", "PENDING_NODE_ASSIGNMENT", "PENDING_OBJ_STORE_MEM_AVAIL":
		return ui.ColorMuted
	default:
		return ui.ColorFg
	}
}

// min returns the smaller of two ints.
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
