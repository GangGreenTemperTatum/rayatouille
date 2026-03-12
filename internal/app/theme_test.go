package app

import (
	"image/color"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTitleStyle_RendersNonEmpty(t *testing.T) {
	output := TitleStyle.Render("Hello")
	assert.NotEmpty(t, output)
	assert.Contains(t, output, "Hello")
}

func TestSectionStyle_IncludesBorderCharacters(t *testing.T) {
	output := SectionStyle.Render("Content")
	assert.NotEmpty(t, output)
	// Rounded border uses characters like "╭", "╮", "│", "╰", "╯"
	assert.Contains(t, output, "╭")
	assert.Contains(t, output, "╯")
}

func TestColorConstants_AreNotNil(t *testing.T) {
	colors := []struct {
		name  string
		color color.Color
	}{
		{"ColorPrimary", ColorPrimary},
		{"ColorSecondary", ColorSecondary},
		{"ColorAccent", ColorAccent},
		{"ColorSuccess", ColorSuccess},
		{"ColorWarning", ColorWarning},
		{"ColorDanger", ColorDanger},
		{"ColorMuted", ColorMuted},
		{"ColorBg", ColorBg},
		{"ColorFg", ColorFg},
	}
	for _, c := range colors {
		t.Run(c.name, func(t *testing.T) {
			assert.NotNil(t, c.color, "%s should not be nil", c.name)
		})
	}
}
