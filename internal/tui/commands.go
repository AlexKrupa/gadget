package tui

import (
	"adx/internal/adb"
	"adx/internal/commands"
	"adx/internal/config"
	"adx/internal/emulator"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// shortenHomePath replaces home directory with ~ if applicable
func shortenHomePath(path string) string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return path
	}
	
	if strings.HasPrefix(path, homeDir) {
		return strings.Replace(path, homeDir, "~", 1)
	}
	
	return path
}

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

// loadDevices loads connected ADB devices asynchronously
func loadDevices(adbPath string) tea.Cmd {
	return func() tea.Msg {
		devices, err := adb.GetConnectedDevices(adbPath)
		if err == nil {
			// Load extended info for each device
			for i := range devices {
				devices[i].LoadExtendedInfo(adbPath)
			}
		}
		return devicesLoadedMsg{
			devices: devices,
			err:     err,
		}
	}
}

// loadAVDs loads available Android Virtual Devices asynchronously
func loadAVDs(cfg *config.Config) tea.Cmd {
	return func() tea.Msg {
		avds, err := emulator.GetAvailableAVDs(cfg)
		return avdsLoadedMsg{
			avds: avds,
			err:  err,
		}
	}
}

// takeScreenshot executes screenshot command asynchronously with detailed logging
func takeScreenshot(cfg *config.Config, device adb.Device) tea.Cmd {
	return func() tea.Msg {
		// Generate detailed progress messages
		timestamp := time.Now().Format("2006-01-02_15-04-05")
		filename := fmt.Sprintf("android-img-%s.png", timestamp)
		localPath := filepath.Join(cfg.MediaPath, filename)
		
		// Execute the command and build detailed message
		err := commands.TakeScreenshot(cfg, device)
		if err != nil {
			return screenshotDoneMsg{
				success: false,
				message: fmt.Sprintf("Screenshot failed on %s: %s", device.Serial, err.Error()),
			}
		}
		
		message := fmt.Sprintf("Screenshot captured on %s\n%s", device.Serial, shortenHomePath(localPath))
		return screenshotDoneMsg{
			success: true,
			message: message,
		}
	}
}

// takeDayNightScreenshots executes day-night screenshot command asynchronously
func takeDayNightScreenshots(cfg *config.Config, device adb.Device) tea.Cmd {
	return func() tea.Msg {
		err := commands.TakeDayNightScreenshots(cfg, device)
		if err != nil {
			return dayNightScreenshotDoneMsg{
				success: false,
				message: err.Error(),
			}
		}
		timestamp := time.Now().Format("2006-01-02_15-04-05")
		filenameDay := fmt.Sprintf("android-img-%s-day.png", timestamp)
		filenameNight := fmt.Sprintf("android-img-%s-night.png", timestamp)
		localPathDay := filepath.Join(cfg.MediaPath, filenameDay)
		localPathNight := filepath.Join(cfg.MediaPath, filenameNight)
		
		message := fmt.Sprintf("Day-night screenshots captured on %s\nDay: %s\nNight: %s", 
			device.Serial, shortenHomePath(localPathDay), shortenHomePath(localPathNight))
		return dayNightScreenshotDoneMsg{
			success: true,
			message: message,
		}
	}
}

// startRecording starts screen recording asynchronously
func startRecording(cfg *config.Config, device adb.Device) tea.Cmd {
	return func() tea.Msg {
		recording, err := commands.StartScreenRecord(cfg, device)
		return recordingStartedMsg{
			recording: recording,
			err:       err,
		}
	}
}

// stopAndSaveRecording stops and saves the recording asynchronously
func stopAndSaveRecording(recording *commands.ScreenRecording) tea.Cmd {
	return func() tea.Msg {
		err := recording.StopAndSave()
		if err != nil {
			return screenRecordDoneMsg{
				success: false,
				message: err.Error(),
			}
		}
		return screenRecordDoneMsg{
			success: true,
			message: fmt.Sprintf("Screen recording saved on %s\n%s", recording.Device.Serial, shortenHomePath(recording.LocalPath)),
		}
	}
}

// getCurrentSetting gets the current setting from device
func getCurrentSetting(cfg *config.Config, device adb.Device, settingType commands.SettingType) tea.Cmd {
	return func() tea.Msg {
		handler := commands.GetSettingHandler(settingType)
		if handler == nil {
			return settingLoadedMsg{
				settingInfo: nil,
				err:         fmt.Errorf("unknown setting type: %s", settingType),
			}
		}
		
		settingInfo, err := handler.GetInfo(cfg, device)
		return settingLoadedMsg{
			settingInfo: settingInfo,
			err:         err,
		}
	}
}

// changeSetting changes the device setting asynchronously
func changeSetting(cfg *config.Config, device adb.Device, settingType commands.SettingType, value string) tea.Cmd {
	return func() tea.Msg {
		handler := commands.GetSettingHandler(settingType)
		if handler == nil {
			return settingChangedMsg{
				settingType: settingType,
				success:     false,
				message:     "Unknown setting type: " + string(settingType),
			}
		}
		
		// Validate input first
		if err := handler.ValidateInput(value); err != nil {
			return settingChangedMsg{
				settingType: settingType,
				success:     false,
				message:     err.Error(),
			}
		}
		
		// Apply the setting
		err := handler.SetValue(cfg, device, value)
		if err != nil {
			return settingChangedMsg{
				settingType: settingType,
				success:     false,
				message:     err.Error(),
			}
		}
		
		return settingChangedMsg{
			settingType: settingType,
			success:     true,
			message:     fmt.Sprintf("%s changed to %s on %s", settingType, value, device.Serial),
		}
	}
}

// connectWiFi connects to a WiFi device asynchronously
func connectWiFi(cfg *config.Config, ipAndPort string) tea.Cmd {
	return func() tea.Msg {
		err := commands.ConnectWiFi(cfg, ipAndPort)
		if err != nil {
			return wifiConnectDoneMsg{
				success: false,
				message: err.Error(),
			}
		}
		
		// Small delay to ensure device list is updated
		time.Sleep(500 * time.Millisecond)
		
		return wifiConnectDoneMsg{
			success: true,
			message: fmt.Sprintf("WiFi device connected: %s", ipAndPort),
		}
	}
}

// disconnectWiFi disconnects from a WiFi device asynchronously
func disconnectWiFi(cfg *config.Config, ipAndPort string) tea.Cmd {
	return func() tea.Msg {
		err := commands.DisconnectWiFi(cfg, ipAndPort)
		if err != nil {
			return wifiDisconnectDoneMsg{
				success: false,
				message: err.Error(),
			}
		}
		
		// Small delay to ensure device list is updated
		time.Sleep(500 * time.Millisecond)
		
		return wifiDisconnectDoneMsg{
			success: true,
			message: fmt.Sprintf("WiFi device disconnected: %s", ipAndPort),
		}
	}
}

// pairWiFi pairs with a WiFi device asynchronously
func pairWiFi(cfg *config.Config, ipAndPort, pairingCode string) tea.Cmd {
	return func() tea.Msg {
		err := commands.PairWiFiDevice(cfg, ipAndPort, pairingCode)
		if err != nil {
			return wifiPairDoneMsg{
				success: false,
				message: err.Error(),
			}
		}
		
		return wifiPairDoneMsg{
			success: true,
			message: fmt.Sprintf("WiFi device paired and connected: %s", ipAndPort),
		}
	}
}