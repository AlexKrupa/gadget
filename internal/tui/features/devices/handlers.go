package devices

import (
	"fmt"
	"gadget/internal/emulator"
	"gadget/internal/tui/messaging"

	tea "github.com/charmbracelet/bubbletea"
)

// HandleDevicesLoaded handles the loading of device list
func (d *DevicesFeature) HandleDevicesLoaded(msg messaging.DevicesLoadedMsg) (tea.Model, tea.Cmd, string, string) {
	d.SetDevices(msg.Devices)

	if msg.Err != nil {
		return nil, nil, "", msg.Err.Error()
	}

	if len(msg.Devices) == 0 {
		return nil, nil, "", "no devices connected"
	}

	return nil, nil, "", ""
}

// HandleAvdsLoaded handles the loading of AVD list
func (d *DevicesFeature) HandleAvdsLoaded(msg messaging.AvdsLoadedMsg) (tea.Model, tea.Cmd, string, string) {
	d.SetAvds(msg.Avds)

	if msg.Err != nil {
		return nil, nil, "", msg.Err.Error()
	}

	return nil, nil, "", ""
}

// LaunchSelectedEmulator launches the currently selected emulator
func (d *DevicesFeature) LaunchSelectedEmulator() (tea.Model, tea.Cmd, string, string) {
	selectedAvd := d.GetSelectedEmulatorInstance()
	if selectedAvd == nil {
		return nil, nil, "", "no emulator selected"
	}

	err := emulator.LaunchEmulator(d.config, *selectedAvd)
	if err != nil {
		return nil, nil, "", fmt.Sprintf("failed to launch emulator: %v", err)
	}

	successMsg := fmt.Sprintf("Launched emulator: %s (may take a moment to appear)", selectedAvd.Name)
	// Return command to refresh device list after launching emulator
	return nil, LoadDevicesCmd(d.config), successMsg, ""
}
