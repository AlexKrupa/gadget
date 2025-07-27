package tui

import (
	"adx/internal/adb"
	"adx/internal/commands"
	"adx/internal/config"
	"adx/internal/tui/features/media"
	"adx/internal/tui/features/wifi"
	"adx/internal/tui/messaging"

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

// Delegate WiFi operations to WiFi feature
func connectWiFi(cfg *config.Config, ipAndPort string) tea.Cmd {
	return wifi.ConnectWiFiCmd(cfg, ipAndPort)
}

func disconnectWiFi(cfg *config.Config, ipAndPort string) tea.Cmd {
	return wifi.DisconnectWiFiCmd(cfg, ipAndPort)
}

func pairWiFi(cfg *config.Config, ipAndPort, pairingCode string) tea.Cmd {
	return wifi.PairWiFiCmd(cfg, ipAndPort, pairingCode)
}
