package app

import (
	"github.com/GangGreenTemperTatum/rayatouille/internal/ui"
)

// Re-export theme constants from the ui package for backward compatibility.
// New code should import internal/ui directly.
var (
	ColorPrimary   = ui.ColorPrimary
	ColorSecondary = ui.ColorSecondary
	ColorAccent    = ui.ColorAccent
	ColorSuccess   = ui.ColorSuccess
	ColorWarning   = ui.ColorWarning
	ColorDanger    = ui.ColorDanger
	ColorMuted     = ui.ColorMuted
	ColorBg        = ui.ColorBg
	ColorFg        = ui.ColorFg

	TitleStyle     = ui.TitleStyle
	SectionStyle   = ui.SectionStyle
	StatusBarStyle = ui.StatusBarStyle
	LabelStyle     = ui.LabelStyle
	ValueStyle     = ui.ValueStyle
)
