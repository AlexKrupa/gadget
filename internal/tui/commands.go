package tui

import (
	"adx/internal/adb"
	"adx/internal/commands"
	"adx/internal/config"
	"adx/internal/emulator"
	"adx/internal/tui/features/devices"
	"adx/internal/tui/features/media"
	"adx/internal/tui/features/settings"
	"adx/internal/tui/features/wifi"
	"os"
	"os/exec"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
)

// loadDevices loads connected ADB devices asynchronously with extended info
func loadDevices(cfg *config.Config) tea.Cmd {
	return devices.LoadDevicesCmd(cfg)
}

// loadAVDs loads available Android Virtual Devices asynchronously
func loadAVDs(cfg *config.Config) tea.Cmd {
	return devices.LoadAvdsCmd(cfg)
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

// Delegate settings operations to settings feature
func getCurrentSetting(cfg *config.Config, device adb.Device, settingType commands.SettingType) tea.Cmd {
	return settings.LoadSettingCmd(cfg, device, settingType)
}

func changeSetting(cfg *config.Config, device adb.Device, settingType commands.SettingType, value string) tea.Cmd {
	return settings.ChangeSettingCmd(cfg, device, settingType, value)
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

// configureEmulatorCmd opens the AVD configuration file in editor using tea.ExecProcess
func configureEmulatorCmd(cfg *config.Config, avd emulator.AVD) tea.Cmd {
	configPath := filepath.Join(avd.Path, emulator.AVDConfigFile)

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return func() tea.Msg {
			return emulatorConfigureDoneMsg{
				Success: false,
				Message: "AVD config file not found: " + configPath,
			}
		}
	}

	// Get editor from environment, default to vi
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vi"
	}

	// Create the command to run the editor
	cmd := exec.Command(editor, configPath)

	// Use tea.ExecProcess to suspend TUI and run editor
	return tea.ExecProcess(cmd, func(err error) tea.Msg {
		if err != nil {
			return emulatorConfigureDoneMsg{
				Success: false,
				Message: "Failed to open editor: " + err.Error(),
			}
		}
		return emulatorConfigureDoneMsg{
			Success: true,
			Message: "Emulator configuration updated",
		}
	})
}
