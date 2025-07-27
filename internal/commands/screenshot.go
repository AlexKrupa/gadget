package commands

import (
	"adx/internal/adb"
	"adx/internal/config"
	"fmt"
	"path/filepath"
	"time"
)

// TakeScreenshot captures a screenshot from the specified device
func TakeScreenshot(cfg *config.Config, device adb.Device) error {
	// Generate timestamp filename (format: 2021-08-31_14-23-45)
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	filename := fmt.Sprintf("android-img-%s.png", timestamp)
	localPath := filepath.Join(cfg.MediaPath, filename)

	// Remote path on device
	remotePath := "/sdcard/screenshot.png"

	// Take screenshot on device
	adbPath := cfg.GetADBPath()
	err := adb.ExecuteCommand(adbPath, device.Serial, "shell", "screencap", remotePath)
	if err != nil {
		return fmt.Errorf("failed to take screenshot: %w", err)
	}

	// Pull screenshot to local machine
	err = adb.ExecuteCommand(adbPath, device.Serial, "pull", remotePath, localPath)
	if err != nil {
		return fmt.Errorf("failed to pull screenshot: %w", err)
	}

	// Clean up screenshot from device
	adb.ExecuteCommand(adbPath, device.Serial, "shell", "rm", remotePath)

	fmt.Printf("Screenshot saved to: %s\n", localPath)
	return nil
}

// setDarkMode toggles dark mode on the device using raw ADB commands
func setDarkMode(cfg *config.Config, device adb.Device, enabled bool) error {
	adbPath := cfg.GetADBPath()
	mode := "no"
	if enabled {
		mode = "yes"
	}

	return adb.ExecuteCommand(adbPath, device.Serial, "shell", "cmd", "uimode", "night", mode)
}

// TakeDayNightScreenshots captures screenshots in both day and night mode
func TakeDayNightScreenshots(cfg *config.Config, device adb.Device) error {
	// Generate timestamp for both files
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	filenameDay := fmt.Sprintf("android-img-%s-day.png", timestamp)
	filenameNight := fmt.Sprintf("android-img-%s-night.png", timestamp)
	localPathDay := filepath.Join(cfg.MediaPath, filenameDay)
	localPathNight := filepath.Join(cfg.MediaPath, filenameNight)

	// Remote path on device
	remotePath := "/sdcard/screenshot.png"
	adbPath := cfg.GetADBPath()

	fmt.Printf("Taking day and night screenshots of %s\n", device.Serial)

	// Take day screenshot
	fmt.Println("Setting light mode...")
	err := setDarkMode(cfg, device, false)
	if err != nil {
		return fmt.Errorf("failed to set light mode: %w", err)
	}

	time.Sleep(2 * time.Second) // Wait for UI to update

	fmt.Println("Taking day screenshot...")
	err = adb.ExecuteCommand(adbPath, device.Serial, "shell", "screencap", remotePath)
	if err != nil {
		return fmt.Errorf("failed to take day screenshot: %w", err)
	}

	err = adb.ExecuteCommand(adbPath, device.Serial, "pull", remotePath, localPathDay)
	if err != nil {
		return fmt.Errorf("failed to pull day screenshot: %w", err)
	}

	fmt.Printf("Day screenshot saved to: %s\n", localPathDay)

	// Take night screenshot
	fmt.Println("Setting dark mode...")
	err = setDarkMode(cfg, device, true)
	if err != nil {
		return fmt.Errorf("failed to set dark mode: %w", err)
	}

	time.Sleep(2 * time.Second) // Wait for UI to update

	fmt.Println("Taking night screenshot...")
	err = adb.ExecuteCommand(adbPath, device.Serial, "shell", "screencap", remotePath)
	if err != nil {
		return fmt.Errorf("failed to take night screenshot: %w", err)
	}

	err = adb.ExecuteCommand(adbPath, device.Serial, "pull", remotePath, localPathNight)
	if err != nil {
		return fmt.Errorf("failed to pull night screenshot: %w", err)
	}

	fmt.Printf("Night screenshot saved to: %s\n", localPathNight)

	// Restore light mode
	fmt.Println("Restoring light mode...")
	time.Sleep(2 * time.Second)
	err = setDarkMode(cfg, device, false)
	if err != nil {
		fmt.Printf("Warning: failed to restore light mode: %v\n", err)
	}

	// Clean up screenshot from device
	adb.ExecuteCommand(adbPath, device.Serial, "shell", "rm", remotePath)

	return nil
}
