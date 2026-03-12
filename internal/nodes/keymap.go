package nodes

import (
	"charm.land/bubbles/v2/key"

	"github.com/GangGreenTemperTatum/rayatouille/internal/config"
)

// KeyMap defines keybindings specific to the nodes list view.
type KeyMap struct {
	StatusAll   key.Binding
	StatusAlive key.Binding
	StatusDead  key.Binding
	Sort        key.Binding
	Filter      key.Binding
	Enter       key.Binding
}

// ShortHelp returns keybindings for the short help view.
func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.StatusAll, k.StatusAlive, k.StatusDead, k.Sort, k.Filter, k.Enter}
}

// FullHelp returns keybindings grouped for the full help view.
func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.StatusAll, k.StatusAlive, k.StatusDead},
		{k.Sort, k.Filter, k.Enter},
	}
}

// Keys is the default set of nodes-specific keybindings.
var Keys = KeyMap{
	StatusAll: key.NewBinding(
		key.WithKeys("a"),
		key.WithHelp("a", "all"),
	),
	StatusAlive: key.NewBinding(
		key.WithKeys("l"),
		key.WithHelp("l", "alive"),
	),
	StatusDead: key.NewBinding(
		key.WithKeys("d"),
		key.WithHelp("d", "dead"),
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
func ApplyBindings(b config.NodesBindings) {
	config.OverrideBinding(&Keys.StatusAll, b.All)
	config.OverrideBinding(&Keys.StatusAlive, b.Alive)
	config.OverrideBinding(&Keys.StatusDead, b.Dead)
	config.OverrideBinding(&Keys.Sort, b.Sort)
	config.OverrideBinding(&Keys.Filter, b.Filter)
	config.OverrideBinding(&Keys.Enter, b.Open)
}
