package emulator

import (
	"adx/internal/config"
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// AVD represents an Android Virtual Device
type AVD struct {
	Name         string
	Target       string
	Path         string
	DisplayName  string
	Architecture string
	Resolution   string
	APILevel     string
}

// String returns a formatted string representation of the AVD
func (a AVD) String() string {
	var parts []string

	// Use DisplayName if available, otherwise Name
	name := a.Name
	if a.DisplayName != "" {
		name = a.DisplayName
	}
	parts = append(parts, name)

	// Add API level and architecture info
	var details []string
	if a.APILevel != "" {
		details = append(details, "API "+a.APILevel)
	}
	if a.Architecture != "" {
		details = append(details, a.Architecture)
	}
	if a.Resolution != "" {
		details = append(details, a.Resolution)
	}

	if len(details) > 0 {
		parts = append(parts, "("+strings.Join(details, " • ")+")")
	}

	return strings.Join(parts, " ")
}

// GetAvailableAVDs returns a list of available Android Virtual Devices
func GetAvailableAVDs(cfg *config.Config) ([]AVD, error) {
	// Use pure manual parsing for single source of truth
	return getAVDsFromDirectory()
}

// getAVDsFromDirectory parses AVDs directly from ~/.android/avd/
func getAVDsFromDirectory() ([]AVD, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	avdDir := filepath.Join(homeDir, ".android", "avd")
	entries, err := os.ReadDir(avdDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read AVD directory: %w", err)
	}

	var avds []AVD
	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), ".ini") {
			avdName := strings.TrimSuffix(entry.Name(), ".ini")

			// Read the .ini file to get target and path info
			iniPath := filepath.Join(avdDir, entry.Name())

			target, actualPath := readTargetAndPathFromIni(iniPath)

			// Use the actual path from .ini file, fallback to constructed path
			configPath := filepath.Join(actualPath, "config.ini")
			if actualPath == "" {
				actualPath = filepath.Join(avdDir, avdName+".avd")
				configPath = filepath.Join(actualPath, "config.ini")
			}

			avdDetails := readAVDDetails(configPath)

			avd := AVD{
				Name:   avdName,
				Target: target,
				Path:   actualPath,
			}

			// Populate with detailed info if available
			if avdDetails != nil {
				avd.DisplayName = avdDetails.DisplayName
				avd.Architecture = avdDetails.Architecture
				avd.Resolution = avdDetails.Resolution
				avd.APILevel = avdDetails.APILevel
			}

			avds = append(avds, avd)
		}
	}

	return avds, nil
}

// AVDDetails holds parsed config details from config.ini
type AVDDetails struct {
	DisplayName  string
	Architecture string
	Resolution   string
	APILevel     string
}

// readTargetFromIni reads target information from AVD .ini file
func readTargetFromIni(iniPath string) (string, error) {
	target, _ := readTargetAndPathFromIni(iniPath)
	return target, nil
}

// readTargetAndPathFromIni reads both target and path from AVD .ini file
func readTargetAndPathFromIni(iniPath string) (string, string) {
	file, err := os.Open(iniPath)
	if err != nil {
		return "", ""
	}
	defer file.Close()

	var target, path string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "target=") {
			target = strings.TrimPrefix(line, "target=")
		} else if strings.HasPrefix(line, "path=") {
			path = strings.TrimPrefix(line, "path=")
		}
	}

	return target, path
}

// readAVDDetails reads detailed configuration from config.ini file
func readAVDDetails(configPath string) *AVDDetails {
	file, err := os.Open(configPath)
	if err != nil {
		return nil
	}
	defer file.Close()

	details := &AVDDetails{}
	scanner := bufio.NewScanner(file)

	var width, height string

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, " = ", 2)
		if len(parts) != 2 {
			continue
		}

		key, value := strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])

		switch key {
		case "avd.ini.displayname":
			details.DisplayName = value
		case "abi.type", "hw.cpu.arch":
			if details.Architecture == "" {
				details.Architecture = value
			}
		case "hw.lcd.width":
			width = value
		case "hw.lcd.height":
			height = value
		}

		// Extract API level from target
		if strings.HasPrefix(key, "image.sysdir") && strings.Contains(value, "android-") {
			parts := strings.Split(value, "/")
			for _, part := range parts {
				if strings.HasPrefix(part, "android-") {
					details.APILevel = strings.TrimPrefix(part, "android-")
					break
				}
			}
		}
	}

	// Build resolution if we have both width and height
	if width != "" && height != "" {
		details.Resolution = width + "x" + height
	}

	return details
}

// LaunchEmulator starts the specified AVD
func LaunchEmulator(cfg *config.Config, avd AVD) error {
	emulatorPath := cfg.GetEmulatorPath()
	cmd := exec.Command(emulatorPath, "-avd", avd.Name, "-dns-server", "8.8.8.8")

	// Start emulator in background
	err := cmd.Start()
	if err != nil {
		return fmt.Errorf("failed to launch emulator: %w", err)
	}

	fmt.Printf("Launched emulator: %s (PID: %d)\n", avd.Name, cmd.Process.Pid)
	return nil
}
