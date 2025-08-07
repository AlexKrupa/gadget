package logger

import (
	"fmt"
	"os"
)

// CLIRenderer renders log entries to the terminal with colors
type CLIRenderer struct{}

// ANSI color codes
const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorGray   = "\033[90m"
)

// NewCLIRenderer creates a new CLI renderer
func NewCLIRenderer() *CLIRenderer {
	return &CLIRenderer{}
}

// Render outputs the log entry to the terminal with appropriate colors
func (r *CLIRenderer) Render(entry LogEntry) {
	var color string
	switch entry.Level {
	case LogLevelError:
		color = ColorRed
	case LogLevelSuccess:
		color = ColorGreen
	case LogLevelDebug:
		color = ColorGray
	default: // LogLevelInfo
		color = ColorReset
	}

	// Print colored message to stdout/stderr based on level
	output := fmt.Sprintf("%s%s%s", color, entry.Message, ColorReset)

	if entry.Level == LogLevelError {
		fmt.Fprint(os.Stderr, output)
		if output[len(output)-1] != '\n' {
			fmt.Fprint(os.Stderr, "\n")
		}
	} else {
		fmt.Print(output)
		if output[len(output)-1] != '\n' {
			fmt.Print("\n")
		}
	}
}
