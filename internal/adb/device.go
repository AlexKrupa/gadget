package adb

import (
	"bufio"
	"fmt"
	"gadget/internal/display"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// Device represents an ADB device
type Device struct {
	Serial      string
	Status      string
	Product     string
	Model       string
	DeviceType  string
	TransportID string

	// Extended info (populated lazily)
	BatteryLevel    int // -1 if unknown
	AndroidVersion  string
	ScreenRes       string
	CPUArchitecture string
	APILevel        int // -1 if unknown
	IPAddress       string
}

// DeviceConnectionType represents the type of device connection
type DeviceConnectionType int

const (
	DeviceTypePhysical DeviceConnectionType = iota
	DeviceTypeEmulator
	DeviceTypeWiFi
)

// GetConnectionType returns the connection type of the device
func (d Device) GetConnectionType() DeviceConnectionType {
	if strings.HasPrefix(d.Serial, "emulator-") {
		return DeviceTypeEmulator
	}
	if strings.Contains(d.Serial, ":") {
		return DeviceTypeWiFi
	}
	return DeviceTypePhysical
}

// GetStatusIndicator returns a colored status indicator for the device
func (d Device) GetStatusIndicator() string {
	switch d.GetConnectionType() {
	case DeviceTypeEmulator:
		return "🟡" // Yellow dot for emulators
	case DeviceTypeWiFi:
		return "🟢" // Green dot for WiFi devices
	case DeviceTypePhysical:
		return "🔵" // Blue dot for physical devices
	default:
		return "⚪" // White dot for unknown
	}
}

// String returns a formatted string representation of the device
func (d Device) String() string {
	// Check if this is an emulator
	if strings.HasPrefix(d.Serial, "emulator-") {
		var details []string

		// Try to get AVD display name first
		avdDisplayName := getAVDDisplayNameForEmulator(d.Serial)
		if avdDisplayName != "" {
			details = append(details, avdDisplayName)
		} else {
			// Fallback to cleaning up model names for emulators
			if d.Model != "" && !strings.Contains(d.Model, "sdk_gphone") {
				details = append(details, d.Model)
			} else if d.Product != "" {
				// Use product name for emulators, clean it up
				productName := d.Product
				if strings.HasPrefix(productName, "sdk_") {
					productName = strings.TrimPrefix(productName, "sdk_")
					productName = strings.ReplaceAll(productName, "_", " ")
				}
				details = append(details, productName)
			}
		}

		details = append(details, "Emulator")

		if len(details) > 0 {
			return fmt.Sprintf("%s (%s)", d.Serial, strings.Join(details, " • "))
		}
	}

	// Regular device formatting
	if d.Model != "" && d.Product != "" {
		return fmt.Sprintf("%s (%s - %s)", d.Serial, d.Model, d.Product)
	}
	return fmt.Sprintf("%s (%s)", d.Serial, d.Status)
}

// GetConnectedDevices returns a list of connected ADB devices
func GetConnectedDevices(adbPath string) ([]Device, error) {
	cmd := exec.Command(adbPath, "devices", "-l")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get devices: %w", err)
	}

	var devices []Device
	scanner := bufio.NewScanner(strings.NewReader(string(output)))

	// Skip the header line "List of devices attached"
	if scanner.Scan() {
		// Skip header
	}

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		device := parseDeviceLine(line)
		if device != nil {
			devices = append(devices, *device)
		}
	}

	return devices, nil
}

// parseDeviceLine parses a single line from adb devices -l output
// Example: "emulator-5554    device product:sdk_gphone64_arm64 model:sdk_gphone64_arm64 device:emulator64_arm64"
func parseDeviceLine(line string) *Device {
	parts := strings.Fields(line)
	if len(parts) < 2 {
		return nil
	}

	device := &Device{
		Serial: parts[0],
		Status: parts[1],
	}

	// Parse additional properties like model:, product:, etc.
	for i := 2; i < len(parts); i++ {
		if strings.HasPrefix(parts[i], "model:") {
			device.Model = strings.TrimPrefix(parts[i], "model:")
		} else if strings.HasPrefix(parts[i], "product:") {
			device.Product = strings.TrimPrefix(parts[i], "product:")
		} else if strings.HasPrefix(parts[i], "device:") {
			device.DeviceType = strings.TrimPrefix(parts[i], "device:")
		} else if strings.HasPrefix(parts[i], "transport_id:") {
			device.TransportID = strings.TrimPrefix(parts[i], "transport_id:")
		}
	}

	return device
}

// ExecuteCommand runs an adb command on a specific device
func ExecuteCommand(adbPath, deviceSerial string, args ...string) error {
	cmdArgs := []string{"-s", deviceSerial}
	cmdArgs = append(cmdArgs, args...)

	cmd := exec.Command(adbPath, cmdArgs...)
	return cmd.Run()
}

// ExecuteGlobalCommand runs an adb command without targeting a specific device
func ExecuteGlobalCommand(adbPath string, args ...string) error {
	cmd := exec.Command(adbPath, args...)
	return cmd.Run()
}

// ExecuteGlobalCommandWithOutput runs an adb command without targeting a specific device and returns output
func ExecuteGlobalCommandWithOutput(adbPath string, args ...string) (string, error) {
	cmd := exec.Command(adbPath, args...)
	output, err := cmd.Output()
	return string(output), err
}

// ExecuteCommandWithOutput runs an adb command and returns output
func ExecuteCommandWithOutput(adbPath, deviceSerial string, args ...string) (string, error) {
	cmdArgs := []string{"-s", deviceSerial}
	cmdArgs = append(cmdArgs, args...)

	cmd := exec.Command(adbPath, cmdArgs...)
	output, err := cmd.Output()
	return string(output), err
}

// getAVDDisplayNameForEmulator tries to find the AVD display name for a running emulator
func getAVDDisplayNameForEmulator(serial string) string {
	// This is complex because there's no direct mapping between emulator-XXXX and AVD names
	// The emulator processes contain AVD names but not port info
	// For now, let's return empty and rely on fallback display
	return ""
}

// getDisplayNameFromAVDName gets the display name from AVD config files
func getDisplayNameFromAVDName(avdName string) string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	configPath := filepath.Join(homeDir, ".android", "avd", avdName+".avd", "config.ini")
	file, err := os.Open(configPath)
	if err != nil {
		return ""
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "avd.ini.displayname = ") {
			return strings.TrimPrefix(line, "avd.ini.displayname = ")
		}
	}

	return ""
}

// LoadExtendedInfo populates battery, Android version, screen resolution, CPU architecture, API level, and IP address for the device
func (d *Device) LoadExtendedInfo(adbPath string) {
	if d.Status != "device" {
		return // Only load info for connected devices
	}

	// Load battery level
	if batteryOutput, err := ExecuteCommandWithOutput(adbPath, d.Serial, "shell", "dumpsys", "battery"); err == nil {
		lines := strings.Split(strings.TrimSpace(batteryOutput), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.Contains(line, "level:") {
				parts := strings.Split(line, ":")
				if len(parts) >= 2 {
					if level, err := strconv.Atoi(strings.TrimSpace(parts[1])); err == nil {
						d.BatteryLevel = level
						break
					}
				}
			}
		}
	}
	if d.BatteryLevel == 0 {
		d.BatteryLevel = -1 // Unknown
	}

	// Load Android version
	if versionOutput, err := ExecuteCommandWithOutput(adbPath, d.Serial, "shell", "getprop", "ro.build.version.release"); err == nil {
		d.AndroidVersion = strings.TrimSpace(versionOutput)
	}

	// Load screen resolution
	if resOutput, err := ExecuteCommandWithOutput(adbPath, d.Serial, "shell", "wm", "size"); err == nil {
		lines := strings.Split(strings.TrimSpace(resOutput), "\n")
		for _, line := range lines {
			if strings.Contains(line, "Physical size:") {
				parts := strings.Split(line, ":")
				if len(parts) >= 2 {
					d.ScreenRes = strings.TrimSpace(parts[1])
					break
				}
			}
		}
	}

	// Load CPU architecture
	if cpuOutput, err := ExecuteCommandWithOutput(adbPath, d.Serial, "shell", "getprop", "ro.product.cpu.abi"); err == nil {
		d.CPUArchitecture = strings.TrimSpace(cpuOutput)
	}

	// Load API level
	if apiOutput, err := ExecuteCommandWithOutput(adbPath, d.Serial, "shell", "getprop", "ro.build.version.sdk"); err == nil {
		if apiLevel, err := strconv.Atoi(strings.TrimSpace(apiOutput)); err == nil {
			d.APILevel = apiLevel
		} else {
			d.APILevel = -1 // Unknown
		}
	} else {
		d.APILevel = -1 // Unknown
	}

	// Load IP address - try multiple methods
	d.loadIPAddress(adbPath)
}

// loadIPAddress attempts to get the device's IP address using various methods
func (d *Device) loadIPAddress(adbPath string) {
	// Method 1: Try to get WiFi IP address from wlan0 interface
	if ipOutput, err := ExecuteCommandWithOutput(adbPath, d.Serial, "shell", "ip", "addr", "show", "wlan0"); err == nil {
		lines := strings.Split(strings.TrimSpace(ipOutput), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.Contains(line, "inet ") && !strings.Contains(line, "127.0.0.1") {
				// Extract IP from line like "inet 192.168.1.100/24 brd 192.168.1.255 scope global wlan0"
				parts := strings.Fields(line)
				for i, part := range parts {
					if part == "inet" && i+1 < len(parts) {
						ipWithMask := parts[i+1]
						if slashIndex := strings.Index(ipWithMask, "/"); slashIndex != -1 {
							d.IPAddress = ipWithMask[:slashIndex]
						} else {
							d.IPAddress = ipWithMask
						}
						return
					}
				}
			}
		}
	}

	// Method 2: Try using ifconfig (fallback for older devices)
	if ipOutput, err := ExecuteCommandWithOutput(adbPath, d.Serial, "shell", "ifconfig", "wlan0"); err == nil {
		lines := strings.Split(strings.TrimSpace(ipOutput), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.Contains(line, "inet addr:") {
				// Extract IP from line like "inet addr:192.168.1.100  Bcast:192.168.1.255  Mask:255.255.255.0"
				if startIndex := strings.Index(line, "inet addr:"); startIndex != -1 {
					remaining := line[startIndex+10:] // Skip "inet addr:"
					if spaceIndex := strings.Index(remaining, " "); spaceIndex != -1 {
						d.IPAddress = remaining[:spaceIndex]
					} else {
						d.IPAddress = remaining
					}
					return
				}
			}
		}
	}

	// Method 3: For WiFi devices, extract IP from serial if it's in IP:port format
	if d.GetConnectionType() == DeviceTypeWiFi && strings.Contains(d.Serial, ":") {
		parts := strings.Split(d.Serial, ":")
		if len(parts) >= 2 {
			d.IPAddress = parts[0]
		}
	}
}

// GetExtendedInfo returns a formatted string with extended device information
func (d Device) GetExtendedInfo() string {
	var info []string

	// Android version and API level
	if d.AndroidVersion != "" && d.APILevel > 0 {
		info = append(info, fmt.Sprintf("Android %s (API %d)", d.AndroidVersion, d.APILevel))
	} else if d.AndroidVersion != "" {
		info = append(info, fmt.Sprintf("Android %s", d.AndroidVersion))
	} else if d.APILevel > 0 {
		info = append(info, fmt.Sprintf("API %d", d.APILevel))
	}

	// CPU Architecture
	if d.CPUArchitecture != "" {
		cpuDisplay := display.NormalizeCPUArchitecture(d.CPUArchitecture)
		info = append(info, fmt.Sprintf("%s %s", display.IconCPU, cpuDisplay))
	}

	// Screen Resolution
	if d.ScreenRes != "" {
		info = append(info, fmt.Sprintf("%s %s", display.IconScreen, d.ScreenRes))
	}

	// IP Address
	if d.IPAddress != "" {
		info = append(info, fmt.Sprintf("%s %s", display.IconNetwork, d.IPAddress))
	}

	if len(info) == 0 {
		return ""
	}

	return strings.Join(info, " • ")
}
