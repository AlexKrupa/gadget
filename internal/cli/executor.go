package cli

import (
	"fmt"
	"gadget/internal/adb"
	"gadget/internal/commands"
	"gadget/internal/config"
	"gadget/internal/display"
	"gadget/internal/emulator"
	"os"
	"os/signal"
	"syscall"
)

// CommandExecutor defines the signature for command execution functions
type CommandExecutor func(cfg *config.Config, deviceSerial, ip, code, value string) error

// NestedCommandExecutor defines the signature for nested command execution functions
type NestedCommandExecutor func(cfg *config.Config, args []string) error

// CommandRegistry holds all available commands and their executors
var CommandRegistry = map[string]CommandExecutor{
	"screenshot":           executeScreenshot,
	"screenshot-day-night": executeScreenshotDayNight,
	"screen-record":        executeScreenRecord,
	"dpi":                  executeDPI,
	"font-size":            executeFontSize,
	"screen-size":          executeScreenSize,
	"launch-emulator":      executeLaunchEmulator,
	"configure-emulator":   executeConfigureEmulator,
	"pair-wifi":            executePairWiFi,
	"connect-wifi":         executeConnectWiFi,
	"disconnect-wifi":      executeDisconnectWiFi,
	"refresh-devices":      executeRefreshDevices,
}

// NestedCommandRegistry holds nested commands and their executors
var NestedCommandRegistry = map[string]NestedCommandExecutor{
	"wifi": executeWiFiCommand,
}

// ExecuteCommand dispatches a command using the registry
func ExecuteCommand(cfg *config.Config, command, deviceSerial, ip, code, value string) error {
	executor, exists := CommandRegistry[command]
	if !exists {
		return fmt.Errorf("unknown command: %s", command)
	}
	return executor(cfg, deviceSerial, ip, code, value)
}

// ExecuteNestedCommand dispatches a nested command using the nested registry
func ExecuteNestedCommand(cfg *config.Config, command string, args []string) error {
	executor, exists := NestedCommandRegistry[command]
	if !exists {
		return fmt.Errorf("unknown nested command: %s", command)
	}
	return executor(cfg, args)
}

func executeScreenshot(cfg *config.Config, deviceSerial, _, _, _ string) error {
	return ExecuteScreenshotDirect(cfg, deviceSerial)
}

func executeScreenshotDayNight(cfg *config.Config, deviceSerial, _, _, _ string) error {
	return ExecuteScreenshotDayNightDirect(cfg, deviceSerial)
}

func executeScreenRecord(cfg *config.Config, deviceSerial, _, _, _ string) error {
	return ExecuteScreenRecordDirect(cfg, deviceSerial)
}

func executeDPI(cfg *config.Config, deviceSerial, _, _, value string) error {
	return ExecuteDPIDirect(cfg, deviceSerial, value)
}

func executeChangeFontSize(cfg *config.Config, deviceSerial, _, _, value string) error {
	return ExecuteFontSizeDirect(cfg, deviceSerial, value)
}

func executeChangeScreenSize(cfg *config.Config, deviceSerial, _, _, value string) error {
	return ExecuteScreenSizeDirect(cfg, deviceSerial, value)
}

func executeLaunchEmulator(cfg *config.Config, _, _, _, value string) error {
	return ExecuteLaunchEmulatorDirect(cfg, value)
}

func executeConfigureEmulator(cfg *config.Config, _, _, _, value string) error {
	return ExecuteConfigureEmulatorDirect(cfg, value)
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

	if deviceSerial != "" {
		for _, device := range devices {
			if device.Serial == deviceSerial {
				return device, nil
			}
		}
		return adb.Device{}, fmt.Errorf("device with serial %s not found", deviceSerial)
	}

	if len(devices) == 1 {
		return devices[0], nil
	}

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

func ExecuteDPIDirect(cfg *config.Config, deviceSerial, value string) error {
	return executeSettingCommand(cfg, deviceSerial, value, commands.SettingTypeDPI, "Physical DPI", "Current DPI")
}

// executeSettingCommand is a generic function for all setting commands
func executeSettingCommand(cfg *config.Config, deviceSerial, value string, settingType commands.SettingType, defaultLabel, currentLabel string) error {
	device, err := selectDevice(cfg, deviceSerial)
	if err != nil {
		return err
	}

	handler := commands.GetSettingHandler(settingType)

	// If no value provided, show current setting info
	if value == "" {
		info, err := handler.GetInfo(cfg, device)
		if err != nil {
			return err
		}
		fmt.Printf("%s: %s\n", defaultLabel, info.Default)
		fmt.Printf("%s: %s\n", currentLabel, info.Current)
		return nil
	}

	// Validate and set new value
	if err := handler.ValidateInput(value); err != nil {
		return err
	}

	if err := handler.SetValue(cfg, device, value); err != nil {
		return err
	}

	// Show setting info after setting
	info, err := handler.GetInfo(cfg, device)
	if err != nil {
		return err
	}
	fmt.Printf("%s: %s\n", defaultLabel, info.Default)
	fmt.Printf("%s: %s\n", currentLabel, info.Current)
	return nil
}

func executeFontSize(cfg *config.Config, deviceSerial, _, _, value string) error {
	return ExecuteFontSizeDirect(cfg, deviceSerial, value)
}

func ExecuteFontSizeDirect(cfg *config.Config, deviceSerial, value string) error {
	return executeSettingCommand(cfg, deviceSerial, value, commands.SettingTypeFontSize, "Default font size", "Current font size")
}

func executeScreenSize(cfg *config.Config, deviceSerial, _, _, value string) error {
	return ExecuteScreenSizeDirect(cfg, deviceSerial, value)
}

func ExecuteScreenSizeDirect(cfg *config.Config, deviceSerial, value string) error {
	return executeSettingCommand(cfg, deviceSerial, value, commands.SettingTypeScreenSize, "Physical screen size", "Current screen size")
}

func ExecuteLaunchEmulatorDirect(cfg *config.Config, avdName string) error {
	avd, err := emulator.SelectAVD(cfg, avdName)
	if err != nil {
		return err
	}
	fmt.Printf("Launching emulator: %s\n", avd.Name)
	return emulator.LaunchEmulator(cfg, *avd)
}

func ExecuteConfigureEmulatorDirect(cfg *config.Config, avdName string) error {
	avd, err := emulator.SelectAVD(cfg, avdName)
	if err != nil {
		return err
	}
	return emulator.OpenConfigInEditor(*avd)
}

func ExecuteRefreshDevices(cfg *config.Config) error {
	devices, err := adb.GetConnectedDevices(cfg.GetADBPath())
	if err != nil {
		return err
	}

	fmt.Printf("Connected devices: %d\n", len(devices))
	for i := range devices {
		// Load extended info for each device
		devices[i].LoadExtendedInfo(cfg.GetADBPath())

		formattedInfo := display.FormatExtendedInfoWithIndent(devices[i].String(), devices[i].GetExtendedInfo())
		fmt.Printf("  %s\n", formattedInfo)
	}
	return nil
}

func executeWiFiCommand(cfg *config.Config, args []string) error {
	if len(args) == 0 {
		// Show help when no subcommand provided
		fmt.Println("WiFi commands:")
		fmt.Println("  wifi pair <ip:port> <code>     - Pair with WiFi device")
		fmt.Println("  wifi connect <ip[:port]>       - Connect to WiFi device")
		fmt.Println("  wifi disconnect <ip[:port]>    - Disconnect from WiFi device")
		fmt.Println("")
		fmt.Println("Examples:")
		fmt.Println("  ./gadget wifi pair 192.168.1.100:5555 123456")
		fmt.Println("  ./gadget wifi connect 192.168.1.100")
		fmt.Println("  ./gadget wifi disconnect 192.168.1.100")
		return nil
	}

	subcommand := args[0]
	subArgs := args[1:]

	switch subcommand {
	case "pair":
		if len(subArgs) < 2 {
			return fmt.Errorf("wifi pair requires IP address and pairing code")
		}
		return commands.PairWiFiDevice(cfg, subArgs[0], subArgs[1])
	case "connect":
		if len(subArgs) < 1 {
			return fmt.Errorf("wifi connect requires IP address")
		}
		return commands.ConnectWiFi(cfg, subArgs[0])
	case "disconnect":
		if len(subArgs) < 1 {
			return fmt.Errorf("wifi disconnect requires IP address")
		}
		return commands.DisconnectWiFi(cfg, subArgs[0])
	default:
		return fmt.Errorf("unknown wifi subcommand: %s", subcommand)
	}
}
