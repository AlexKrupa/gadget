package commands

import (
	"fmt"
	"gadget/internal/adb"
	"gadget/internal/config"
	"strconv"
	"strings"
)

// FontSizeInfo holds font size information from the device
type FontSizeInfo struct {
	Default float64
	Current float64 // The effective font scale
}

// GetCurrentFontSize retrieves the current font size setting from the device
func GetCurrentFontSize(cfg *config.Config, device adb.Device) (*FontSizeInfo, error) {
	adbPath := cfg.GetADBPath()
	output, err := adb.ExecuteCommandWithOutput(adbPath, device.Serial, "shell", "settings", "get", "system", "font_scale")
	if err != nil {
		return nil, fmt.Errorf("failed to get current font size: %w", err)
	}

	currentStr := strings.TrimSpace(output)
	if currentStr == "null" || currentStr == "" {
		// Default font scale is 1.0 when not set
		return &FontSizeInfo{
			Default: 1.0,
			Current: 1.0,
		}, nil
	}

	current, err := strconv.ParseFloat(currentStr, 64)
	if err != nil {
		return nil, fmt.Errorf("could not parse font size from output: %s", output)
	}

	return &FontSizeInfo{
		Default: 1.0, // Android default is always 1.0
		Current: current,
	}, nil
}

// SetFontSize changes the device font size to the specified scale
func SetFontSize(cfg *config.Config, device adb.Device, scale float64) error {
	adbPath := cfg.GetADBPath()
	scaleStr := strconv.FormatFloat(scale, 'f', 1, 64)
	err := adb.ExecuteCommand(adbPath, device.Serial, "shell", "settings", "put", "system", "font_scale", scaleStr)
	if err != nil {
		return fmt.Errorf("failed to set font size to %s: %w", scaleStr, err)
	}

	fmt.Printf("Font size changed to %s on device %s\n", scaleStr, device.Serial)
	return nil
}
