package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// KeybindingsConfig defines user-customizable keybindings.
// Keys are specified as strings matching Bubble Tea key names
// (e.g., "q", "ctrl+c", "esc", "tab", "enter", "space").
type KeybindingsConfig struct {
	Global  GlobalBindings  `json:"global,omitempty"`
	Jobs    JobsBindings    `json:"jobs,omitempty"`
	Nodes   NodesBindings   `json:"nodes,omitempty"`
	Actors  ActorsBindings  `json:"actors,omitempty"`
	Serve   ServeBindings   `json:"serve,omitempty"`
	Events  EventsBindings  `json:"events,omitempty"`
	Detail  DetailBindings  `json:"detail,omitempty"`
	Logging LoggingBindings `json:"logging,omitempty"`
}

// GlobalBindings defines overridable global keybindings.
type GlobalBindings struct {
	Quit   []string `json:"quit,omitempty"`
	Help   []string `json:"help,omitempty"`
	Back   []string `json:"back,omitempty"`
	Search []string `json:"search,omitempty"`
	Tab    []string `json:"tab,omitempty"`
}

// JobsBindings defines overridable jobs view keybindings.
type JobsBindings struct {
	All       []string `json:"all,omitempty"`
	Running   []string `json:"running,omitempty"`
	Failed    []string `json:"failed,omitempty"`
	Pending   []string `json:"pending,omitempty"`
	Completed []string `json:"completed,omitempty"`
	Sort      []string `json:"sort,omitempty"`
	Filter    []string `json:"filter,omitempty"`
	Open      []string `json:"open,omitempty"`
}

// NodesBindings defines overridable nodes view keybindings.
type NodesBindings struct {
	All    []string `json:"all,omitempty"`
	Alive  []string `json:"alive,omitempty"`
	Dead   []string `json:"dead,omitempty"`
	Sort   []string `json:"sort,omitempty"`
	Filter []string `json:"filter,omitempty"`
	Open   []string `json:"open,omitempty"`
}

// ActorsBindings defines overridable actors view keybindings.
type ActorsBindings struct {
	All     []string `json:"all,omitempty"`
	Alive   []string `json:"alive,omitempty"`
	Dead    []string `json:"dead,omitempty"`
	Pending []string `json:"pending,omitempty"`
	Sort    []string `json:"sort,omitempty"`
	Filter  []string `json:"filter,omitempty"`
	Open    []string `json:"open,omitempty"`
}

// ServeBindings defines overridable serve view keybindings.
type ServeBindings struct {
	All       []string `json:"all,omitempty"`
	Running   []string `json:"running,omitempty"`
	Deploying []string `json:"deploying,omitempty"`
	Failed    []string `json:"failed,omitempty"`
	Sort      []string `json:"sort,omitempty"`
	Filter    []string `json:"filter,omitempty"`
	Open      []string `json:"open,omitempty"`
}

// EventsBindings defines overridable events view keybindings.
type EventsBindings struct {
	All     []string `json:"all,omitempty"`
	Error   []string `json:"error,omitempty"`
	Warning []string `json:"warning,omitempty"`
	Info    []string `json:"info,omitempty"`
	Sort    []string `json:"sort,omitempty"`
	Filter  []string `json:"filter,omitempty"`
}

// DetailBindings defines overridable detail view keybindings.
type DetailBindings struct {
	Tab     []string `json:"tab,omitempty"`
	Refresh []string `json:"refresh,omitempty"`
}

// LoggingBindings defines overridable log viewer keybindings.
type LoggingBindings struct {
	Search    []string `json:"search,omitempty"`
	NextMatch []string `json:"next_match,omitempty"`
	PrevMatch []string `json:"prev_match,omitempty"`
	CopyLine  []string `json:"copy_line,omitempty"`
	CopyPage  []string `json:"copy_page,omitempty"`
	Top       []string `json:"top,omitempty"`
	Bottom    []string `json:"bottom,omitempty"`
}

// LoadKeybindings reads the keybindings config from ~/.config/rayatouille/keybindings.json.
// Returns an empty config (no overrides) if the file doesn't exist.
func LoadKeybindings() KeybindingsConfig {
	dir, err := ConfigDir()
	if err != nil {
		return KeybindingsConfig{}
	}
	path := filepath.Join(dir, "keybindings.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return KeybindingsConfig{}
	}
	var kb KeybindingsConfig
	if err := json.Unmarshal(data, &kb); err != nil {
		return KeybindingsConfig{}
	}
	return kb
}
