package ui

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
)

func TestNewFilter_InitialState(t *testing.T) {
	f := NewFilter()
	assert.False(t, f.Active())
	assert.Equal(t, "", f.Value())
}

func TestFilter_ActivateDeactivate(t *testing.T) {
	f := NewFilter()

	cmd := f.Activate()
	assert.True(t, f.Active())
	assert.NotNil(t, cmd, "Activate should return focus command")

	f.Deactivate()
	assert.False(t, f.Active())
}

func TestFilter_Matches_EmptyValue(t *testing.T) {
	f := NewFilter()
	assert.True(t, f.Matches("anything"))
	assert.True(t, f.Matches(""))
}

func TestFilter_Matches_CaseInsensitive(t *testing.T) {
	f := NewFilter()
	f.value = "hello"

	assert.True(t, f.Matches("Hello World"))
	assert.True(t, f.Matches("HELLO"))
	assert.True(t, f.Matches("say hello"))
	assert.False(t, f.Matches("goodbye"))
}

func TestFilter_Matches_SubstringMatch(t *testing.T) {
	f := NewFilter()
	f.value = "run"

	assert.True(t, f.Matches("RUNNING"))
	assert.True(t, f.Matches("running"))
	assert.True(t, f.Matches("forerunner"))
	assert.False(t, f.Matches("PENDING"))
}

func TestFilter_Update_EnterCommitsValue(t *testing.T) {
	f := NewFilter()
	f.Activate()
	f.input.SetValue("test-filter")

	updated, _ := f.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

	assert.Equal(t, "test-filter", updated.Value())
	assert.False(t, updated.Active())
}

func TestFilter_Update_EscRestoresPreviousValue(t *testing.T) {
	f := NewFilter()
	f.value = "original"
	f.Activate()
	f.input.SetValue("changed")

	updated, _ := f.Update(tea.KeyPressMsg{Code: tea.KeyEscape})

	assert.Equal(t, "original", updated.Value())
	assert.False(t, updated.Active())
	// Input should be restored to original value.
	assert.Equal(t, "original", updated.input.Value())
}

func TestFilter_Update_WhenNotActive(t *testing.T) {
	f := NewFilter()
	// Not active, should return unchanged.
	updated, cmd := f.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

	assert.False(t, updated.Active())
	assert.Equal(t, "", updated.Value())
	assert.Nil(t, cmd)
}

func TestFilter_Clear(t *testing.T) {
	f := NewFilter()
	f.value = "something"
	f.Activate()
	f.input.SetValue("something")

	f.Clear()

	assert.Equal(t, "", f.Value())
	assert.False(t, f.Active())
	assert.Equal(t, "", f.input.Value())
}

func TestFilter_View_EmptyAndInactive(t *testing.T) {
	f := NewFilter()
	assert.Equal(t, "", f.View())
}

func TestFilter_View_ActiveShowsInput(t *testing.T) {
	f := NewFilter()
	f.Activate()
	view := f.View()
	assert.NotEmpty(t, view)
}

func TestFilter_View_InactiveWithValueShowsInput(t *testing.T) {
	f := NewFilter()
	f.value = "active-filter"
	f.input.SetValue("active-filter")
	// Not active but has value -- shows the input.
	view := f.View()
	assert.NotEmpty(t, view)
}
