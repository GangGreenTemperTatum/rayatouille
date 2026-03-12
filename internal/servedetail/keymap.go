package servedetail

import "charm.land/bubbles/v2/key"

// KeyMap defines keybindings specific to the serve detail view.
type KeyMap struct {
	Enter key.Binding
	Back  key.Binding
}

// ShortHelp returns keybindings for the short help view.
func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Enter, k.Back}
}

// FullHelp returns keybindings grouped for the full help view.
func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Enter, k.Back},
	}
}

// Keys is the default set of serve detail keybindings.
var Keys = KeyMap{
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "view replicas"),
	),
	Back: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "back to deployments"),
	),
}
