package cli

import (
	"adx/internal/adb"
	"adx/internal/commands"
	"adx/internal/config"
	"adx/internal/emulator"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

// CommandExecutor defines the signature for command execution functions
type CommandExecutor func(cfg *config.Config, deviceSerial, ip, code, value string) error

// CommandRegistry holds all available commands and their executors
var CommandRegistry = map[string]CommandExecutor{
	"screenshot":            executeScreenshot,
	"screenshot-day-night":  executeScreenshotDayNight,
	"screen-record":         executeScreenRecord,
	"change-dpi":            executeChangeDPI,
	"change-font-size":      executeChangeFontSize,
	"change-screen-size":    executeChangeScreenSize,
	"launch-emulator":       executeLaunchEmulator,
	"pair-wifi":             executePairWiFi,
	"connect-wifi":          executeConnectWiFi,
	"disconnect-wifi":       executeDisconnectWiFi,
	"refresh-devices":       executeRefreshDevices,
}

// ExecuteCommand dispatches a command using the registry
func ExecuteCommand(cfg *config.Config, command, deviceSerial, ip, code, value string) error {
	executor, exists := CommandRegistry[command]
	if !exists {
		return fmt.Errorf("unknown command: %s", command)
	}
	return executor(cfg, deviceSerial, ip, code, value)
}

// Command execution wrapper functions
func executeScreenshot(cfg *config.Config, deviceSerial, _, _, _ string) error {
	return ExecuteScreenshotDirect(cfg, deviceSerial)
}

func executeScreenshotDayNight(cfg *config.Config, deviceSerial, _, _, _ string) error {
	return ExecuteScreenshotDayNightDirect(cfg, deviceSerial)
}

func executeScreenRecord(cfg *config.Config, deviceSerial, _, _, _ string) error {
	return ExecuteScreenRecordDirect(cfg, deviceSerial)
}

func executeChangeDPI(cfg *config.Config, deviceSerial, _, _, value string) error {
	return ExecuteChangeDPIDirect(cfg, deviceSerial, value)
}

func executeChangeFontSize(cfg *config.Config, deviceSerial, _, _, value string) error {
	return ExecuteChangeFontSizeDirect(cfg, deviceSerial, value)
}

func executeChangeScreenSize(cfg *config.Config, deviceSerial, _, _, value string) error {
	return ExecuteChangeScreenSizeDirect(cfg, deviceSerial, value)
}

func executeLaunchEmulator(cfg *config.Config, _, _, _, value string) error {
	return ExecuteLaunchEmulatorDirect(cfg, value)
}

func executePairWiFi(cfg *config.Config, _, ip, code, _ string) error {
	if ip == "" || code == "" {
		return fmt.Errorf("pair-wifi requires IP address and pairing code")
	}
	return commands.PairWiFiDevice(cfg, ip, code)
}

func executeConnectWiFi(cfg *config.Config, _, ip, _, _ string) error {
	if ip == "" {
		return fmt.Errorf("connect-wifi requires IP address")
	}
	return commands.ConnectWiFi(cfg, ip)
}

func executeDisconnectWiFi(cfg *config.Config, _, ip, _, _ string) error {
	if ip == "" {
		return fmt.Errorf("disconnect-wifi requires IP address")
	}
	return commands.DisconnectWiFi(cfg, ip)
}

func executeRefreshDevices(cfg *config.Config, _, _, _, _ string) error {
	return ExecuteRefreshDevices(cfg)
}

// selectDevice selects a device based on serial, or prompts if multiple devices
func selectDevice(cfg *config.Config, deviceSerial string) (adb.Device, error) {
	devices, err := adb.GetConnectedDevices(cfg.GetADBPath())
	if err != nil {
		return adb.Device{}, err
	}

	if len(devices) == 0 {
		return adb.Device{}, fmt.Errorf("no devices connected")
	}

	// If device serial specified, find it
	if deviceSerial != "" {
		for _, device := range devices {
			if device.Serial == deviceSerial {
				return device, nil
			}
		}
		return adb.Device{}, fmt.Errorf("device with serial %s not found", deviceSerial)
	}

	// If only one device, use it
	if len(devices) == 1 {
		return devices[0], nil
	}

	// Multiple devices, require explicit selection
	fmt.Println("Multiple devices connected. Please specify device with -device flag:")
	for _, device := range devices {
		fmt.Printf("  %s\n", device.String())
	}
	return adb.Device{}, fmt.Errorf("multiple devices connected, please specify -device")
}

func ExecuteScreenshotDirect(cfg *config.Config, deviceSerial string) error {
	device, err := selectDevice(cfg, deviceSerial)
	if err != nil {
		return err
	}

	fmt.Printf("Taking screenshot on device: %s\n", device.Serial)
	return commands.TakeScreenshot(cfg, device)
}

func ExecuteScreenshotDayNightDirect(cfg *config.Config, deviceSerial string) error {
	device, err := selectDevice(cfg, deviceSerial)
	if err != nil {
		return err
	}

	fmt.Printf("Taking day-night screenshots on device: %s\n", device.Serial)
	return commands.TakeDayNightScreenshots(cfg, device)
}

func ExecuteScreenRecordDirect(cfg *config.Config, deviceSerial string) error {
	device, err := selectDevice(cfg, deviceSerial)
	if err != nil {
		return err
	}

	fmt.Printf("Starting screen recording on device: %s\n", device.Serial)
	fmt.Println("Press Ctrl+C to stop recording...")

	recording, err := commands.StartScreenRecord(cfg, device)
	if err != nil {
		return err
	}

	// Wait for interrupt signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	fmt.Println("\nStopping recording...")
	return recording.StopAndSave()
}

func executeSettingChange(cfg *config.Config, deviceSerial, value string, settingType commands.SettingType, commandName, valueDescription, actionDescription string) error {
	if value == "" {
		return fmt.Errorf("%s requires -value (%s)", commandName, valueDescription)
	}

	device, err := selectDevice(cfg, deviceSerial)
	if err != nil {
		return err
	}

	handler := commands.GetSettingHandler(settingType)
	if err := handler.ValidateInput(value); err != nil {
		return err
	}

	fmt.Printf("%s to %s on device: %s\n", actionDescription, value, device.Serial)
	return handler.SetValue(cfg, device, value)
}

func ExecuteChangeDPIDirect(cfg *config.Config, deviceSerial, value string) error {
	return executeSettingChange(cfg, deviceSerial, value, commands.SettingTypeDPI, "change-dpi", "DPI number", "Changing DPI")
}

func ExecuteChangeFontSizeDirect(cfg *config.Config, deviceSerial, value string) error {
	return executeSettingChange(cfg, deviceSerial, value, commands.SettingTypeFontSize, "change-font-size", "font scale number", "Changing font size")
}

func ExecuteChangeScreenSizeDirect(cfg *config.Config, deviceSerial, value string) error {
	return executeSettingChange(cfg, deviceSerial, value, commands.SettingTypeScreenSize, "change-screen-size", "WIDTHxHEIGHT", "Changing screen size")
}

func ExecuteLaunchEmulatorDirect(cfg *config.Config, avdName string) error {
	avds, err := emulator.GetAvailableAVDs(cfg)
	if err != nil {
		return err
	}

	if len(avds) == 0 {
		return fmt.Errorf("no AVDs found")
	}

	// If AVD name specified, find it
	if avdName != "" {
		for _, avd := range avds {
			if avd.Name == avdName {
				fmt.Printf("Launching emulator: %s\n", avd.Name)
				return emulator.LaunchEmulator(cfg, avd)
			}
		}
		return fmt.Errorf("AVD with name %s not found", avdName)
	}

	// If only one AVD, use it
	if len(avds) == 1 {
		fmt.Printf("Launching emulator: %s\n", avds[0].Name)
		return emulator.LaunchEmulator(cfg, avds[0])
	}

	// Multiple AVDs, require explicit selection
	fmt.Println("Multiple AVDs available. Please specify AVD with -value flag:")
	for _, avd := range avds {
		fmt.Printf("  %s\n", avd.String())
	}
	return fmt.Errorf("multiple AVDs available, please specify -value")
}

func ExecuteRefreshDevices(cfg *config.Config) error {
	devices, err := adb.GetConnectedDevices(cfg.GetADBPath())
	if err != nil {
		return err
	}
	fmt.Printf("Connected devices: %d\n", len(devices))
	for _, device := range devices {
		fmt.Printf("  %s\n", device.String())
	}
	return nil
}