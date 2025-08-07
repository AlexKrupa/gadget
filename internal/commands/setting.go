package commands

import (
	"fmt"
	"gadget/internal/adb"
	"gadget/internal/config"
	"strconv"
)

// SettingType represents different types of device settings
type SettingType string

const (
	SettingTypeDPI        SettingType = "dpi"
	SettingTypeFontSize   SettingType = "fontsize"
	SettingTypeScreenSize SettingType = "screensize"
)

type SettingInfo struct {
	Type        SettingType
	DisplayName string
	Current     string
	Default     string
	InputPrompt string
}

// SettingHandler defines the interface for device settings
type SettingHandler interface {
	GetInfo(cfg *config.Config, device adb.Device) (*SettingInfo, error)
	SetValue(cfg *config.Config, device adb.Device, value string) error
}

// GetSettingHandler returns the appropriate handler for a setting type
func GetSettingHandler(settingType SettingType) SettingHandler {
	switch settingType {
	case SettingTypeDPI:
		return &dpiHandler{}
	case SettingTypeFontSize:
		return &fontSizeHandler{}
	case SettingTypeScreenSize:
		return &screenSizeHandler{}
	default:
		return nil
	}
}

type dpiHandler struct{}

func (h *dpiHandler) GetInfo(cfg *config.Config, device adb.Device) (*SettingInfo, error) {
	dpiInfo, err := GetCurrentDPI(cfg, device)
	if err != nil {
		return nil, err
	}

	return &SettingInfo{
		Type:        SettingTypeDPI,
		DisplayName: "DPI",
		Current:     fmt.Sprintf("%d", dpiInfo.Current),
		Default:     fmt.Sprintf("%d", dpiInfo.Physical),
		InputPrompt: "Enter new DPI:",
	}, nil
}

func (h *dpiHandler) SetValue(cfg *config.Config, device adb.Device, value string) error {
	dpi, err := strconv.Atoi(value)
	if err != nil {
		return fmt.Errorf("invalid DPI value: %s", value)
	}
	return SetDPI(cfg, device, dpi)
}


type fontSizeHandler struct{}

func (h *fontSizeHandler) GetInfo(cfg *config.Config, device adb.Device) (*SettingInfo, error) {
	fontInfo, err := GetCurrentFontSize(cfg, device)
	if err != nil {
		return nil, err
	}

	return &SettingInfo{
		Type:        SettingTypeFontSize,
		DisplayName: "Font size",
		Current:     fmt.Sprintf("%.1f", fontInfo.Current),
		Default:     fmt.Sprintf("%.1f", fontInfo.Default),
		InputPrompt: "Enter new font size (e.g., 1.2):",
	}, nil
}

func (h *fontSizeHandler) SetValue(cfg *config.Config, device adb.Device, value string) error {
	scale, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return fmt.Errorf("invalid font size value: %s", value)
	}
	return SetFontSize(cfg, device, scale)
}


type screenSizeHandler struct{}

func (h *screenSizeHandler) GetInfo(cfg *config.Config, device adb.Device) (*SettingInfo, error) {
	screenInfo, err := GetCurrentScreenSize(cfg, device)
	if err != nil {
		return nil, err
	}

	return &SettingInfo{
		Type:        SettingTypeScreenSize,
		DisplayName: "Screen size",
		Current:     screenInfo.Current,
		Default:     screenInfo.Physical,
		InputPrompt: "Enter new screen size (e.g., 1080x1920):",
	}, nil
}

func (h *screenSizeHandler) SetValue(cfg *config.Config, device adb.Device, value string) error {
	return SetScreenSize(cfg, device, value)
}

