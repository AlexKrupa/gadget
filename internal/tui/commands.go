package tui

import (
	"adx/internal/adb"
	"adx/internal/commands"
	"adx/internal/config"
	"adx/internal/tui/features/media"
	"adx/internal/tui/messaging"
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// loadDevices loads connected ADB devices asynchronously with extended info
func loadDevices(adbPath string) tea.Cmd {
	return func() tea.Msg {
		devices, err := adb.GetConnectedDevices(adbPath)
		if err == nil {
			// Load extended info for each device
			for i := range devices {
				devices[i].LoadExtendedInfo(adbPath)
			}
		}
		return messaging.DevicesLoadedMsg{
			Devices: devices,
			Err:     err,
		}
	}
}

// loadAVDs loads available Android Virtual Devices asynchronously
func loadAVDs(cfg *config.Config) tea.Cmd {
	return messaging.LoadAvdsCmd(cfg)
}

// Delegate screenshot and recording operations to media feature
func takeScreenshot(cfg *config.Config, device adb.Device) tea.Cmd {
	return media.TakeScreenshotCmd(cfg, device)
}

func takeDayNightScreenshots(cfg *config.Config, device adb.Device) tea.Cmd {
	return media.TakeDayNightScreenshotsCmd(cfg, device)
}

func startRecording(cfg *config.Config, device adb.Device) tea.Cmd {
	return media.StartScreenRecordCmd(cfg, device)
}

func stopAndSaveRecording(recording *commands.ScreenRecording) tea.Cmd {
	return media.StopAndSaveRecordingCmd(recording)
}

// getCurrentSetting gets the current setting from device
func getCurrentSetting(cfg *config.Config, device adb.Device, settingType commands.SettingType) tea.Cmd {
	return messaging.LoadSettingCmd(cfg, device, settingType)
}

// changeSetting changes the device setting asynchronously
func changeSetting(cfg *config.Config, device adb.Device, settingType commands.SettingType, value string) tea.Cmd {
	return messaging.ChangeSettingCmd(cfg, device, settingType, value)
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
				return wifiConnectDoneMsg{Success: false, Message: err.Error()}
			case WiFiDisconnect:
				return wifiDisconnectDoneMsg{Success: false, Message: err.Error()}
			case WiFiPair:
				return wifiPairDoneMsg{Success: false, Message: err.Error()}
			}
		}

		// Small delay for connect/disconnect operations to ensure device list is updated
		if operation == WiFiConnect || operation == WiFiDisconnect {
			time.Sleep(500 * time.Millisecond)
		}

		switch operation {
		case WiFiConnect:
			return wifiConnectDoneMsg{Success: true, Message: successMsg}
		case WiFiDisconnect:
			return wifiDisconnectDoneMsg{Success: true, Message: successMsg}
		case WiFiPair:
			return wifiPairDoneMsg{Success: true, Message: successMsg}
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
