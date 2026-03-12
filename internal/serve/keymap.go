package serve

import (
	"charm.land/bubbles/v2/key"

	"github.com/GangGreenTemperTatum/rayatouille/internal/config"
)

// KeyMap defines keybindings specific to the serve applications list view.
type KeyMap struct {
	StatusAll       key.Binding
	StatusRunning   key.Binding
	StatusDeploying key.Binding
	StatusFailed    key.Binding
	Sort            key.Binding
	Filter          key.Binding
	Enter           key.Binding
}

// ShortHelp returns keybindings for the short help view.
func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.StatusAll, k.StatusRunning, k.StatusDeploying, k.StatusFailed, k.Sort, k.Filter, k.Enter}
}

// FullHelp returns keybindings grouped for the full help view.
func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.StatusAll, k.StatusRunning, k.StatusDeploying, k.StatusFailed},
		{k.Sort, k.Filter, k.Enter},
	}
}

// Keys is the default set of serve-specific keybindings.
var Keys = KeyMap{
	StatusAll: key.NewBinding(
		key.WithKeys("a"),
		key.WithHelp("a", "all"),
	),
	StatusRunning: key.NewBinding(
		key.WithKeys("r"),
		key.WithHelp("r", "running"),
	),
	StatusDeploying: key.NewBinding(
		key.WithKeys("d"),
		key.WithHelp("d", "deploying"),
	),
	StatusFailed: key.NewBinding(
		key.WithKeys("f"),
		key.WithHelp("f", "failed"),
	),
	Sort: key.NewBinding(
		key.WithKeys("s"),
		key.WithHelp("s", "sort"),
	),
	Filter: key.NewBinding(
		key.WithKeys("/"),
		key.WithHelp("/", "filter"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "open"),
	),
}

// ApplyBindings applies user keybinding overrides.
func ApplyBindings(b config.ServeBindings) {
	config.OverrideBinding(&Keys.StatusAll, b.All)
	config.OverrideBinding(&Keys.StatusRunning, b.Running)
	config.OverrideBinding(&Keys.StatusDeploying, b.Deploying)
	config.OverrideBinding(&Keys.StatusFailed, b.Failed)
	config.OverrideBinding(&Keys.Sort, b.Sort)
	config.OverrideBinding(&Keys.Filter, b.Filter)
	config.OverrideBinding(&Keys.Enter, b.Open)
}
