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

// takeScreenshotSilent takes a screenshot without printing output
func takeScreenshotSilent(adbPath, serial, remotePath, localPath string) error {
	err := adb.ExecuteCommand(adbPath, serial, "shell", "screencap", remotePath)
	if err != nil {
		return fmt.Errorf("failed to take screenshot: %w", err)
	}

	err = adb.ExecuteCommand(adbPath, serial, "pull", remotePath, localPath)
	if err != nil {
		return fmt.Errorf("failed to pull screenshot: %w", err)
	}

	return nil
}

// StreamingDayNightScreenshot represents a request to start streaming day-night screenshots
type StreamingDayNightScreenshot struct {
	Config    *config.Config
	Device    adb.Device
	Timestamp string
}

// TakeScreenshotCmd returns a command to take a single screenshot
func TakeScreenshotCmd(cfg *config.Config, device adb.Device) tea.Cmd {
	return StreamCommand(func() error {
		return commands.TakeScreenshot(cfg, device)
	})
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
			// Use generic streaming for single screenshots
			return StreamCommand(func() error {
				return commands.TakeScreenshot(cfg, device)
			})()

		case ScreenshotDayNight:
			// Use live streaming for day-night (needs progress updates)
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

	go func() {
		defer close(outputChan)

		sendProgress := func(message string) {
			select {
			case outputChan <- message:
			default:
				// Channel full, skip to prevent blocking
			}
		}

		err := executeDayNightWithProgress(cfg, device, timestamp, sendProgress)
		if err != nil {
			sendProgress(fmt.Sprintf("Command failed: %v", err))
		}
	}()

	return StreamingCommandStart{
		OutputChan: outputChan,
		Config:     cfg,
		Device:     device,
		Timestamp:  timestamp,
	}
}

// StreamCommand wraps any existing command function to make it stream output to logs
func StreamCommand(commandFunc func() error) tea.Cmd {
	return func() tea.Msg {
		outputChan := make(chan string, 100)

		go func() {
			defer close(outputChan)

			// Capture the command's output and stream it line by line
			capturedOutput, err := capture.CaptureCommand(commandFunc)

			// Send each captured line immediately
			for _, line := range capturedOutput {
				select {
				case outputChan <- line:
				default:
					// Channel full, skip to prevent blocking
				}
			}

			// Send error if command failed
			if err != nil {
				select {
				case outputChan <- fmt.Sprintf("Command failed: %v", err):
				default:
				}
			}
		}()

		return GenericStreamingStart{
			OutputChan: outputChan,
		}
	}
}

// GenericStreamingStart signals the start of any streaming command
type GenericStreamingStart struct {
	OutputChan <-chan string
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
	err := commands.SetDarkMode(cfg, device, false)
	if err != nil {
		progress(fmt.Sprintf("Error setting light mode: %v", err))
		return err
	}
	time.Sleep(1 * time.Second)

	progress("Taking day screenshot...")
	err = takeScreenshotSilent(adbPath, device.Serial, remotePath, localPathDay)
	if err != nil {
		progress(fmt.Sprintf("Error taking day screenshot: %v", err))
		return err
	}
	progress(fmt.Sprintf("Day screenshot saved to: %s", localPathDay))

	progress("Setting dark mode...")
	err = commands.SetDarkMode(cfg, device, true)
	if err != nil {
		progress(fmt.Sprintf("Error setting dark mode: %v", err))
		return err
	}
	time.Sleep(1 * time.Second)

	progress("Taking night screenshot...")
	err = takeScreenshotSilent(adbPath, device.Serial, remotePath, localPathNight)
	if err != nil {
		progress(fmt.Sprintf("Error taking night screenshot: %v", err))
		return err
	}
	progress(fmt.Sprintf("Night screenshot saved to: %s", localPathNight))

	progress("Restoring light mode...")
	time.Sleep(1 * time.Second)
	err = commands.SetDarkMode(cfg, device, false)
	if err != nil {
		progress(fmt.Sprintf("Warning: failed to restore light mode: %v", err))
	}

	commands.CleanupRemoteFile(adbPath, device.Serial, remotePath)
	return nil
}
