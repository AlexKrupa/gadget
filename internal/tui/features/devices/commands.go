package devices

import (
	"gadget/internal/adb"
	"gadget/internal/config"
	"gadget/internal/tui/messaging"
	"time"

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

// StartDeviceTrackingCmd starts monitoring device changes via adb track-devices
func StartDeviceTrackingCmd(cfg *config.Config) tea.Cmd {
	return func() tea.Msg {
		eventChan, err := adb.StartDeviceTracking(cfg.GetADBPath())
		if err != nil {
			return messaging.DeviceRefreshMsg{Reason: "tracking-error"}
		}

		// Create a command that listens to device changes
		return DeviceTrackingStartedMsg{EventChan: eventChan}
	}
}

// DeviceTrackingStartedMsg contains the device event channel
type DeviceTrackingStartedMsg struct {
	EventChan <-chan adb.DeviceChangeEvent
}

// WaitForDeviceChangeCmd waits for the next device change event
func WaitForDeviceChangeCmd(eventChan <-chan adb.DeviceChangeEvent) tea.Cmd {
	return func() tea.Msg {
		event := <-eventChan
		_ = event // Use event for debugging if needed
		
		// Return after a brief delay to let device settle
		time.Sleep(500 * time.Millisecond)
		return messaging.DeviceRefreshMsg{Reason: "device-changed"}
	}
}

// ScheduleEmulatorRefreshCmd schedules a device refresh after emulator launch
func ScheduleEmulatorRefreshCmd() tea.Cmd {
	return func() tea.Msg {
		// Emulators typically take 15-30 seconds to become available
		time.Sleep(20 * time.Second)
		return messaging.DeviceRefreshMsg{Reason: "emulator-ready"}
	}
}

// StartPeriodicRefreshCmd starts periodic device refresh every 15 seconds
func StartPeriodicRefreshCmd() tea.Cmd {
	return func() tea.Msg {
		time.Sleep(15 * time.Second)
		return messaging.DeviceRefreshMsg{Reason: "periodic"}
	}
}
