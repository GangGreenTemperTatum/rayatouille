package ui

import (
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
)

// FilterModel is a reusable filter component wrapping textinput.Model.
// Any list view can compose it, call Activate() on "/", and use Matches()
// to filter rows by case-insensitive substring match.
type FilterModel struct {
	input  textinput.Model
	active bool
	value  string // last committed filter value
}

// NewFilter creates a new FilterModel with default settings.
func NewFilter() FilterModel {
	ti := textinput.New()
	ti.Prompt = "/ "
	ti.Placeholder = "filter..."
	ti.SetWidth(30)
	return FilterModel{
		input: ti,
	}
}

// Activate enables the filter and focuses the text input.
func (f *FilterModel) Activate() tea.Cmd {
	f.active = true
	return f.input.Focus()
}

// Deactivate disables the filter and blurs the text input.
func (f *FilterModel) Deactivate() {
	f.active = false
	f.input.Blur()
}

// Clear resets the filter value and deactivates.
func (f *FilterModel) Clear() {
	f.value = ""
	f.input.SetValue("")
	f.Deactivate()
}

// Active returns whether the filter is currently active.
func (f *FilterModel) Active() bool {
	return f.active
}

// Value returns the last committed filter value.
func (f *FilterModel) Value() string {
	return f.value
}

// SetValueForTest sets the committed filter value directly.
// Intended for use in tests and by composing views that need to set the filter programmatically.
func (f *FilterModel) SetValueForTest(v string) {
	f.value = v
	f.input.SetValue(v)
}

// Matches returns true if s contains the filter value (case-insensitive).
// If the filter value is empty, always returns true.
func (f *FilterModel) Matches(s string) bool {
	if f.value == "" {
		return true
	}
	return strings.Contains(
		strings.ToLower(s),
		strings.ToLower(f.value),
	)
}

// Update handles messages for the filter. If not active, returns immediately.
func (f FilterModel) Update(msg tea.Msg) (FilterModel, tea.Cmd) {
	if !f.active {
		return f, nil
	}

	if msg, ok := msg.(tea.KeyPressMsg); ok {
		switch msg.String() {
		case "enter":
			f.value = f.input.Value()
			f.Deactivate()
			return f, nil
		case "esc":
			f.input.SetValue(f.value)
			f.Deactivate()
			return f, nil
		}
	}

	var cmd tea.Cmd
	f.input, cmd = f.input.Update(msg)
	return f, cmd
}

// View renders the filter input. Returns empty string if not active and no value set.
func (f FilterModel) View() string {
	if !f.active && f.value == "" {
		return ""
	}
	return f.input.View()
}
