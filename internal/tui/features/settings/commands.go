package settings

import (
	"gadget/internal/adb"
	"gadget/internal/commands"
	"gadget/internal/config"
	"gadget/internal/tui/messaging"

	tea "github.com/charmbracelet/bubbletea"
)

// LoadSettingCmd returns a command to load current setting value
func LoadSettingCmd(cfg *config.Config, device adb.Device, settingType commands.SettingType) tea.Cmd {
	return messaging.LoadSettingCmd(cfg, device, settingType)
}

// ChangeSettingCmd returns a command to change a device setting
func ChangeSettingCmd(cfg *config.Config, device adb.Device, settingType commands.SettingType, value string) tea.Cmd {
	return messaging.ChangeSettingCmd(cfg, device, settingType, value)
}

// ChangeDPICmd returns a command to change device DPI
func ChangeDPICmd(cfg *config.Config, device adb.Device, value string) tea.Cmd {
	return ChangeSettingCmd(cfg, device, commands.SettingTypeDPI, value)
}

// ChangeFontSizeCmd returns a command to change device font size
func ChangeFontSizeCmd(cfg *config.Config, device adb.Device, value string) tea.Cmd {
	return ChangeSettingCmd(cfg, device, commands.SettingTypeFontSize, value)
}

// ChangeScreenSizeCmd returns a command to change device screen size
func ChangeScreenSizeCmd(cfg *config.Config, device adb.Device, value string) tea.Cmd {
	return ChangeSettingCmd(cfg, device, commands.SettingTypeScreenSize, value)
}
