package messaging

import (
	"gadget/internal/adb"
	"gadget/internal/commands"
	"gadget/internal/emulator"
)

// Base message types for async operations

// DevicesLoadedMsg is sent when device loading is complete
type DevicesLoadedMsg struct {
	Devices []adb.Device
	Err     error
}

// AvdsLoadedMsg is sent when AVD loading is complete
type AvdsLoadedMsg struct {
	Avds []emulator.AVD
	Err  error
}

// SettingLoadedMsg is sent when current setting is retrieved
type SettingLoadedMsg struct {
	SettingInfo *commands.SettingInfo
	Err         error
}

// SettingChangedMsg is sent when setting change is complete
type SettingChangedMsg struct {
	SettingType    commands.SettingType
	Success        bool
	Message        string
	CapturedOutput []string // Changed: Added captured command output
}

// RecordingStartedMsg is sent when screen recording starts successfully
type RecordingStartedMsg struct {
	Recording *commands.ScreenRecording
	Err       error
}

// Base result message for simple operations
type OperationResult struct {
	Success        bool
	Message        string
	CapturedOutput []string // Changed: Added captured command output
}

// Specific operation result messages
type ScreenshotDoneMsg OperationResult
type DayNightScreenshotDoneMsg OperationResult
type ScreenRecordDoneMsg OperationResult
type WiFiConnectDoneMsg OperationResult
type WiFiDisconnectDoneMsg OperationResult
type WiFiPairDoneMsg OperationResult
type EmulatorConfigureDoneMsg OperationResult

// LiveOutputMsg is sent when command output is captured in real-time
type LiveOutputMsg struct {
	Message string
}
