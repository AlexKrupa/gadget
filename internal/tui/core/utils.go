package core

import (
	"os"
	"strings"
)

// ShortenHomePath replaces home directory with ~ if applicable
func ShortenHomePath(path string) string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return path
	}

	if strings.HasPrefix(path, homeDir) {
		return strings.Replace(path, homeDir, "~", 1)
	}

	return path
}
