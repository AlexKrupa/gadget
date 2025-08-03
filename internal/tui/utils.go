package tui

import (
	"gadget/internal/tui/core"
)

// Delegate to core functions
func shortenHomePath(path string) string {
	return core.ShortenHomePath(path)
}

func formatErrorMessage(operation, deviceSerial string, err error) string {
	return core.FormatErrorMessage(operation, deviceSerial, err)
}

func formatSuccessMessage(operation, deviceSerial, details string) string {
	return core.FormatSuccessMessage(operation, deviceSerial, details)
}
