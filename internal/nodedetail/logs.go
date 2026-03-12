package nodedetail

import (
	"fmt"
	"sort"
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/GangGreenTemperTatum/rayatouille/internal/ui"
)

// renderLogFileList renders the log file browser showing files grouped by category.
func renderLogFileList(logFiles []LogFile, selectedIdx int, loading bool, err error, nodeAlive bool, width int) string {
	if !nodeAlive {
		return lipgloss.NewStyle().Foreground(ui.ColorMuted).Render("Logs unavailable for dead nodes")
	}

	if loading {
		return lipgloss.NewStyle().Foreground(ui.ColorMuted).Render("Fetching log files...")
	}

	if err != nil {
		return lipgloss.NewStyle().Foreground(ui.ColorDanger).Render(
			fmt.Sprintf("Error fetching log files: %s", err.Error()),
		)
	}

	if len(logFiles) == 0 {
		return lipgloss.NewStyle().Foreground(ui.ColorMuted).Render("No log files found")
	}

	// Group files by category for display.
	var rows []string
	currentCategory := ""

	for i, lf := range logFiles {
		if lf.Category != currentCategory {
			if currentCategory != "" {
				rows = append(rows, "") // blank line between categories
			}
			currentCategory = lf.Category
			catLabel := lipgloss.NewStyle().Bold(true).Foreground(ui.ColorFg).Render(lf.Category + "/")
			rows = append(rows, catLabel)
		}

		prefix := "  "
		style := lipgloss.NewStyle().Foreground(ui.ColorMuted)
		if i == selectedIdx {
			prefix = "> "
			style = lipgloss.NewStyle().Foreground(ui.ColorPrimary).Bold(true)
		}

		rows = append(rows, style.Render(prefix+lf.Filename))
	}

	// Footer with hint.
	rows = append(rows, "")
	hint := lipgloss.NewStyle().Foreground(ui.ColorMuted).Render(
		fmt.Sprintf("  %d files | Press Enter to view", len(logFiles)),
	)
	rows = append(rows, hint)

	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}

// renderLogContent renders either the file list or the log viewer depending on state.
func renderLogContent(m Model) string {
	if m.viewingLog {
		if m.logContentLoading {
			return lipgloss.NewStyle().Foreground(ui.ColorMuted).Render("Fetching log content...")
		}
		if m.logContentErr != nil {
			return lipgloss.NewStyle().Foreground(ui.ColorDanger).Render(
				fmt.Sprintf("Error: %s", m.logContentErr.Error()),
			)
		}
		// Show filename header + log viewer.
		var filename string
		if m.selectedLogFile >= 0 && m.selectedLogFile < len(m.logFiles) {
			filename = m.logFiles[m.selectedLogFile].Filename
		}
		header := lipgloss.NewStyle().Bold(true).Foreground(ui.ColorFg).Render(filename) +
			lipgloss.NewStyle().Foreground(ui.ColorMuted).Render(" (Esc to go back)")
		return lipgloss.JoinVertical(lipgloss.Left, header, "", m.logViewer.View())
	}

	nodeAlive := strings.EqualFold(m.node.State, "ALIVE")
	return renderLogFileList(m.logFiles, m.selectedLogFile, m.logListLoading, m.logListErr, nodeAlive, m.width)
}

// flattenLogFiles converts a categorized map of log files into a sorted flat list.
func flattenLogFiles(categories map[string][]string) []LogFile {
	if len(categories) == 0 {
		return nil
	}

	// Sort categories for deterministic order.
	catNames := make([]string, 0, len(categories))
	for cat := range categories {
		catNames = append(catNames, cat)
	}
	sort.Strings(catNames)

	var files []LogFile
	for _, cat := range catNames {
		filenames := categories[cat]
		sorted := make([]string, len(filenames))
		copy(sorted, filenames)
		sort.Strings(sorted)
		for _, f := range sorted {
			files = append(files, LogFile{Category: cat, Filename: f})
		}
	}

	return files
}
