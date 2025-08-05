package media

import (
	"fmt"
	"gadget/internal/adb"
	"gadget/internal/commands"
	"gadget/internal/config"
	"gadget/internal/tui/core"
	"gadget/internal/tui/messaging"
	"path/filepath"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// TakeScreenshotCmd returns a command to take a single screenshot
func TakeScreenshotCmd(cfg *config.Config, device adb.Device) tea.Cmd {
	return executeScreenshotOperation(cfg, device, ScreenshotSingle)
}

// TakeDayNightScreenshotsCmd returns a command to take day-night screenshots
func TakeDayNightScreenshotsCmd(cfg *config.Config, device adb.Device) tea.Cmd {
	return executeScreenshotOperation(cfg, device, ScreenshotDayNight)
}

// StartScreenRecordCmd returns a command to start screen recording
func StartScreenRecordCmd(cfg *config.Config, device adb.Device) tea.Cmd {
	return messaging.StartScreenRecordCmd(cfg, device)
}

// StopAndSaveRecordingCmd returns a command to stop and save screen recording
func StopAndSaveRecordingCmd(recording *commands.ScreenRecording) tea.Cmd {
	return func() tea.Msg {
		err := recording.StopAndSave()
		if err != nil {
			return messaging.ScreenRecordDoneMsg{
				Success: false,
				Message: err.Error(),
			}
		}
		return messaging.ScreenRecordDoneMsg{
			Success: true,
			Message: fmt.Sprintf("Screen recording saved on %s\n%s", recording.Device.Serial, core.ShortenHomePath(recording.LocalPath)),
		}
	}
}

// executeScreenshotOperation executes a screenshot operation asynchronously with common handling
func executeScreenshotOperation(cfg *config.Config, device adb.Device, operation ScreenshotOperation) tea.Cmd {
	return func() tea.Msg {
		timestamp := time.Now().Format("2006-01-02_15-04-05")

		switch operation {
		case ScreenshotSingle:
			err := commands.TakeScreenshot(cfg, device)
			if err != nil {
				return messaging.ScreenshotDoneMsg{
					Success: false,
					Message: fmt.Sprintf("Screenshot failed on %s: %s", device.Serial, err.Error()),
				}
			}

			filename := fmt.Sprintf("android-img-%s.png", timestamp)
			localPath := filepath.Join(cfg.MediaPath, filename)
			message := fmt.Sprintf("Screenshot captured on %s\n%s", device.Serial, core.ShortenHomePath(localPath))
			return messaging.ScreenshotDoneMsg{Success: true, Message: message}

		case ScreenshotDayNight:
			err := commands.TakeDayNightScreenshots(cfg, device)
			if err != nil {
				return messaging.DayNightScreenshotDoneMsg{Success: false, Message: err.Error()}
			}

			filenameDay := fmt.Sprintf("android-img-%s-day.png", timestamp)
			filenameNight := fmt.Sprintf("android-img-%s-night.png", timestamp)
			localPathDay := filepath.Join(cfg.MediaPath, filenameDay)
			localPathNight := filepath.Join(cfg.MediaPath, filenameNight)

			message := fmt.Sprintf("Day-night screenshots captured on %s\nDay:   %s\nNight: %s",
				device.Serial, core.ShortenHomePath(localPathDay), core.ShortenHomePath(localPathNight))
			return messaging.DayNightScreenshotDoneMsg{Success: true, Message: message}
		}

		return nil // Should never reach here
	}
}
