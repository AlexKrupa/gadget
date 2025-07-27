package messaging

import (
	"adx/internal/adb"
	"adx/internal/commands"
	"adx/internal/config"
	"adx/internal/emulator"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

// Command factories for async operations

// LoadDevicesCmd returns a command that loads connected devices
func LoadDevicesCmd(cfg *config.Config) tea.Cmd {
	return func() tea.Msg {
		devices, err := adb.GetConnectedDevices(cfg.GetADBPath())
		return DevicesLoadedMsg{Devices: devices, Err: err}
	}
}

// LoadAvdsCmd returns a command that loads available AVDs
func LoadAvdsCmd(cfg *config.Config) tea.Cmd {
	return func() tea.Msg {
		avds, err := emulator.GetAvailableAVDs(cfg)
		return AvdsLoadedMsg{Avds: avds, Err: err}
	}
}

// LoadSettingCmd returns a command that loads current setting value
func LoadSettingCmd(cfg *config.Config, device adb.Device, settingType commands.SettingType) tea.Cmd {
	return func() tea.Msg {
		handler := commands.GetSettingHandler(settingType)
		settingInfo, err := handler.GetInfo(cfg, device)
		return SettingLoadedMsg{SettingInfo: settingInfo, Err: err}
	}
}

// ChangeSettingCmd returns a command that changes a device setting
func ChangeSettingCmd(cfg *config.Config, device adb.Device, settingType commands.SettingType, value string) tea.Cmd {
	return func() tea.Msg {
		handler := commands.GetSettingHandler(settingType)

		// Validate input first
		if err := handler.ValidateInput(value); err != nil {
			return SettingChangedMsg{
				SettingType: settingType,
				Success:     false,
				Message:     err.Error(),
			}
		}

		err := handler.SetValue(cfg, device, value)

		var message string
		success := err == nil

		if success {
			message = fmt.Sprintf("%s changed to %s on %s", settingType, value, device.Serial)
		} else {
			message = fmt.Sprintf("Failed to change %s: %s", settingType, err.Error())
		}

		return SettingChangedMsg{
			SettingType: settingType,
			Success:     success,
			Message:     message,
		}
	}
}

// StartScreenRecordCmd returns a command that starts screen recording
func StartScreenRecordCmd(cfg *config.Config, device adb.Device) tea.Cmd {
	return func() tea.Msg {
		recording, err := commands.StartScreenRecord(cfg, device)
		return RecordingStartedMsg{Recording: recording, Err: err}
	}
}
