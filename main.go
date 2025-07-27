package main

import (
	"adx/internal/adb"
	"adx/internal/commands"
	"adx/internal/config"
	"adx/internal/emulator"
	"adx/internal/tui"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	// Create configuration
	cfg := config.NewConfig()

	// Check if adb exists
	if _, err := os.Stat(cfg.GetADBPath()); os.IsNotExist(err) {
		fmt.Printf("Error: ADB not found at %s\n", cfg.GetADBPath())
		fmt.Printf("Please check your ANDROID_HOME environment variable: %s\n", cfg.AndroidHome)
		os.Exit(1)
	}

	// Get available commands for help text
	availableCommands := tui.GetAvailableCommandNames()
	commandHelp := fmt.Sprintf("Command to execute directly (%s)", strings.Join(availableCommands, ", "))

	// Parse command line arguments
	command := flag.String("command", "", commandHelp)
	deviceSerial := flag.String("device", "", "Device serial for device-specific commands")
	ip := flag.String("ip", "", "IP address for WiFi commands")
	code := flag.String("code", "", "Pairing code for WiFi pairing")
	value := flag.String("value", "", "Value for setting commands (DPI, font size, screen size)")
	flag.Parse()

	// Get remaining arguments after flags
	args := flag.Args()

	// Determine command from either flag or first positional argument
	var cmdToExecute string
	if *command != "" {
		cmdToExecute = *command
	} else if len(args) > 0 {
		cmdToExecute = args[0]
		args = args[1:] // Remove command from args
	}

	// If no command specified, start TUI
	if cmdToExecute == "" {
		// Create and start the TUI
		model := tui.NewModel(cfg)
		program := tea.NewProgram(model, tea.WithAltScreen())

		if _, err := program.Run(); err != nil {
			fmt.Printf("Error running program: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Parse positional arguments based on command
	parsedArgs := parsePositionalArgs(cmdToExecute, args, *deviceSerial, *ip, *code, *value)

	// Execute direct command
	if err := executeDirectCommand(cfg, cmdToExecute, parsedArgs.device, parsedArgs.ip, parsedArgs.code, parsedArgs.value); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

// ParsedArgs holds the parsed command arguments
type ParsedArgs struct {
	device string
	ip     string
	code   string
	value  string
}

// parsePositionalArgs parses positional arguments based on command type
func parsePositionalArgs(command string, args []string, flagDevice, flagIP, flagCode, flagValue string) ParsedArgs {
	result := ParsedArgs{
		device: flagDevice,
		ip:     flagIP,
		code:   flagCode,
		value:  flagValue,
	}

	switch command {
	case "pair-wifi":
		// pair-wifi [ip] [code]
		if len(args) >= 1 && result.ip == "" {
			result.ip = args[0]
		}
		if len(args) >= 2 && result.code == "" {
			result.code = args[1]
		}
	case "connect-wifi", "disconnect-wifi":
		// connect-wifi [ip], disconnect-wifi [ip]
		if len(args) >= 1 && result.ip == "" {
			result.ip = args[0]
		}
	case "change-dpi", "change-font-size", "change-screen-size":
		// change-* [device] [value] or change-* [value]
		if len(args) >= 1 && result.value == "" {
			// If it looks like a device serial (contains letters/numbers), treat as device
			// Otherwise treat as value
			if result.device == "" && (strings.Contains(args[0], "emulator") || strings.Contains(args[0], ":") || len(args) > 1) {
				result.device = args[0]
				if len(args) >= 2 {
					result.value = args[1]
				}
			} else {
				result.value = args[0]
			}
		}
	case "launch-emulator":
		// launch-emulator [avd-name]
		if len(args) >= 1 && result.value == "" {
			result.value = args[0]
		}
	case "screenshot", "screenshot-day-night", "screen-record":
		// screenshot [device]
		if len(args) >= 1 && result.device == "" {
			result.device = args[0]
		}
	}

	return result
}

// executeDirectCommand executes a command directly without the TUI
func executeDirectCommand(cfg *config.Config, command, deviceSerial, ip, code, value string) error {
	switch command {
	case "screenshot":
		return executeScreenshotDirect(cfg, deviceSerial)
	case "screenshot-day-night":
		return executeScreenshotDayNightDirect(cfg, deviceSerial)
	case "screen-record":
		return executeScreenRecordDirect(cfg, deviceSerial)
	case "change-dpi":
		return executeChangeDPIDirect(cfg, deviceSerial, value)
	case "change-font-size":
		return executeChangeFontSizeDirect(cfg, deviceSerial, value)
	case "change-screen-size":
		return executeChangeScreenSizeDirect(cfg, deviceSerial, value)
	case "launch-emulator":
		return executeLaunchEmulatorDirect(cfg, value)
	case "pair-wifi":
		if ip == "" || code == "" {
			return fmt.Errorf("pair-wifi requires IP address and pairing code")
		}
		return commands.PairWiFiDevice(cfg, ip, code)
	case "connect-wifi":
		if ip == "" {
			return fmt.Errorf("connect-wifi requires IP address")
		}
		return commands.ConnectWiFi(cfg, ip)
	case "disconnect-wifi":
		if ip == "" {
			return fmt.Errorf("disconnect-wifi requires IP address")
		}
		return commands.DisconnectWiFi(cfg, ip)
	case "refresh-devices":
		devices, err := adb.GetConnectedDevices(cfg.GetADBPath())
		if err != nil {
			return err
		}
		fmt.Printf("Connected devices: %d\n", len(devices))
		for _, device := range devices {
			fmt.Printf("  %s\n", device.String())
		}
		return nil
	default:
		return fmt.Errorf("unknown command: %s", command)
	}
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

func executeScreenshotDirect(cfg *config.Config, deviceSerial string) error {
	device, err := selectDevice(cfg, deviceSerial)
	if err != nil {
		return err
	}

	fmt.Printf("Taking screenshot on device: %s\n", device.Serial)
	return commands.TakeScreenshot(cfg, device)
}

func executeScreenshotDayNightDirect(cfg *config.Config, deviceSerial string) error {
	device, err := selectDevice(cfg, deviceSerial)
	if err != nil {
		return err
	}

	fmt.Printf("Taking day-night screenshots on device: %s\n", device.Serial)
	return commands.TakeDayNightScreenshots(cfg, device)
}

func executeScreenRecordDirect(cfg *config.Config, deviceSerial string) error {
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

func executeChangeDPIDirect(cfg *config.Config, deviceSerial, value string) error {
	if value == "" {
		return fmt.Errorf("change-dpi requires -value (DPI number)")
	}

	device, err := selectDevice(cfg, deviceSerial)
	if err != nil {
		return err
	}

	handler := commands.GetSettingHandler(commands.SettingTypeDPI)
	if err := handler.ValidateInput(value); err != nil {
		return err
	}

	fmt.Printf("Changing DPI to %s on device: %s\n", value, device.Serial)
	return handler.SetValue(cfg, device, value)
}

func executeChangeFontSizeDirect(cfg *config.Config, deviceSerial, value string) error {
	if value == "" {
		return fmt.Errorf("change-font-size requires -value (font scale number)")
	}

	device, err := selectDevice(cfg, deviceSerial)
	if err != nil {
		return err
	}

	handler := commands.GetSettingHandler(commands.SettingTypeFontSize)
	if err := handler.ValidateInput(value); err != nil {
		return err
	}

	fmt.Printf("Changing font size to %s on device: %s\n", value, device.Serial)
	return handler.SetValue(cfg, device, value)
}

func executeChangeScreenSizeDirect(cfg *config.Config, deviceSerial, value string) error {
	if value == "" {
		return fmt.Errorf("change-screen-size requires -value (WIDTHxHEIGHT)")
	}

	device, err := selectDevice(cfg, deviceSerial)
	if err != nil {
		return err
	}

	handler := commands.GetSettingHandler(commands.SettingTypeScreenSize)
	if err := handler.ValidateInput(value); err != nil {
		return err
	}

	fmt.Printf("Changing screen size to %s on device: %s\n", value, device.Serial)
	return handler.SetValue(cfg, device, value)
}

func executeLaunchEmulatorDirect(cfg *config.Config, avdName string) error {
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
