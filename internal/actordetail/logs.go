package actordetail

import (
	"fmt"

	"charm.land/lipgloss/v2"

	"github.com/GangGreenTemperTatum/rayatouille/internal/ui"
)

// renderLogs renders the logs tab content.
func renderLogs(m Model) string {
	if m.logLoading {
		return lipgloss.NewStyle().Foreground(ui.ColorMuted).Render("Loading actor logs...")
	}

	if m.logErr != nil {
		return lipgloss.NewStyle().Foreground(ui.ColorDanger).Render(
			fmt.Sprintf("Error: %s", m.logErr.Error()),
		)
	}

	if m.logContent == "" {
		return lipgloss.NewStyle().Foreground(ui.ColorMuted).Render("No logs available for this actor.")
	}

	return m.logViewer.View()
}
