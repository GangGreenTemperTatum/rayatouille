package ui

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
)

func TestNewLogViewer_Dimensions(t *testing.T) {
	lv := NewLogViewer(100, 30)
	assert.Equal(t, 100, lv.width)
	assert.Equal(t, 30, lv.height)
	assert.False(t, lv.hasContent)
}

func TestLogViewer_SetContent(t *testing.T) {
	lv := NewLogViewer(80, 20)
	lv.SetContent("hello\nworld")
	assert.True(t, lv.hasContent)
}

func TestLogViewer_View_NoContent(t *testing.T) {
	lv := NewLogViewer(80, 20)
	view := lv.View()
	assert.Contains(t, view, "No logs available")
}

func TestLogViewer_View_WithContent(t *testing.T) {
	lv := NewLogViewer(80, 20)
	lv.SetContent("line 1\nline 2\nline 3")
	view := lv.View()
	assert.NotEmpty(t, view)
	assert.NotContains(t, view, "No logs available")
}

func TestLogViewer_SearchActivation(t *testing.T) {
	lv := NewLogViewer(80, 20)
	lv.SetContent("hello world")

	// Press "/" to activate search.
	lv, _ = lv.Update(tea.KeyPressMsg{Code: -1, Text: "/"})
	assert.True(t, lv.Searching())
}

func TestLogViewer_SearchCancellation(t *testing.T) {
	lv := NewLogViewer(80, 20)
	lv.SetContent("hello world")

	// Activate search.
	lv, _ = lv.Update(tea.KeyPressMsg{Code: -1, Text: "/"})
	assert.True(t, lv.Searching())

	// Cancel with Esc.
	lv, _ = lv.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	assert.False(t, lv.Searching())
}

func TestLogViewer_SearchWithMatches(t *testing.T) {
	lv := NewLogViewer(80, 20)
	lv.SetContent("hello world hello again")

	// Activate search.
	lv, _ = lv.Update(tea.KeyPressMsg{Code: -1, Text: "/"})
	// Type "hello".
	lv, _ = lv.Update(tea.KeyPressMsg{Code: -1, Text: "h"})
	lv, _ = lv.Update(tea.KeyPressMsg{Code: -1, Text: "e"})
	lv, _ = lv.Update(tea.KeyPressMsg{Code: -1, Text: "l"})
	lv, _ = lv.Update(tea.KeyPressMsg{Code: -1, Text: "l"})
	lv, _ = lv.Update(tea.KeyPressMsg{Code: -1, Text: "o"})
	// Confirm search.
	lv, _ = lv.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

	assert.False(t, lv.Searching())
	assert.Equal(t, 2, lv.MatchCount())
}

func TestLogViewer_SearchNoMatches(t *testing.T) {
	lv := NewLogViewer(80, 20)
	lv.SetContent("hello world")

	// Search for something that doesn't exist.
	lv, _ = lv.Update(tea.KeyPressMsg{Code: -1, Text: "/"})
	lv, _ = lv.Update(tea.KeyPressMsg{Code: -1, Text: "x"})
	lv, _ = lv.Update(tea.KeyPressMsg{Code: -1, Text: "y"})
	lv, _ = lv.Update(tea.KeyPressMsg{Code: -1, Text: "z"})
	lv, _ = lv.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

	assert.Equal(t, 0, lv.MatchCount())
}

func TestLogViewer_SearchCaseInsensitive(t *testing.T) {
	lv := NewLogViewer(80, 20)
	lv.SetContent("Hello HELLO hello")

	lv, _ = lv.Update(tea.KeyPressMsg{Code: -1, Text: "/"})
	lv, _ = lv.Update(tea.KeyPressMsg{Code: -1, Text: "h"})
	lv, _ = lv.Update(tea.KeyPressMsg{Code: -1, Text: "e"})
	lv, _ = lv.Update(tea.KeyPressMsg{Code: -1, Text: "l"})
	lv, _ = lv.Update(tea.KeyPressMsg{Code: -1, Text: "l"})
	lv, _ = lv.Update(tea.KeyPressMsg{Code: -1, Text: "o"})
	lv, _ = lv.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

	assert.Equal(t, 3, lv.MatchCount())
}

func TestLogViewer_SetSize(t *testing.T) {
	lv := NewLogViewer(80, 20)
	lv.SetSize(120, 40)
	assert.Equal(t, 120, lv.width)
	assert.Equal(t, 40, lv.height)
}
