package tui

import (
	"github.com/charmbracelet/bubbles/key"
)

// KeyMap defines all key bindings for different modes
type KeyMap struct {
	// Global keys
	Quit key.Binding

	// Menu mode keys
	Up       key.Binding
	Down     key.Binding
	Enter    key.Binding
	Search   key.Binding
	Escape   key.Binding
	Backspace key.Binding

	// Device/Emulator selection keys
	VimUp    key.Binding
	VimDown  key.Binding
	VimLeft  key.Binding
	VimRight key.Binding

	// Text input keys (handled by bubbles TextInput component later)
	Submit key.Binding
	Cancel key.Binding

	// Recording keys
	StopRecording key.Binding

	// Context-specific escape keys
	EscapeBack key.Binding // For going back
}

// DefaultKeyMap returns the default key bindings
func DefaultKeyMap() KeyMap {
	return KeyMap{
		// Global
		Quit: key.NewBinding(
			key.WithKeys("ctrl+c"),
			key.WithHelp("ctrl+c", "quit"),
		),

		// Menu navigation (with vim-style keys)
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "down"),
		),
		Enter: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "select"),
		),
		Search: key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", "search"),
		),
		Escape: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "back/exit search"),
		),
		EscapeBack: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "back"),
		),
		Backspace: key.NewBinding(
			key.WithKeys("backspace"),
			key.WithHelp("backspace", "clear filter"),
		),

		// Device selection (with vim-style navigation)
		VimUp: key.NewBinding(
			key.WithKeys("up", "k", "h"),
			key.WithHelp("↑/k/h", "up"),
		),
		VimDown: key.NewBinding(
			key.WithKeys("down", "j", "l"),
			key.WithHelp("↓/j/l", "down"),
		),
		VimLeft: key.NewBinding(
			key.WithKeys("h"),
			key.WithHelp("h", "left"),
		),
		VimRight: key.NewBinding(
			key.WithKeys("l"),
			key.WithHelp("l", "right"),
		),

		// Text input
		Submit: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "submit"),
		),
		Cancel: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "cancel"),
		),

		// Recording
		StopRecording: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "stop recording"),
		),
	}
}

// MenuKeys returns keys available in menu mode
func (k KeyMap) MenuKeys(searchMode bool) []key.Binding {
	if searchMode {
		return []key.Binding{k.Up, k.Down, k.Enter, k.Escape, k.Backspace, k.Quit}
	}
	return []key.Binding{k.Search, k.Up, k.Down, k.Enter, k.Quit}
}

// SelectionKeys returns keys available in selection modes (device/emulator)
func (k KeyMap) SelectionKeys() []key.Binding {
	return []key.Binding{k.VimUp, k.VimDown, k.Enter, k.EscapeBack, k.Quit}
}

// DeviceSelectKeys returns keys available in device selection mode
func (k KeyMap) DeviceSelectKeys() []key.Binding {
	return k.SelectionKeys()
}

// EmulatorSelectKeys returns keys available in emulator selection mode
func (k KeyMap) EmulatorSelectKeys() []key.Binding {
	return k.SelectionKeys()
}

// TextInputKeys returns keys available in text input mode
func (k KeyMap) TextInputKeys() []key.Binding {
	return []key.Binding{k.Submit, k.Cancel, k.Quit}
}

// RecordingKeys returns keys available during recording
func (k KeyMap) RecordingKeys() []key.Binding {
	return []key.Binding{k.StopRecording, k.Quit}
}