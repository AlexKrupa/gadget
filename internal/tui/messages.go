package tui

import (
	"adx/internal/adb"
	"adx/internal/commands"
	"adx/internal/emulator"
)

// devicesLoadedMsg is sent when device loading is complete
type devicesLoadedMsg struct {
	devices []adb.Device
	err     error
}

// avdsLoadedMsg is sent when AVD loading is complete
type avdsLoadedMsg struct {
	avds []emulator.AVD
	err  error
}

// screenshotDoneMsg is sent when screenshot operation is complete
type screenshotDoneMsg struct {
	success bool
	message string
}

// dayNightScreenshotDoneMsg is sent when day-night screenshot operation is complete
type dayNightScreenshotDoneMsg struct {
	success bool
	message string
}

// screenRecordDoneMsg is sent when screen recording operation is complete
type screenRecordDoneMsg struct {
	success bool
	message string
}

// recordingStartedMsg is sent when screen recording starts successfully
type recordingStartedMsg struct {
	recording *commands.ScreenRecording
	err       error
}

// settingLoadedMsg is sent when current setting is retrieved
type settingLoadedMsg struct {
	settingInfo *commands.SettingInfo
	err         error
}

// settingChangedMsg is sent when setting change is complete
type settingChangedMsg struct {
	settingType commands.SettingType
	success     bool
	message     string
}

// wifiConnectDoneMsg is sent when WiFi connection attempt is complete
type wifiConnectDoneMsg struct {
	success bool
	message string
}

// wifiDisconnectDoneMsg is sent when WiFi disconnection attempt is complete
type wifiDisconnectDoneMsg struct {
	success bool
	message string
}

// wifiPairDoneMsg is sent when WiFi pairing attempt is complete
type wifiPairDoneMsg struct {
	success bool
	message string
}
