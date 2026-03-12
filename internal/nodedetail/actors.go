package nodedetail

import (
	"fmt"
	image_color "image/color"
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/GangGreenTemperTatum/rayatouille/internal/ray"
	"github.com/GangGreenTemperTatum/rayatouille/internal/ui"
)

// renderActors renders the actors tab content.
func renderActors(actors []ray.Actor, loading bool, err error, width int) string {
	if loading {
		return lipgloss.NewStyle().Foreground(ui.ColorMuted).Render("Fetching actors...")
	}

	if err != nil {
		return lipgloss.NewStyle().Foreground(ui.ColorDanger).Render(
			fmt.Sprintf("Error fetching actors: %s", err.Error()),
		)
	}

	if len(actors) == 0 {
		return lipgloss.NewStyle().Foreground(ui.ColorMuted).Render("No actors on this node")
	}

	// Header.
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(ui.ColorFg)
	header := fmt.Sprintf("  %-14s  %-30s  %-10s  %s", "Actor ID", "Class Name", "State", "PID")
	rows := []string{headerStyle.Render(header)}
	rows = append(rows, lipgloss.NewStyle().Foreground(ui.ColorMuted).Render(strings.Repeat("-", min(width-8, 80))))

	for _, a := range actors {
		actorID := a.ActorID
		if len(actorID) > 12 {
			actorID = actorID[:12]
		}

		className := a.ClassName
		if len(className) > 30 {
			className = className[:27] + "..."
		}

		stateStr := actorStateStyled(a.State)

		row := fmt.Sprintf("  %-14s  %-30s  %s  %d", actorID, className, stateStr, a.PID)
		rows = append(rows, row)
	}

	// Total count.
	rows = append(rows, "")
	total := lipgloss.NewStyle().Foreground(ui.ColorMuted).Render(
		fmt.Sprintf("  Total: %d actors", len(actors)),
	)
	rows = append(rows, total)

	content := lipgloss.JoinVertical(lipgloss.Left, rows...)

	sectionWidth := width - 6
	if sectionWidth < 40 {
		sectionWidth = 40
	}

	return ui.SectionStyle.Width(sectionWidth).Render(content)
}

// actorStateStyled returns the actor state with color coding.
func actorStateStyled(state string) string {
	var c image_color.Color
	switch state {
	case "ALIVE":
		c = ui.ColorSuccess
	case "DEAD":
		c = ui.ColorDanger
	case "PENDING":
		c = ui.ColorWarning
	default:
		c = ui.ColorFg
	}
	return lipgloss.NewStyle().Foreground(c).Render(fmt.Sprintf("%-10s", state))
}

// min returns the smaller of two ints.
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
