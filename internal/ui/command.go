package ui

import (
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// CommandResult represents the result of a parsed colon command.
type CommandResult struct {
	Command string
	Arg     string // optional argument (e.g., profile name)
}

// Known colon commands.
const (
	CmdJobs      = "jobs"
	CmdDashboard = "dashboard"
	CmdNodes     = "nodes"
	CmdActors    = "actors"
	CmdServe     = "serve"
	CmdEvents    = "events"
	CmdProfile   = "profile"
)

// paletteEntry describes a command in the palette.
type paletteEntry struct {
	name string
	desc string
	cmd  string
}

// paletteCommands is the ordered list of commands shown in the palette.
var paletteCommands = []paletteEntry{
	{"jobs", "View all jobs", CmdJobs},
	{"nodes", "View cluster nodes", CmdNodes},
	{"actors", "View actors", CmdActors},
	{"serve", "View Serve deployments", CmdServe},
	{"events", "View cluster events", CmdEvents},
	{"dashboard", "Back to overview", CmdDashboard},
	{"profile", "Switch cluster profile", CmdProfile},
}

// CommandModel is a command palette component with filtered list.
type CommandModel struct {
	input    textinput.Model
	active   bool
	matches  []paletteEntry
	cursor   int
	prevText string
}

// NewCommand creates a new CommandModel with default settings.
func NewCommand() CommandModel {
	ti := textinput.New()
	ti.Prompt = ": "
	ti.Placeholder = "type to filter..."
	ti.SetWidth(30)
	return CommandModel{
		input:   ti,
		matches: paletteCommands,
	}
}

// Activate enables the command palette and focuses it.
func (c *CommandModel) Activate() tea.Cmd {
	c.active = true
	c.cursor = 0
	c.matches = paletteCommands
	c.prevText = ""
	return c.input.Focus()
}

// Deactivate disables the command palette and clears it.
func (c *CommandModel) Deactivate() {
	c.active = false
	c.input.Blur()
	c.input.SetValue("")
	c.cursor = 0
	c.matches = paletteCommands
	c.prevText = ""
}

// Active returns whether the command palette is currently active.
func (c *CommandModel) Active() bool {
	return c.active
}

// Update handles messages for the command palette.
// The third return value is non-nil when a command is executed.
func (c CommandModel) Update(msg tea.Msg) (CommandModel, tea.Cmd, *CommandResult) {
	if !c.active {
		return c, nil, nil
	}

	if msg, ok := msg.(tea.KeyPressMsg); ok {
		switch msg.String() {
		case "enter":
			result := c.executeSelection()
			c.Deactivate()
			return c, nil, result
		case "esc":
			c.Deactivate()
			return c, nil, nil
		case "tab", "down":
			if len(c.matches) > 0 {
				c.cursor = (c.cursor + 1) % len(c.matches)
				c.fillFromCursor()
			}
			return c, nil, nil
		case "shift+tab", "up":
			if len(c.matches) > 0 {
				c.cursor = (c.cursor - 1 + len(c.matches)) % len(c.matches)
				c.fillFromCursor()
			}
			return c, nil, nil
		}
	}

	var cmd tea.Cmd
	c.input, cmd = c.input.Update(msg)

	// Re-filter when text changes.
	text := c.input.Value()
	if text != c.prevText {
		c.prevText = text
		c.filterMatches(text)
	}

	return c, cmd, nil
}

// fillFromCursor sets the input text to the currently highlighted command name.
func (c *CommandModel) fillFromCursor() {
	if c.cursor < len(c.matches) {
		name := c.matches[c.cursor].name
		c.input.SetValue(name)
		c.input.SetCursor(len(name))
		c.prevText = name
	}
}

// filterMatches updates the matches list based on the input text.
func (c *CommandModel) filterMatches(text string) {
	prefix := strings.ToLower(strings.TrimSpace(text))
	if prefix == "" {
		c.matches = paletteCommands
		c.cursor = 0
		return
	}

	var matches []paletteEntry
	for _, entry := range paletteCommands {
		if strings.Contains(entry.name, prefix) || strings.Contains(strings.ToLower(entry.desc), prefix) {
			matches = append(matches, entry)
		}
	}
	c.matches = matches
	if c.cursor >= len(c.matches) {
		c.cursor = 0
	}
}

// executeSelection returns the command for the currently selected palette entry.
func (c *CommandModel) executeSelection() *CommandResult {
	text := c.input.Value()

	// Sync filter in case text was set directly (e.g., in tests).
	c.filterMatches(text)

	// If no matches, fall back to direct parse.
	if len(c.matches) == 0 {
		return ParseCommand(text)
	}

	selected := c.matches[c.cursor]

	// For profile command, check if user typed an arg (e.g., "profile prod").
	if selected.cmd == CmdProfile {
		parts := strings.Fields(text)
		if len(parts) > 1 {
			return &CommandResult{Command: CmdProfile, Arg: strings.Join(parts[1:], " ")}
		}
		return &CommandResult{Command: CmdProfile}
	}

	return &CommandResult{Command: selected.cmd}
}

// View renders the command palette. Returns empty string if not active.
func (c CommandModel) View() string {
	if !c.active {
		return ""
	}

	selectedStyle := lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary)
	normalStyle := lipgloss.NewStyle().Foreground(ColorFg)
	descStyle := lipgloss.NewStyle().Foreground(ColorMuted)
	cursorStr := lipgloss.NewStyle().Foreground(ColorAccent).Render("> ")
	pad := "  "

	var lines []string
	for i, entry := range c.matches {
		name := normalStyle.Render(entry.name)
		desc := descStyle.Render("  " + entry.desc)
		prefix := pad
		if i == c.cursor {
			name = selectedStyle.Render(entry.name)
			prefix = cursorStr
		}
		lines = append(lines, prefix+name+desc)
	}

	list := strings.Join(lines, "\n")
	if len(c.matches) == 0 {
		list = descStyle.Render("  no matching commands")
	}

	return c.input.View() + "\n" + list
}

// ParseCommand parses a raw input string and returns a CommandResult if it
// matches a known command. Returns nil if no match.
func ParseCommand(input string) *CommandResult {
	trimmed := strings.TrimSpace(input)
	trimmed = strings.ToLower(trimmed)

	// Split on first space to separate command from argument.
	command := trimmed
	arg := ""
	if idx := strings.IndexByte(trimmed, ' '); idx >= 0 {
		command = trimmed[:idx]
		arg = strings.TrimSpace(trimmed[idx+1:])
	}

	switch command {
	case "jobs":
		return &CommandResult{Command: CmdJobs}
	case "dashboard", "dash":
		return &CommandResult{Command: CmdDashboard}
	case "nodes":
		return &CommandResult{Command: CmdNodes}
	case "actors":
		return &CommandResult{Command: CmdActors}
	case "serve":
		return &CommandResult{Command: CmdServe}
	case "events":
		return &CommandResult{Command: CmdEvents}
	case "profile":
		return &CommandResult{Command: CmdProfile, Arg: arg}
	default:
		return nil
	}
}
