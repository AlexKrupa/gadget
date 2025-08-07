package logger

import (
	"time"
)

// TUILogEntry represents a log entry for the TUI
type TUILogEntry struct {
	Message   string
	Level     LogLevel
	Timestamp time.Time
}

// TUIRenderer renders log entries to a TUI log box
type TUIRenderer struct {
	logChannel chan<- TUILogEntry
}

// NewTUIRenderer creates a new TUI renderer that sends log entries to a channel
func NewTUIRenderer(logChannel chan<- TUILogEntry) *TUIRenderer {
	return &TUIRenderer{
		logChannel: logChannel,
	}
}

// Render sends the log entry to the TUI log channel
func (r *TUIRenderer) Render(entry LogEntry) {
	if r.logChannel == nil {
		return
	}

	tuiEntry := TUILogEntry{
		Message:   entry.Message,
		Level:     entry.Level,
		Timestamp: entry.Timestamp,
	}

	// Send to log channel (non-blocking)
	select {
	case r.logChannel <- tuiEntry:
	default:
		// Channel is full, drop the message to avoid blocking
	}
}
