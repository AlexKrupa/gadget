package tui

import (
	"gadget/internal/adb"
	"gadget/internal/commands"
	"gadget/internal/config"
	"gadget/internal/emulator"
	"gadget/internal/tui/features/devices"
	"gadget/internal/tui/features/media"
	"gadget/internal/tui/features/settings"
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

// loadLaunchableAVDs loads available AVDs excluding currently running ones
func loadLaunchableAVDs(cfg *config.Config, connectedDevices []adb.Device) tea.Cmd {
	return devices.LoadLaunchableAvdsCmd(cfg, connectedDevices)
}

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

func getCurrentSetting(cfg *config.Config, device adb.Device, settingType commands.SettingType) tea.Cmd {
	return settings.LoadSettingCmd(cfg, device, settingType)
}

func changeSetting(cfg *config.Config, device adb.Device, settingType commands.SettingType, value string) tea.Cmd {
	return settings.ChangeSettingCmd(cfg, device, settingType, value)
}

// configureEmulatorCmd opens the AVD configuration file in editor using tea.ExecProcess
func configureEmulatorCmd(cfg *config.Config, avd emulator.AVD) tea.Cmd {
	configPath := filepath.Join(avd.Path, emulator.AVDConfigFile)

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return func() tea.Msg {
			return emulatorConfigureDoneMsg{
				Success: false,
				Message: "AVD config file not found: " + configPath,
			}
		}
	}

	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vi"
	}

	cmd := exec.Command(editor, configPath)

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
