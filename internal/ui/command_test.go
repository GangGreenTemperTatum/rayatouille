package ui

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseCommand_Jobs(t *testing.T) {
	result := ParseCommand("jobs")
	require.NotNil(t, result)
	assert.Equal(t, CmdJobs, result.Command)
}

func TestParseCommand_Dashboard(t *testing.T) {
	result := ParseCommand("dashboard")
	require.NotNil(t, result)
	assert.Equal(t, CmdDashboard, result.Command)
}

func TestParseCommand_DashShorthand(t *testing.T) {
	result := ParseCommand("dash")
	require.NotNil(t, result)
	assert.Equal(t, CmdDashboard, result.Command)
}

func TestParseCommand_Nodes(t *testing.T) {
	result := ParseCommand("nodes")
	require.NotNil(t, result)
	assert.Equal(t, CmdNodes, result.Command)
}

func TestParseCommand_NodesCaseInsensitive(t *testing.T) {
	result := ParseCommand("NODES")
	require.NotNil(t, result)
	assert.Equal(t, CmdNodes, result.Command)
}

func TestParseCommand_Unknown(t *testing.T) {
	result := ParseCommand("unknown")
	assert.Nil(t, result)
}

func TestParseCommand_WithWhitespace(t *testing.T) {
	result := ParseCommand("  jobs  ")
	require.NotNil(t, result)
	assert.Equal(t, CmdJobs, result.Command)
}

func TestParseCommand_CaseInsensitive(t *testing.T) {
	result := ParseCommand("JOBS")
	require.NotNil(t, result)
	assert.Equal(t, CmdJobs, result.Command)

	result = ParseCommand("Dashboard")
	require.NotNil(t, result)
	assert.Equal(t, CmdDashboard, result.Command)
}

func TestParseCommand_Serve(t *testing.T) {
	result := ParseCommand("serve")
	require.NotNil(t, result)
	assert.Equal(t, CmdServe, result.Command)
}

func TestParseCommand_Events(t *testing.T) {
	result := ParseCommand("events")
	require.NotNil(t, result)
	assert.Equal(t, CmdEvents, result.Command)
}

func TestParseCommand_Empty(t *testing.T) {
	result := ParseCommand("")
	assert.Nil(t, result)
}

func TestParseCommand_Profile(t *testing.T) {
	result := ParseCommand("profile production")
	require.NotNil(t, result)
	assert.Equal(t, CmdProfile, result.Command)
	assert.Equal(t, "production", result.Arg)
}

func TestParseCommand_ProfileNoArg(t *testing.T) {
	result := ParseCommand("profile")
	require.NotNil(t, result)
	assert.Equal(t, CmdProfile, result.Command)
	assert.Equal(t, "", result.Arg)
}

func TestParseCommand_ProfileWithExtraSpaces(t *testing.T) {
	result := ParseCommand("  profile   staging  ")
	require.NotNil(t, result)
	assert.Equal(t, CmdProfile, result.Command)
	assert.Equal(t, "staging", result.Arg)
}

func TestCommand_ActivateDeactivate(t *testing.T) {
	c := NewCommand()
	assert.False(t, c.Active())

	cmd := c.Activate()
	assert.True(t, c.Active())
	assert.NotNil(t, cmd)

	c.Deactivate()
	assert.False(t, c.Active())
}

func TestCommand_Update_EnterWithValidCommand(t *testing.T) {
	c := NewCommand()
	c.Activate()
	c.input.SetValue("jobs")

	updated, _, result := c.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

	require.NotNil(t, result)
	assert.Equal(t, CmdJobs, result.Command)
	assert.False(t, updated.Active())
}

func TestCommand_Update_EnterWithUnknownCommand(t *testing.T) {
	c := NewCommand()
	c.Activate()
	c.input.SetValue("foobar")

	updated, _, result := c.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

	assert.Nil(t, result)
	assert.False(t, updated.Active())
}

func TestCommand_Update_EscDeactivates(t *testing.T) {
	c := NewCommand()
	c.Activate()
	c.input.SetValue("jobs")

	updated, _, result := c.Update(tea.KeyPressMsg{Code: tea.KeyEscape})

	assert.Nil(t, result)
	assert.False(t, updated.Active())
}

func TestCommand_Update_WhenNotActive(t *testing.T) {
	c := NewCommand()
	updated, cmd, result := c.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

	assert.False(t, updated.Active())
	assert.Nil(t, cmd)
	assert.Nil(t, result)
}

func TestCommand_View_NotActive(t *testing.T) {
	c := NewCommand()
	assert.Equal(t, "", c.View())
}

func TestCommand_View_Active(t *testing.T) {
	c := NewCommand()
	c.Activate()
	view := c.View()
	assert.NotEmpty(t, view)
}
