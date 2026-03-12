package actordetail

import "charm.land/bubbles/v2/key"

// KeyMap defines keybindings specific to the actor detail view.
type KeyMap struct {
	Tab       key.Binding
	Refresh   key.Binding
	Search    key.Binding
	NextMatch key.Binding
	PrevMatch key.Binding
	GoToJob   key.Binding
	GoToNode  key.Binding
	PageDown  key.Binding
	PageUp    key.Binding
	Top       key.Binding
	Bottom    key.Binding
}

// ShortHelp returns keybindings for the short help view.
func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Tab, k.Refresh, k.GoToJob, k.GoToNode, k.Search, k.PageDown, k.Top}
}

// FullHelp returns keybindings grouped for the full help view.
func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Tab, k.Refresh, k.GoToJob, k.GoToNode},
		{k.Search, k.NextMatch, k.PrevMatch},
		{k.PageDown, k.PageUp, k.Top, k.Bottom},
	}
}

// Keys is the default set of actor detail keybindings.
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
	GoToJob: key.NewBinding(
		key.WithKeys("J"),
		key.WithHelp("J", "go to job"),
	),
	GoToNode: key.NewBinding(
		key.WithKeys("O"),
		key.WithHelp("O", "go to node"),
	),
	PageDown: key.NewBinding(
		key.WithKeys("space", "f", "pgdown"),
		key.WithHelp("space/f", "page down"),
	),
	PageUp: key.NewBinding(
		key.WithKeys("b", "pgup"),
		key.WithHelp("b", "page up"),
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
