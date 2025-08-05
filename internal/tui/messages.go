package tui

import (
	"gadget/internal/tui/messaging"
)

// Type aliases for backward compatibility
type devicesLoadedMsg = messaging.DevicesLoadedMsg
type avdsLoadedMsg = messaging.AvdsLoadedMsg
type screenshotDoneMsg = messaging.ScreenshotDoneMsg
type dayNightScreenshotDoneMsg = messaging.DayNightScreenshotDoneMsg
type screenRecordDoneMsg = messaging.ScreenRecordDoneMsg
type recordingStartedMsg = messaging.RecordingStartedMsg
type settingLoadedMsg = messaging.SettingLoadedMsg
type settingChangedMsg = messaging.SettingChangedMsg
type wifiConnectDoneMsg = messaging.WiFiConnectDoneMsg
type wifiDisconnectDoneMsg = messaging.WiFiDisconnectDoneMsg
type wifiPairDoneMsg = messaging.WiFiPairDoneMsg
type emulatorConfigureDoneMsg = messaging.EmulatorConfigureDoneMsg
type liveOutputMsg = messaging.LiveOutputMsg
