package commands

import (
	"fmt"
	"gadget/internal/adb"
	"gadget/internal/config"
	"strconv"
	"strings"
)

type DPIInfo struct {
	Physical int
	Override int
	Current  int // The effective DPI (override if exists, otherwise physical)
}

func GetCurrentDPI(cfg *config.Config, device adb.Device) (*DPIInfo, error) {
	adbPath := cfg.GetADBPath()
	output, err := adb.ExecuteCommandWithOutput(adbPath, device.Serial, "shell", "wm", "density")
	if err != nil {
		return nil, fmt.Errorf("failed to get current DPI: %w", err)
	}

	// Parse output which can contain:
	// Physical density: 420
	// Override density: 480
	lines := strings.Split(strings.TrimSpace(output), "\n")

	info := &DPIInfo{}

	for _, line := range lines {
		if strings.Contains(line, "Override density:") {
			parts := strings.Fields(line)
			if len(parts) >= 3 {
				dpi, err := strconv.Atoi(parts[2])
				if err == nil {
					info.Override = dpi
				}
			}
		} else if strings.Contains(line, "Physical density:") {
			parts := strings.Fields(line)
			if len(parts) >= 3 {
				dpi, err := strconv.Atoi(parts[2])
				if err == nil {
					info.Physical = dpi
				}
			}
		}
	}

	// Set current to override if it exists, otherwise physical
	if info.Override > 0 {
		info.Current = info.Override
	} else {
		info.Current = info.Physical
	}

	if info.Physical == 0 {
		return nil, fmt.Errorf("could not parse DPI from output: %s", output)
	}

	return info, nil
}

func SetDPI(cfg *config.Config, device adb.Device, dpi int) error {
	adbPath := cfg.GetADBPath()
	err := adb.ExecuteCommand(adbPath, device.Serial, "shell", "wm", "density", strconv.Itoa(dpi))
	if err != nil {
		return fmt.Errorf("failed to set DPI to %d: %w", dpi, err)
	}

	fmt.Printf("DPI changed to %d on device %s\n", dpi, device.Serial)
	return nil
}
