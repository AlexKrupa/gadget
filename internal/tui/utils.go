package tui

import (
	"fmt"
	"os"
	"strings"
)

// shortenHomePath replaces home directory with ~ if applicable
func shortenHomePath(path string) string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return path
	}

	if strings.HasPrefix(path, homeDir) {
		return strings.Replace(path, homeDir, "~", 1)
	}

	return path
}

// formatErrorMessage creates a standardized error message format
func formatErrorMessage(operation, deviceSerial string, err error) string {
	return fmt.Sprintf("%s failed on %s: %s", operation, deviceSerial, err.Error())
}

// formatSuccessMessage creates a standardized success message format
func formatSuccessMessage(operation, deviceSerial, details string) string {
	if details != "" {
		return fmt.Sprintf("%s completed on %s\n%s", operation, deviceSerial, details)
	}
	return fmt.Sprintf("%s completed on %s", operation, deviceSerial)
}