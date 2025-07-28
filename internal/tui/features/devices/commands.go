package devices

import (
	"adx/internal/adb"
	"adx/internal/config"
	"adx/internal/tui/messaging"

	tea "github.com/charmbracelet/bubbletea"
)

// LoadDevicesCmd returns a command to load connected devices with extended info
func LoadDevicesCmd(cfg *config.Config) tea.Cmd {
	return func() tea.Msg {
		devices, err := adb.GetConnectedDevices(cfg.GetADBPath())
		if err == nil {
			// Load extended info for each device
			for i := range devices {
				devices[i].LoadExtendedInfo(cfg.GetADBPath())
			}
		}
		return messaging.DevicesLoadedMsg{
			Devices: devices,
			Err:     err,
		}
	}
}

// LoadAvdsCmd returns a command to load available AVDs
func LoadAvdsCmd(cfg *config.Config) tea.Cmd {
	return messaging.LoadAvdsCmd(cfg)
}
