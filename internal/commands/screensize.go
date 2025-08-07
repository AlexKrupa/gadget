package commands

import (
	"fmt"
	"gadget/internal/adb"
	"gadget/internal/config"
	"strconv"
	"strings"
)

type ScreenSizeInfo struct {
	Physical string
	Current  string // The effective screen size (override if exists, otherwise physical)
}

func GetCurrentScreenSize(cfg *config.Config, device adb.Device) (*ScreenSizeInfo, error) {
	adbPath := cfg.GetADBPath()
	output, err := adb.ExecuteCommandWithOutput(adbPath, device.Serial, "shell", "wm", "size")
	if err != nil {
		return nil, fmt.Errorf("failed to get current screen size: %w", err)
	}

	// Parse output which can contain:
	// Physical size: 1080x1920
	// Override size: 1080x1800
	lines := strings.Split(strings.TrimSpace(output), "\n")

	info := &ScreenSizeInfo{}

	for _, line := range lines {
		if strings.Contains(line, "Override size:") {
			parts := strings.Fields(line)
			if len(parts) >= 3 {
				info.Current = parts[2]
			}
		} else if strings.Contains(line, "Physical size:") {
			parts := strings.Fields(line)
			if len(parts) >= 3 {
				info.Physical = parts[2]
			}
		}
	}

	// Set current to override if it exists, otherwise physical
	if info.Current == "" {
		info.Current = info.Physical
	}

	if info.Physical == "" {
		return nil, fmt.Errorf("could not parse screen size from output: %s", output)
	}

	return info, nil
}

func SetScreenSize(cfg *config.Config, device adb.Device, size string) error {
	// Validate format (should be like "1080x1920")
	parts := strings.Split(size, "x")
	if len(parts) != 2 {
		return fmt.Errorf("invalid screen size format: %s (expected format: 1080x1920)", size)
	}

	// Validate that both parts are numbers
	for _, part := range parts {
		if _, err := strconv.Atoi(part); err != nil {
			return fmt.Errorf("invalid screen size format: %s (both width and height must be numbers)", size)
		}
	}

	adbPath := cfg.GetADBPath()
	err := adb.ExecuteCommand(adbPath, device.Serial, "shell", "wm", "size", size)
	if err != nil {
		return fmt.Errorf("failed to set screen size to %s: %w", size, err)
	}

	fmt.Printf("Screen size changed to %s on device %s\n", size, device.Serial)
	return nil
}
