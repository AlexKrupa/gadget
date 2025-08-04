package core

import (
	"time"
)

// Mode represents the current TUI mode
type Mode string

const (
	ModeMenu           Mode = "menu"
	ModeDeviceSelect   Mode = "device-select"
	ModeEmulatorSelect Mode = "emulator-select"
	ModeCommand        Mode = "command"
	ModeTextInput      Mode = "text-input"
)

// LogType represents the type of log message
type LogType int

const (
	LogTypeSuccess LogType = iota
	LogTypeError
	LogTypeInfo
)

// LogEntry represents a log message with metadata
type LogEntry struct {
	Message   string
	Type      LogType
	Timestamp time.Time
}

// Command represents a menu command
type Command struct {
	Command     string // kebab-case command name for CLI
	Name        string
	Description string
	Category    string
}

// CommandCategory represents a group of related commands
type CommandCategory struct {
	Name     string
	Commands []Command
}

// BaseModel contains core TUI state shared across features
type BaseModel struct {
	Mode         Mode
	Width        int
	Height       int
	SearchQuery  string
	Loading      bool
	CurrentIndex int
	Logs         []LogEntry
}

// TickMsg is sent for progress animation
type TickMsg time.Time
