package settings

import (
	"adx/internal/commands"
	"adx/internal/config"
)

// SettingsFeature handles device settings operations
type SettingsFeature struct {
	config             *config.Config
	currentSettingInfo *commands.SettingInfo
	currentSettingType commands.SettingType
}

// NewSettingsFeature creates a new settings feature instance
func NewSettingsFeature(cfg *config.Config) *SettingsFeature {
	return &SettingsFeature{
		config: cfg,
	}
}

// GetCurrentSettingInfo returns the current setting info
func (s *SettingsFeature) GetCurrentSettingInfo() *commands.SettingInfo {
	return s.currentSettingInfo
}

// GetCurrentSettingType returns the current setting type
func (s *SettingsFeature) GetCurrentSettingType() commands.SettingType {
	return s.currentSettingType
}

// SetCurrentSettingInfo sets the current setting info and type
func (s *SettingsFeature) SetCurrentSettingInfo(info *commands.SettingInfo) {
	s.currentSettingInfo = info
	if info != nil {
		s.currentSettingType = info.Type
	}
}

// ClearCurrentSetting clears the current setting info
func (s *SettingsFeature) ClearCurrentSetting() {
	s.currentSettingInfo = nil
	s.currentSettingType = ""
}
