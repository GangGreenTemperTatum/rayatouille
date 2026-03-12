package actordetail

import (
	"fmt"
	image_color "image/color"

	"charm.land/lipgloss/v2"

	"github.com/GangGreenTemperTatum/rayatouille/internal/ray"
	"github.com/GangGreenTemperTatum/rayatouille/internal/ui"
)

// renderInfo renders the actor info panel showing metadata and death cause.
func renderInfo(actor ray.Actor, width int) string {
	labelStyle := ui.LabelStyle.Width(16)
	valueStyle := ui.ValueStyle

	rows := []string{
		renderRow(labelStyle, valueStyle, "Actor ID", actor.ActorID),
		renderRow(labelStyle, valueStyle, "Name", orDash(actor.Name)),
		renderRow(labelStyle, valueStyle, "Class", actor.ClassName),
		renderRow(labelStyle, valueStyle, "State", actorStateStyled(actor.State)),
		renderRow(labelStyle, valueStyle, "PID", fmt.Sprintf("%d", actor.PID)),
		renderRow(labelStyle, valueStyle, "Job ID", actor.JobID),
		renderRow(labelStyle, valueStyle, "Node ID", actor.NodeID),
		renderRow(labelStyle, valueStyle, "Namespace", actor.RayNamespace),
		renderRow(labelStyle, valueStyle, "Detached", fmt.Sprintf("%v", actor.IsDetached)),
		renderRow(labelStyle, valueStyle, "Restarts", orDash(actor.NumRestarts)),
	}

	// Death cause for dead actors.
	if actor.DeathCause != nil && actor.DeathCause.ActorDiedErrorContext != nil {
		ctx := actor.DeathCause.ActorDiedErrorContext
		rows = append(rows, "")
		rows = append(rows, lipgloss.NewStyle().Bold(true).Foreground(ui.ColorDanger).Render("Death Cause:"))
		rows = append(rows, renderRow(labelStyle, valueStyle, "Reason", ctx.Reason))
		rows = append(rows, renderRow(labelStyle, valueStyle, "Error", ctx.ErrorMessage))
	}

	// Required resources section.
	if len(actor.RequiredResources) > 0 {
		rows = append(rows, "")
		rows = append(rows, lipgloss.NewStyle().Bold(true).Foreground(ui.ColorFg).Render("Required Resources:"))
		for k, v := range actor.RequiredResources {
			rows = append(rows, renderRow(labelStyle, valueStyle, k, fmt.Sprintf("%v", v)))
		}
	}

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
	case "PENDING_CREATION", "DEPENDENCIES_UNREADY", "RESTARTING":
		c = ui.ColorWarning
	default:
		c = ui.ColorMuted
	}
	return lipgloss.NewStyle().Foreground(c).Bold(true).Render(state)
}

// renderRow renders a single label: value row.
func renderRow(labelStyle, valueStyle lipgloss.Style, label, value string) string {
	return labelStyle.Render(label+":") + " " + valueStyle.Render(value)
}

// orDash returns the value if non-empty, otherwise "-".
func orDash(s string) string {
	if s == "" {
		return "-"
	}
	return s
}
