package settings

import (
	"adx/internal/adb"
	"adx/internal/tui/messaging"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

// HandleSettingLoaded handles the loading of setting information
func (s *SettingsFeature) HandleSettingLoaded(msg messaging.SettingLoadedMsg) (tea.Model, tea.Cmd, string, string) {
	if msg.Err != nil {
		return nil, nil, "", fmt.Sprintf("Failed to load setting info: %s", msg.Err.Error())
	}

	s.SetCurrentSettingInfo(msg.SettingInfo)
	return nil, nil, "", ""
}

// HandleSettingChanged handles the completion of a setting change operation
func (s *SettingsFeature) HandleSettingChanged(msg messaging.SettingChangedMsg, device adb.Device) (tea.Model, tea.Cmd, string, string) {
	if msg.Success {
		successMsg := fmt.Sprintf("Setting changed successfully: %s", msg.Message)
		return nil, LoadSettingCmd(s.config, device, msg.SettingType), successMsg, ""
	}
	return nil, nil, "", fmt.Sprintf("Setting change failed: %s", msg.Message)
}
