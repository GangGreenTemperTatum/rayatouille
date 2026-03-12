package ui

import (
	"charm.land/bubbles/v2/key"
	"charm.land/lipgloss/v2"
)

// Color palette for consistent styling across all views.
var (
	ColorPrimary   = lipgloss.Color("#7D56F4")
	ColorSecondary = lipgloss.Color("#5A56E0")
	ColorAccent    = lipgloss.Color("#EE6FF8")
	ColorSuccess   = lipgloss.Color("#04B575")
	ColorWarning   = lipgloss.Color("#FFCC00")
	ColorDanger    = lipgloss.Color("#FF4444")
	ColorMuted     = lipgloss.Color("#626262")
	ColorBg        = lipgloss.Color("#1A1A2E")
	ColorFg        = lipgloss.Color("#FAFAFA")
)

// Reusable styles for consistent rendering across views.
var (
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorFg).
			MarginBottom(1)

	SectionStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorMuted).
			Padding(1, 2)

	StatusBarStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#333333")).
			Foreground(lipgloss.Color("#AAAAAA")).
			Padding(0, 1)

	LabelStyle = lipgloss.NewStyle().
			Foreground(ColorMuted).
			Width(10)

	ValueStyle = lipgloss.NewStyle().
			Foreground(ColorFg).
			Bold(true)

	hintKeyStyle  = lipgloss.NewStyle().Foreground(ColorAccent).Bold(true)
	hintDescStyle = lipgloss.NewStyle().Foreground(ColorMuted)
)

// HintPair represents a single key hint (e.g., "r" → "running").
type HintPair struct {
	Key  string
	Desc string
}

// BindingHint extracts a HintPair from a key.Binding's help text.
// The key label comes from Help().Key, the description from Help().Desc.
func BindingHint(b key.Binding) HintPair {
	h := b.Help()
	return HintPair{Key: h.Key, Desc: h.Desc}
}

// RenderHints renders a compact hotkey hint bar from key-description pairs.
func RenderHints(hints []HintPair) string {
	parts := make([]string, len(hints))
	for i, h := range hints {
		parts[i] = hintKeyStyle.Render(h.Key) + " " + hintDescStyle.Render(h.Desc)
	}
	sep := hintDescStyle.Render("  ")
	result := ""
	for i, p := range parts {
		if i > 0 {
			result += sep
		}
		result += p
	}
	return result
}
