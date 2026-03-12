package dashboard

import (
	"fmt"
	"image/color"
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/GangGreenTemperTatum/rayatouille/internal/ui"
)

// renderJobStatusBar renders a proportional segmented bar showing job status distribution.
// Each segment is color-coded and sized proportionally to the count.
//
//	████████████░░░░░░░░░░░░
//	 8 running  2 pending  1 failed  3 done
func (m Model) renderJobStatusBar(width int) string {
	if m.jobs.Total == 0 {
		return ""
	}

	barWidth := width - 6
	if barWidth > 60 {
		barWidth = 60
	}
	if barWidth < 10 {
		barWidth = 10
	}

	type segment struct {
		count int
		char  string
		color color.Color
		label string
	}

	segments := []segment{
		{m.jobs.Running, "█", ui.ColorSuccess, "running"},
		{m.jobs.Pending, "█", ui.ColorWarning, "pending"},
		{m.jobs.Failed, "█", ui.ColorDanger, "failed"},
		{m.jobs.Succeeded + m.jobs.Stopped, "░", ui.ColorMuted, "done"},
	}

	// Build proportional bar.
	var bar strings.Builder
	remaining := barWidth
	for i, seg := range segments {
		if seg.count == 0 {
			continue
		}
		w := barWidth * seg.count / m.jobs.Total
		// Last non-zero segment gets remaining width to avoid rounding gaps.
		isLast := true
		for _, s := range segments[i+1:] {
			if s.count > 0 {
				isLast = false
				break
			}
		}
		if isLast {
			w = remaining
		}
		if w <= 0 {
			w = 1
		}
		remaining -= w
		bar.WriteString(lipgloss.NewStyle().Foreground(seg.color).Render(strings.Repeat(seg.char, w)))
	}

	// Legend line.
	var legendParts []string
	for _, seg := range segments {
		if seg.count == 0 {
			continue
		}
		dot := lipgloss.NewStyle().Foreground(seg.color).Render("●")
		legendParts = append(legendParts, fmt.Sprintf("%s %d %s", dot, seg.count, seg.label))
	}
	legend := lipgloss.NewStyle().Foreground(ui.ColorMuted).Render(strings.Join(legendParts, "  "))

	content := bar.String() + "\n" + legend
	return ui.SectionStyle.Width(width - 4).Render("Job Status\n\n" + content)
}

// renderNodeDots renders a dot grid showing node health at a glance.
// Each dot represents one node: green = alive, red = dead.
//
//	● ● ● ○  3/4 nodes alive
func (m Model) renderNodeDots(width int) string {
	if len(m.nodes) == 0 {
		return ""
	}

	var dots strings.Builder
	alive := 0
	for i, n := range m.nodes {
		if i > 0 {
			dots.WriteString(" ")
		}
		if n.State == "ALIVE" {
			dots.WriteString(lipgloss.NewStyle().Foreground(ui.ColorSuccess).Render("●"))
			alive++
		} else {
			dots.WriteString(lipgloss.NewStyle().Foreground(ui.ColorDanger).Render("○"))
		}
	}

	summary := lipgloss.NewStyle().Foreground(ui.ColorMuted).Render(
		fmt.Sprintf("  %d/%d nodes alive", alive, len(m.nodes)))

	return dots.String() + summary
}
