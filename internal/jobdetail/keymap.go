package jobdetail

import (
	"charm.land/bubbles/v2/key"

	"github.com/GangGreenTemperTatum/rayatouille/internal/config"
)

// KeyMap defines keybindings specific to the job detail view.
type KeyMap struct {
	Tab       key.Binding
	Refresh   key.Binding
	Search    key.Binding
	NextMatch key.Binding
	PrevMatch key.Binding
	PageDown  key.Binding
	PageUp    key.Binding
	HalfDown  key.Binding
	HalfUp    key.Binding
	Top       key.Binding
	Bottom    key.Binding
}

// ShortHelp returns keybindings for the short help view.
func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Tab, k.Refresh, k.Search, k.PageDown, k.Top, k.Bottom}
}

// FullHelp returns keybindings grouped for the full help view.
func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Tab, k.Refresh},
		{k.Search, k.NextMatch, k.PrevMatch},
		{k.PageDown, k.PageUp, k.HalfDown, k.HalfUp},
		{k.Top, k.Bottom},
	}
}

// Keys is the default set of job detail keybindings.
var Keys = KeyMap{
	Tab: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "switch section"),
	),
	Refresh: key.NewBinding(
		key.WithKeys("r"),
		key.WithHelp("r", "refresh"),
	),
	Search: key.NewBinding(
		key.WithKeys("/"),
		key.WithHelp("/", "search logs"),
	),
	NextMatch: key.NewBinding(
		key.WithKeys("n"),
		key.WithHelp("n", "next match"),
	),
	PrevMatch: key.NewBinding(
		key.WithKeys("N"),
		key.WithHelp("N", "prev match"),
	),
	PageDown: key.NewBinding(
		key.WithKeys("space", "f", "pgdown"),
		key.WithHelp("space/f", "page down"),
	),
	PageUp: key.NewBinding(
		key.WithKeys("b", "pgup"),
		key.WithHelp("b", "page up"),
	),
	HalfDown: key.NewBinding(
		key.WithKeys("d", "ctrl+d"),
		key.WithHelp("d", "½ page down"),
	),
	HalfUp: key.NewBinding(
		key.WithKeys("u", "ctrl+u"),
		key.WithHelp("u", "½ page up"),
	),
	Top: key.NewBinding(
		key.WithKeys("g"),
		key.WithHelp("g", "top"),
	),
	Bottom: key.NewBinding(
		key.WithKeys("G"),
		key.WithHelp("G", "bottom"),
	),
}

// ApplyBindings applies user keybinding overrides for detail and logging keys.
func ApplyBindings(detail config.DetailBindings, logging config.LoggingBindings) {
	config.OverrideBinding(&Keys.Tab, detail.Tab)
	config.OverrideBinding(&Keys.Refresh, detail.Refresh)
	config.OverrideBinding(&Keys.Search, logging.Search)
	config.OverrideBinding(&Keys.NextMatch, logging.NextMatch)
	config.OverrideBinding(&Keys.PrevMatch, logging.PrevMatch)
	config.OverrideBinding(&Keys.Top, logging.Top)
	config.OverrideBinding(&Keys.Bottom, logging.Bottom)
}
