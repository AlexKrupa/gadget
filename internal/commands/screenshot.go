package commands

import (
	"fmt"
	"gadget/internal/adb"
	"gadget/internal/config"
	"gadget/internal/logger"
	"path/filepath"
	"time"
)

func TakeScreenshot(cfg *config.Config, device adb.Device) error {
	return takeScreenshot(cfg, device, "")
}

func takeScreenshot(cfg *config.Config, device adb.Device, suffix string) error {
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	var filename string
	if suffix == "" {
		filename = fmt.Sprintf("android-img-%s.png", timestamp)
	} else {
		filename = fmt.Sprintf("android-img-%s-%s.png", timestamp, suffix)
	}
	localPath := filepath.Join(cfg.MediaPath, filename)
	remotePath := "/sdcard/screenshot.png"
	adbPath := cfg.GetADBPath()

	err := adb.ExecuteCommand(adbPath, device.Serial, "shell", "screencap", remotePath)
	if err != nil {
		return fmt.Errorf("failed to take screenshot: %w", err)
	}

	err = adb.ExecuteCommand(adbPath, device.Serial, "pull", remotePath, localPath)
	if err != nil {
		return fmt.Errorf("failed to pull screenshot: %w", err)
	}

	adb.ExecuteCommand(adbPath, device.Serial, "shell", "rm", remotePath)

	logger.Success("Screenshot saved to: %s", localPath)
	return nil
}

func SetDarkMode(cfg *config.Config, device adb.Device, enabled bool) error {
	adbPath := cfg.GetADBPath()
	mode := "no"
	if enabled {
		mode = "yes"
	}

	return adb.ExecuteCommand(adbPath, device.Serial, "shell", "cmd", "uimode", "night", mode)
}

func TakeDayNightScreenshots(cfg *config.Config, device adb.Device) error {
	return takeDayNightScreenshots(cfg, device)
}

func takeDayNightScreenshots(cfg *config.Config, device adb.Device) error {
	logger.Info("Taking day and night screenshots of %s", device.Serial)

	logger.Info("Setting light mode...")
	err := SetDarkMode(cfg, device, false)
	if err != nil {
		return fmt.Errorf("failed to set light mode: %w", err)
	}

	time.Sleep(2 * time.Second) // Wait for UI to update

	logger.Info("Taking day screenshot...")
	err = takeScreenshot(cfg, device, "day")
	if err != nil {
		return fmt.Errorf("failed to take day screenshot: %w", err)
	}

	logger.Info("Setting dark mode...")
	err = SetDarkMode(cfg, device, true)
	if err != nil {
		return fmt.Errorf("failed to set dark mode: %w", err)
	}

	time.Sleep(2 * time.Second) // Wait for UI to update

	logger.Info("Taking night screenshot...")
	err = takeScreenshot(cfg, device, "night")
	if err != nil {
		return fmt.Errorf("failed to take night screenshot: %w", err)
	}

	logger.Info("Restoring light mode...")
	time.Sleep(2 * time.Second)
	err = SetDarkMode(cfg, device, false)
	if err != nil {
		logger.Error("Warning: failed to restore light mode: %v", err)
	}

	return nil
}

// CleanupRemoteFile removes a file from the device
func CleanupRemoteFile(adbPath, serial, remotePath string) {
	adb.ExecuteCommand(adbPath, serial, "shell", "rm", remotePath)
}
