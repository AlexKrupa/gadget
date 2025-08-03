package devices

import (
	"gadget/internal/adb"
	"gadget/internal/config"
	"gadget/internal/emulator"
)

// DevicesFeature handles device selection and emulator management
type DevicesFeature struct {
	config           *config.Config
	devices          []adb.Device
	avds             []emulator.AVD
	selectedDevice   int
	selectedEmulator int
}

// NewDevicesFeature creates a new devices feature instance
func NewDevicesFeature(cfg *config.Config) *DevicesFeature {
	return &DevicesFeature{
		config:           cfg,
		selectedDevice:   0,
		selectedEmulator: 0,
	}
}

// GetDevices returns the current device list
func (d *DevicesFeature) GetDevices() []adb.Device {
	return d.devices
}

// GetAvds returns the current AVD list
func (d *DevicesFeature) GetAvds() []emulator.AVD {
	return d.avds
}

// GetSelectedDevice returns the currently selected device index
func (d *DevicesFeature) GetSelectedDevice() int {
	return d.selectedDevice
}

// GetSelectedEmulator returns the currently selected emulator index
func (d *DevicesFeature) GetSelectedEmulator() int {
	return d.selectedEmulator
}

// SetSelectedDevice sets the selected device index
func (d *DevicesFeature) SetSelectedDevice(index int) {
	if index >= 0 && index < len(d.devices) {
		d.selectedDevice = index
	}
}

// SetSelectedEmulator sets the selected emulator index
func (d *DevicesFeature) SetSelectedEmulator(index int) {
	if index >= 0 && index < len(d.avds) {
		d.selectedEmulator = index
	}
}

// GetSelectedDeviceInstance returns the currently selected device instance
func (d *DevicesFeature) GetSelectedDeviceInstance() *adb.Device {
	if d.selectedDevice < len(d.devices) {
		return &d.devices[d.selectedDevice]
	}
	return nil
}

// GetSelectedEmulatorInstance returns the currently selected emulator instance
func (d *DevicesFeature) GetSelectedEmulatorInstance() *emulator.AVD {
	if d.selectedEmulator < len(d.avds) {
		return &d.avds[d.selectedEmulator]
	}
	return nil
}

// SetDevices updates the device list
func (d *DevicesFeature) SetDevices(devices []adb.Device) {
	d.devices = devices
	// Reset selection if out of bounds
	if d.selectedDevice >= len(devices) {
		d.selectedDevice = 0
	}
}

// SetAvds updates the AVD list
func (d *DevicesFeature) SetAvds(avds []emulator.AVD) {
	d.avds = avds
	// Reset selection if out of bounds
	if d.selectedEmulator >= len(avds) {
		d.selectedEmulator = 0
	}
}
