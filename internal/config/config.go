package config

import (
	"os"
	"path/filepath"
)

// Config holds the application configuration
type Config struct {
	AndroidHome     string
	MediaPath       string
	ADBStaticPort   int
}

// NewConfig creates a new configuration with default values
func NewConfig() *Config {
	// Get Android SDK path from environment or default location
	androidHome := os.Getenv("ANDROID_HOME")
	if androidHome == "" {
		androidHome = os.Getenv("ANDROID_SDK_ROOT")
	}
	if androidHome == "" {
		// Default macOS location
		home, _ := os.UserHomeDir()
		androidHome = filepath.Join(home, "Library", "Android", "sdk")
	}

	// Default download path for media files
	home, _ := os.UserHomeDir()
	mediaPath := filepath.Join(home, "Downloads")

	return &Config{
		AndroidHome:   androidHome,
		MediaPath:     mediaPath,
		ADBStaticPort: 4444,
	}
}

// GetADBPath returns the path to adb executable
func (c *Config) GetADBPath() string {
	return filepath.Join(c.AndroidHome, "platform-tools", "adb")
}

// GetEmulatorPath returns the path to emulator executable
func (c *Config) GetEmulatorPath() string {
	return filepath.Join(c.AndroidHome, "emulator", "emulator")
}

// GetAVDManagerPath returns the path to avdmanager executable
func (c *Config) GetAVDManagerPath() string {
	return filepath.Join(c.AndroidHome, "cmdline-tools", "latest", "bin", "avdmanager")
}