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

