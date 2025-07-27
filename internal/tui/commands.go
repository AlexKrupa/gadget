package tui

import (
	"adx/internal/adb"
	"adx/internal/commands"
	"adx/internal/config"
	"adx/internal/emulator"
	"fmt"
	"path/filepath"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// loadDevices loads connected ADB devices asynchronously
func loadDevices(adbPath string) tea.Cmd {
	return func() tea.Msg {
		devices, err := adb.GetConnectedDevices(adbPath)
		if err == nil {
			// Load extended info for each device
			for i := range devices {
				devices[i].LoadExtendedInfo(adbPath)
			}
		}
		return devicesLoadedMsg{
			devices: devices,
			err:     err,
		}
	}
}

// loadAVDs loads available Android Virtual Devices asynchronously
func loadAVDs(cfg *config.Config) tea.Cmd {
	return func() tea.Msg {
		avds, err := emulator.GetAvailableAVDs(cfg)
		return avdsLoadedMsg{
			avds: avds,
			err:  err,
		}
	}
}

// ScreenshotOperation defines the type of screenshot operation
type ScreenshotOperation int

const (
	ScreenshotSingle ScreenshotOperation = iota
	ScreenshotDayNight
)

// executeScreenshotOperation executes a screenshot operation asynchronously with common handling
func executeScreenshotOperation(cfg *config.Config, device adb.Device, operation ScreenshotOperation) tea.Cmd {
	return func() tea.Msg {
		timestamp := time.Now().Format("2006-01-02_15-04-05")

		switch operation {
		case ScreenshotSingle:
			err := commands.TakeScreenshot(cfg, device)
			if err != nil {
				return screenshotDoneMsg{
					success: false,
					message: fmt.Sprintf("Screenshot failed on %s: %s", device.Serial, err.Error()),
				}
			}

			filename := fmt.Sprintf("android-img-%s.png", timestamp)
			localPath := filepath.Join(cfg.MediaPath, filename)
			message := fmt.Sprintf("Screenshot captured on %s\n%s", device.Serial, shortenHomePath(localPath))
			return screenshotDoneMsg{success: true, message: message}

		case ScreenshotDayNight:
			err := commands.TakeDayNightScreenshots(cfg, device)
			if err != nil {
				return dayNightScreenshotDoneMsg{success: false, message: err.Error()}
			}

			filenameDay := fmt.Sprintf("android-img-%s-day.png", timestamp)
			filenameNight := fmt.Sprintf("android-img-%s-night.png", timestamp)
			localPathDay := filepath.Join(cfg.MediaPath, filenameDay)
			localPathNight := filepath.Join(cfg.MediaPath, filenameNight)

			message := fmt.Sprintf("Day-night screenshots captured on %s\nDay: %s\nNight: %s",
				device.Serial, shortenHomePath(localPathDay), shortenHomePath(localPathNight))
			return dayNightScreenshotDoneMsg{success: true, message: message}
		}

		return nil // Should never reach here
	}
}

// takeScreenshot executes screenshot command asynchronously with detailed logging
func takeScreenshot(cfg *config.Config, device adb.Device) tea.Cmd {
	return executeScreenshotOperation(cfg, device, ScreenshotSingle)
}

// takeDayNightScreenshots executes day-night screenshot command asynchronously
func takeDayNightScreenshots(cfg *config.Config, device adb.Device) tea.Cmd {
	return executeScreenshotOperation(cfg, device, ScreenshotDayNight)
}

// startRecording starts screen recording asynchronously
func startRecording(cfg *config.Config, device adb.Device) tea.Cmd {
	return func() tea.Msg {
		recording, err := commands.StartScreenRecord(cfg, device)
		return recordingStartedMsg{
			recording: recording,
			err:       err,
		}
	}
}

// stopAndSaveRecording stops and saves the recording asynchronously
func stopAndSaveRecording(recording *commands.ScreenRecording) tea.Cmd {
	return func() tea.Msg {
		err := recording.StopAndSave()
		if err != nil {
			return screenRecordDoneMsg{
				success: false,
				message: err.Error(),
			}
		}
		return screenRecordDoneMsg{
			success: true,
			message: fmt.Sprintf("Screen recording saved on %s\n%s", recording.Device.Serial, shortenHomePath(recording.LocalPath)),
		}
	}
}

// getCurrentSetting gets the current setting from device
func getCurrentSetting(cfg *config.Config, device adb.Device, settingType commands.SettingType) tea.Cmd {
	return func() tea.Msg {
		handler := commands.GetSettingHandler(settingType)
		if handler == nil {
			return settingLoadedMsg{
				settingInfo: nil,
				err:         fmt.Errorf("unknown setting type: %s", settingType),
			}
		}

		settingInfo, err := handler.GetInfo(cfg, device)
		return settingLoadedMsg{
			settingInfo: settingInfo,
			err:         err,
		}
	}
}

// changeSetting changes the device setting asynchronously
func changeSetting(cfg *config.Config, device adb.Device, settingType commands.SettingType, value string) tea.Cmd {
	return func() tea.Msg {
		handler := commands.GetSettingHandler(settingType)
		if handler == nil {
			return settingChangedMsg{
				settingType: settingType,
				success:     false,
				message:     "Unknown setting type: " + string(settingType),
			}
		}

		// Validate input first
		if err := handler.ValidateInput(value); err != nil {
			return settingChangedMsg{
				settingType: settingType,
				success:     false,
				message:     err.Error(),
			}
		}

		// Apply the setting
		err := handler.SetValue(cfg, device, value)
		if err != nil {
			return settingChangedMsg{
				settingType: settingType,
				success:     false,
				message:     err.Error(),
			}
		}

		return settingChangedMsg{
			settingType: settingType,
			success:     true,
			message:     fmt.Sprintf("%s changed to %s on %s", settingType, value, device.Serial),
		}
	}
}

// WiFiOperation defines the type of WiFi operation
type WiFiOperation int

const (
	WiFiConnect WiFiOperation = iota
	WiFiDisconnect
	WiFiPair
)

// executeWiFiOperation executes a WiFi operation asynchronously with generic handling
func executeWiFiOperation(cfg *config.Config, operation WiFiOperation, ipAndPort, pairingCode string) tea.Cmd {
	return func() tea.Msg {
		var err error
		var successMsg string

		switch operation {
		case WiFiConnect:
			err = commands.ConnectWiFi(cfg, ipAndPort)
			successMsg = fmt.Sprintf("WiFi device connected: %s", ipAndPort)
		case WiFiDisconnect:
			err = commands.DisconnectWiFi(cfg, ipAndPort)
			successMsg = fmt.Sprintf("WiFi device disconnected: %s", ipAndPort)
		case WiFiPair:
			err = commands.PairWiFiDevice(cfg, ipAndPort, pairingCode)
			successMsg = fmt.Sprintf("WiFi device paired and connected: %s", ipAndPort)
		}

		if err != nil {
			switch operation {
			case WiFiConnect:
				return wifiConnectDoneMsg{success: false, message: err.Error()}
			case WiFiDisconnect:
				return wifiDisconnectDoneMsg{success: false, message: err.Error()}
			case WiFiPair:
				return wifiPairDoneMsg{success: false, message: err.Error()}
			}
		}

		// Small delay for connect/disconnect operations to ensure device list is updated
		if operation == WiFiConnect || operation == WiFiDisconnect {
			time.Sleep(500 * time.Millisecond)
		}

		switch operation {
		case WiFiConnect:
			return wifiConnectDoneMsg{success: true, message: successMsg}
		case WiFiDisconnect:
			return wifiDisconnectDoneMsg{success: true, message: successMsg}
		case WiFiPair:
			return wifiPairDoneMsg{success: true, message: successMsg}
		}

		return nil // Should never reach here
	}
}

// connectWiFi connects to a WiFi device asynchronously
func connectWiFi(cfg *config.Config, ipAndPort string) tea.Cmd {
	return executeWiFiOperation(cfg, WiFiConnect, ipAndPort, "")
}

// disconnectWiFi disconnects from a WiFi device asynchronously
func disconnectWiFi(cfg *config.Config, ipAndPort string) tea.Cmd {
	return executeWiFiOperation(cfg, WiFiDisconnect, ipAndPort, "")
}

// pairWiFi pairs with a WiFi device asynchronously
func pairWiFi(cfg *config.Config, ipAndPort, pairingCode string) tea.Cmd {
	return executeWiFiOperation(cfg, WiFiPair, ipAndPort, pairingCode)
}
