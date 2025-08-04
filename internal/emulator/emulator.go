package emulator

import (
	"bufio"
	"fmt"
	"gadget/internal/config"
	"gadget/internal/display"
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
	// Always use Name (AVD ID) for consistency and CLI compatibility
	// Display name can be shown in extended info if needed
	return a.Name
}

// GetExtendedInfo returns a formatted string with extended AVD information
func (a AVD) GetExtendedInfo() string {
	var info []string

	if a.APILevel != "" {
		info = append(info, fmt.Sprintf("API %s", a.APILevel))
	}

	// CPU Architecture - use same format as devices
	if a.Architecture != "" {
		cpuDisplay := display.NormalizeCPUArchitecture(a.Architecture)
		info = append(info, fmt.Sprintf("%s %s", display.IconCPU, cpuDisplay))
	}

	if a.Resolution != "" {
		info = append(info, fmt.Sprintf("%s %s", display.IconScreen, a.Resolution))
	}

	if len(info) == 0 {
		return ""
	}

	return strings.Join(info, " • ")
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

			iniPath := filepath.Join(avdDir, entry.Name())

			target, actualPath := readTargetAndPathFromIni(iniPath)

			// Use the actual path from .ini file, fallback to constructed path
			configPath := filepath.Join(actualPath, AVDConfigFile)
			if actualPath == "" {
				actualPath = filepath.Join(avdDir, avdName+".avd")
				configPath = filepath.Join(actualPath, AVDConfigFile)
			}

			avdDetails := readAVDDetails(configPath)

			avd := AVD{
				Name:   avdName,
				Target: target,
				Path:   actualPath,
			}

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

	if width != "" && height != "" {
		details.Resolution = width + "x" + height
	}

	return details
}

const AVDConfigFile = "config.ini"

// SelectAVD handles common AVD selection logic for CLI commands
func SelectAVD(cfg *config.Config, avdName string) (*AVD, error) {
	avds, err := GetAvailableAVDs(cfg)
	if err != nil {
		return nil, err
	}

	if len(avds) == 0 {
		return nil, fmt.Errorf("no AVDs found")
	}

	if avdName != "" {
		for _, avd := range avds {
			if avd.Name == avdName {
				return &avd, nil
			}
		}
		return nil, fmt.Errorf("AVD with name %s not found", avdName)
	}

	if len(avds) == 1 {
		return &avds[0], nil
	}

	fmt.Println("Multiple AVDs available. Please specify AVD name:")
	for _, avd := range avds {
		fmt.Printf("  %s\n", avd.String())
	}
	return nil, fmt.Errorf("AVD name required")
}

// OpenConfigInEditor opens the AVD's config.ini file in the user's $EDITOR
func OpenConfigInEditor(avd AVD) error {
	configPath := filepath.Join(avd.Path, AVDConfigFile)

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return fmt.Errorf("AVD config file not found: %s", configPath)
	}

	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vi"
	}

	fmt.Printf("Opening AVD config for %s in %s...\n", avd.Name, editor)

	cmd := exec.Command(editor, configPath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
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
