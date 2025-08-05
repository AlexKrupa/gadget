package media

import (
	"fmt"
	"gadget/internal/adb"
	"gadget/internal/commands"
	"gadget/internal/config"
	"gadget/internal/tui/capture"
	"gadget/internal/tui/core"
	"gadget/internal/tui/messaging"
	"path/filepath"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// StreamingDayNightScreenshot represents a request to start streaming day-night screenshots
type StreamingDayNightScreenshot struct {
	Config    *config.Config
	Device    adb.Device
	Timestamp string
}

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
		// Changed: Capture command output
		capturedOutput, err := capture.CaptureCommand(func() error {
			return recording.StopAndSave()
		})

		if err != nil {
			return messaging.ScreenRecordDoneMsg{
				Success:        false,
				Message:        err.Error(),
				CapturedOutput: capturedOutput,
			}
		}
		return messaging.ScreenRecordDoneMsg{
			Success:        true,
			Message:        fmt.Sprintf("Screen recording saved on %s\n%s", recording.Device.Serial, core.ShortenHomePath(recording.LocalPath)),
			CapturedOutput: capturedOutput,
		}
	}
}

// executeScreenshotOperation executes a screenshot operation asynchronously with common handling
func executeScreenshotOperation(cfg *config.Config, device adb.Device, operation ScreenshotOperation) tea.Cmd {
	return func() tea.Msg {
		timestamp := time.Now().Format("2006-01-02_15-04-05")

		switch operation {
		case ScreenshotSingle:
			// Changed: Capture command output
			capturedOutput, err := capture.CaptureCommand(func() error {
				return commands.TakeScreenshot(cfg, device)
			})

			if err != nil {
				return messaging.ScreenshotDoneMsg{
					Success:        false,
					Message:        fmt.Sprintf("Screenshot failed on %s: %s", device.Serial, err.Error()),
					CapturedOutput: capturedOutput,
				}
			}

			filename := fmt.Sprintf("android-img-%s.png", timestamp)
			localPath := filepath.Join(cfg.MediaPath, filename)
			message := fmt.Sprintf("Screenshot captured on %s\n%s", device.Serial, core.ShortenHomePath(localPath))
			return messaging.ScreenshotDoneMsg{
				Success:        true,
				Message:        message,
				CapturedOutput: capturedOutput,
			}

		case ScreenshotDayNight:
			// Create a simple streaming command using tea.Batch
			return createStreamingDayNightCommand(cfg, device, timestamp)
		}

		return nil // Should never reach here
	}
}

// StreamingCommandStart signals the start of a streaming command
type StreamingCommandStart struct {
	OutputChan <-chan string
	Config     *config.Config
	Device     adb.Device
	Timestamp  string
}

// createStreamingDayNightCommand creates a command that shows progress as it happens
func createStreamingDayNightCommand(cfg *config.Config, device adb.Device, timestamp string) tea.Msg {
	outputChan := make(chan string, 100)

	// Start the command in a goroutine
	go func() {
		defer close(outputChan) // Only close outputChan when truly done

		// Use a helper function to send progress updates safely
		sendProgress := func(message string) {
			select {
			case outputChan <- message:
			default:
				// Channel full, skip to prevent blocking
			}
		}

		// Execute day-night screenshots with live progress
		err := executeDayNightWithProgress(cfg, device, timestamp, sendProgress)
		if err != nil {
			sendProgress(fmt.Sprintf("Command failed: %v", err))
		}
		// Channel closes here via defer, signaling completion
	}()

	return StreamingCommandStart{
		OutputChan: outputChan,
		Config:     cfg,
		Device:     device,
		Timestamp:  timestamp,
	}
}

// executeDayNightWithProgress executes day-night screenshots with progress callbacks
func executeDayNightWithProgress(cfg *config.Config, device adb.Device, timestamp string, progress func(string)) error {
	filenameDay := fmt.Sprintf("android-img-%s-day.png", timestamp)
	filenameNight := fmt.Sprintf("android-img-%s-night.png", timestamp)
	localPathDay := filepath.Join(cfg.MediaPath, filenameDay)
	localPathNight := filepath.Join(cfg.MediaPath, filenameNight)
	remotePath := "/sdcard/screenshot.png"
	adbPath := cfg.GetADBPath()

	progress(fmt.Sprintf("Taking day and night screenshots of %s", device.Serial))

	progress("Setting light mode...")
	err := commands.SetDarkModeForTUI(cfg, device, false)
	if err != nil {
		progress(fmt.Sprintf("Error setting light mode: %v", err))
		return err
	}
	time.Sleep(1 * time.Second) // Reduced from 2s - modern devices update faster

	progress("Taking day screenshot...")
	err = commands.TakeScreenshotForTUI(adbPath, device.Serial, remotePath, localPathDay)
	if err != nil {
		progress(fmt.Sprintf("Error taking day screenshot: %v", err))
		return err
	}
	progress(fmt.Sprintf("Day screenshot saved to: %s", localPathDay))

	progress("Setting dark mode...")
	err = commands.SetDarkModeForTUI(cfg, device, true)
	if err != nil {
		progress(fmt.Sprintf("Error setting dark mode: %v", err))
		return err
	}
	time.Sleep(1 * time.Second) // Reduced from 2s

	progress("Taking night screenshot...")
	err = commands.TakeScreenshotForTUI(adbPath, device.Serial, remotePath, localPathNight)
	if err != nil {
		progress(fmt.Sprintf("Error taking night screenshot: %v", err))
		return err
	}
	progress(fmt.Sprintf("Night screenshot saved to: %s", localPathNight))

	progress("Restoring light mode...")
	time.Sleep(1 * time.Second) // Reduced from 2s
	err = commands.SetDarkModeForTUI(cfg, device, false)
	if err != nil {
		progress(fmt.Sprintf("Warning: failed to restore light mode: %v", err))
	}

	commands.CleanupRemoteFile(adbPath, device.Serial, remotePath)

	return nil
}
