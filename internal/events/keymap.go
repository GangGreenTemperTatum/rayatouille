package events

import (
	"charm.land/bubbles/v2/key"

	"github.com/GangGreenTemperTatum/rayatouille/internal/config"
)

// KeyMap defines keybindings specific to the events view.
type KeyMap struct {
	SeverityAll     key.Binding
	SeverityError   key.Binding
	SeverityWarning key.Binding
	SeverityInfo    key.Binding
	Sort            key.Binding
	Filter          key.Binding
}

// ShortHelp returns keybindings for the short help view.
func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.SeverityAll, k.SeverityError, k.SeverityWarning, k.SeverityInfo, k.Sort, k.Filter}
}

// FullHelp returns keybindings grouped for the full help view.
func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.SeverityAll, k.SeverityError, k.SeverityWarning, k.SeverityInfo},
		{k.Sort, k.Filter},
	}
}

// Keys is the default set of events keybindings.
var Keys = KeyMap{
	SeverityAll: key.NewBinding(
		key.WithKeys("a"),
		key.WithHelp("a", "all"),
	),
	SeverityError: key.NewBinding(
		key.WithKeys("e"),
		key.WithHelp("e", "error"),
	),
	SeverityWarning: key.NewBinding(
		key.WithKeys("w"),
		key.WithHelp("w", "warning"),
	),
	SeverityInfo: key.NewBinding(
		key.WithKeys("i"),
		key.WithHelp("i", "info"),
	),
	Sort: key.NewBinding(
		key.WithKeys("s"),
		key.WithHelp("s", "sort"),
	),
	Filter: key.NewBinding(
		key.WithKeys("/"),
		key.WithHelp("/", "filter"),
	),
}

// ApplyBindings applies user keybinding overrides.
func ApplyBindings(b config.EventsBindings) {
	config.OverrideBinding(&Keys.SeverityAll, b.All)
	config.OverrideBinding(&Keys.SeverityError, b.Error)
	config.OverrideBinding(&Keys.SeverityWarning, b.Warning)
	config.OverrideBinding(&Keys.SeverityInfo, b.Info)
	config.OverrideBinding(&Keys.Sort, b.Sort)
	config.OverrideBinding(&Keys.Filter, b.Filter)
}
