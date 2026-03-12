package servedetail

import (
	image_color "image/color"

	"charm.land/lipgloss/v2"

	"github.com/GangGreenTemperTatum/rayatouille/internal/ray"
	"github.com/GangGreenTemperTatum/rayatouille/internal/ui"
)

// renderMetadata renders application metadata as key-value pairs.
func renderMetadata(app ray.ApplicationDetails, name string, width int) string {
	labelStyle := ui.LabelStyle.Width(16)
	valueStyle := ui.ValueStyle

	rows := []string{
		renderRow(labelStyle, valueStyle, "Application", name),
		renderRow(labelStyle, valueStyle, "Status", appStatusStyled(app.Status)),
		renderRow(labelStyle, valueStyle, "Route", derefOrDash(app.RoutePrefix)),
		renderRow(labelStyle, valueStyle, "Docs Path", derefOrDash(app.DocsPath)),
		renderRow(labelStyle, valueStyle, "Last Deployed", formatLastDeployed(app.LastDeployedTimeS)),
		renderRow(labelStyle, valueStyle, "Message", orDash(app.Message)),
	}

	content := lipgloss.JoinVertical(lipgloss.Left, rows...)
	sectionWidth := width - 6
	if sectionWidth < 40 {
		sectionWidth = 40
	}
	return ui.SectionStyle.Width(sectionWidth).Render(content)
}

// appStatusStyled returns the application status with color coding.
func appStatusStyled(status string) string {
	var c image_color.Color
	switch status {
	case "RUNNING":
		c = ui.ColorSuccess
	case "DEPLOYING", "NOT_STARTED":
		c = ui.ColorWarning
	case "DEPLOY_FAILED", "UNHEALTHY":
		c = ui.ColorDanger
	case "DELETING":
		c = ui.ColorMuted
	default:
		c = ui.ColorMuted
	}
	return lipgloss.NewStyle().Foreground(c).Bold(true).Render(status)
}

// renderRow renders a single label: value row.
func renderRow(labelStyle, valueStyle lipgloss.Style, label, value string) string {
	return labelStyle.Render(label+":") + " " + valueStyle.Render(value)
}

// derefOrDash returns the dereferenced string pointer value or "-" if nil.
func derefOrDash(s *string) string {
	if s == nil {
		return "-"
	}
	if *s == "" {
		return "-"
	}
	return *s
}

// orDash returns the value if non-empty, otherwise "-".
func orDash(s string) string {
	if s == "" {
		return "-"
	}
	return s
}

// formatLastDeployed formats a Unix timestamp as relative time, or "-" if zero.
func formatLastDeployed(ts float64) string {
	if ts <= 0 {
		return "-"
	}
	return formatRelativeTime(ts)
}
