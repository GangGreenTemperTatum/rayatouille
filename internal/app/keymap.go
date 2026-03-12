package app

import (
	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"

	"github.com/GangGreenTemperTatum/rayatouille/internal/config"
)

// GlobalKeyMap defines keybindings handled at the root model level.
type GlobalKeyMap struct {
	Quit   key.Binding
	Help   key.Binding
	Back   key.Binding
	Search key.Binding
	Tab    key.Binding
}

// ShortHelp returns keybindings for the short help view.
func (k GlobalKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Tab, k.Quit, k.Help, k.Back, k.Search}
}

// FullHelp returns keybindings grouped for the full help view.
func (k GlobalKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Tab, k.Quit, k.Help},
		{k.Back, k.Search},
	}
}

// GlobalKeys is the default set of global keybindings.
var GlobalKeys = GlobalKeyMap{
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "help"),
	),
	Back: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "back"),
	),
	Search: key.NewBinding(
		key.WithKeys("/"),
		key.WithHelp("/", "filter"),
	),
	Tab: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "next view"),
	),
}

// ApplyGlobalBindings applies user keybinding overrides to the global keys.
func ApplyGlobalBindings(b config.GlobalBindings) {
	config.OverrideBinding(&GlobalKeys.Quit, b.Quit)
	config.OverrideBinding(&GlobalKeys.Help, b.Help)
	config.OverrideBinding(&GlobalKeys.Back, b.Back)
	config.OverrideBinding(&GlobalKeys.Search, b.Search)
	config.OverrideBinding(&GlobalKeys.Tab, b.Tab)
}

// CombinedKeyMap merges two help.KeyMaps so the help overlay shows both
// global and view-specific keybindings.
type CombinedKeyMap struct {
	Global help.KeyMap
	View   help.KeyMap
}

// ShortHelp returns the combined short help bindings (view-specific first, then global).
func (c CombinedKeyMap) ShortHelp() []key.Binding {
	all := c.View.ShortHelp()
	all = append(all, c.Global.ShortHelp()...)
	return all
}

// FullHelp returns the combined full help groups (view-specific groups first, then global).
func (c CombinedKeyMap) FullHelp() [][]key.Binding {
	all := c.View.FullHelp()
	all = append(all, c.Global.FullHelp()...)
	return all
}
