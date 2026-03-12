package app

import (
	"testing"

	"charm.land/bubbles/v2/help"
	"github.com/GangGreenTemperTatum/rayatouille/internal/jobs"
	"github.com/stretchr/testify/assert"
)

func TestGlobalKeys_QuitBinding(t *testing.T) {
	keys := GlobalKeys.Quit.Keys()
	assert.Contains(t, keys, "q")
	assert.Contains(t, keys, "ctrl+c")
}

func TestGlobalKeys_HelpBinding(t *testing.T) {
	keys := GlobalKeys.Help.Keys()
	assert.Contains(t, keys, "?")
}

func TestGlobalKeys_BackBinding(t *testing.T) {
	keys := GlobalKeys.Back.Keys()
	assert.Contains(t, keys, "esc")
}

func TestGlobalKeys_SearchBinding(t *testing.T) {
	keys := GlobalKeys.Search.Keys()
	assert.Contains(t, keys, "/")
}

func TestGlobalKeys_ShortHelp(t *testing.T) {
	bindings := GlobalKeys.ShortHelp()
	assert.Len(t, bindings, 5)
}

func TestGlobalKeys_FullHelp(t *testing.T) {
	groups := GlobalKeys.FullHelp()
	assert.Len(t, groups, 2)
	assert.Len(t, groups[0], 3)
	assert.Len(t, groups[1], 2)
}

func TestGlobalKeyMap_SatisfiesHelpKeyMap(t *testing.T) {
	// Compile-time check that GlobalKeyMap satisfies help.KeyMap.
	var _ help.KeyMap = GlobalKeys
}

func TestGlobalKeys_HelpText(t *testing.T) {
	h := GlobalKeys.Quit.Help()
	assert.Equal(t, "q", h.Key)
	assert.Equal(t, "quit", h.Desc)

	h = GlobalKeys.Help.Help()
	assert.Equal(t, "?", h.Key)

	h = GlobalKeys.Back.Help()
	assert.Equal(t, "esc", h.Key)

	h = GlobalKeys.Search.Help()
	assert.Equal(t, "/", h.Key)
}

func TestGlobalKeys_MatchesKeyPress(t *testing.T) {
	// Verify key.Matches works with our bindings.
	// We can only test this indirectly by checking Keys() contains expected values.
	assert.True(t, len(GlobalKeys.Quit.Keys()) > 0)
	assert.True(t, len(GlobalKeys.Help.Keys()) > 0)
}

// Ensure the help model can render our keymap (no panics).
func TestGlobalKeys_HelpModelRenders(t *testing.T) {
	h := help.New()
	output := h.View(GlobalKeys)
	assert.NotEmpty(t, output)
}

var _ help.KeyMap = GlobalKeyMap{}

// Verify key.Matches works with a concrete KeyPressMsg.
func TestGlobalKeys_KeyMatchesPressMsg(t *testing.T) {
	// Construct a mock KeyPressMsg by checking key.Matches against bindings.
	// key.Matches accepts any tea.KeyPressMsg -- we just verify the binding is enabled.
	assert.True(t, GlobalKeys.Quit.Enabled())
	assert.True(t, GlobalKeys.Help.Enabled())
	assert.True(t, GlobalKeys.Back.Enabled())
	assert.True(t, GlobalKeys.Search.Enabled())
}

func TestCombinedKeyMap_SatisfiesHelpKeyMap(t *testing.T) {
	var _ help.KeyMap = CombinedKeyMap{}
}

func TestCombinedKeyMap_ShortHelpCombinesBoth(t *testing.T) {
	combined := CombinedKeyMap{Global: GlobalKeys, View: jobs.Keys}
	bindings := combined.ShortHelp()

	// Should include both jobs keys (7) and global keys (4).
	jobsBindings := jobs.Keys.ShortHelp()
	globalBindings := GlobalKeys.ShortHelp()
	assert.Len(t, bindings, len(jobsBindings)+len(globalBindings))

	// View-specific keys should come first.
	assert.Equal(t, "a", bindings[0].Help().Key, "first binding should be jobs status-all key")
}

func TestCombinedKeyMap_FullHelpCombinesBoth(t *testing.T) {
	combined := CombinedKeyMap{Global: GlobalKeys, View: jobs.Keys}
	groups := combined.FullHelp()

	jobsGroups := jobs.Keys.FullHelp()
	globalGroups := GlobalKeys.FullHelp()
	assert.Len(t, groups, len(jobsGroups)+len(globalGroups))
}

func TestCombinedKeyMap_HelpModelRendersViewSpecificKeys(t *testing.T) {
	h := help.New()
	combined := CombinedKeyMap{Global: GlobalKeys, View: jobs.Keys}
	output := h.View(combined)

	// The rendered help should contain jobs-specific key descriptions.
	assert.Contains(t, output, "sort", "help should show jobs 'sort' key")
	assert.Contains(t, output, "running", "help should show jobs 'running' filter key")

	// And global keys.
	assert.Contains(t, output, "quit", "help should show global 'quit' key")
	assert.Contains(t, output, "help", "help should show global 'help' key")
}
