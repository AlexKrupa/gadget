package main

import (
	"flag"
	"fmt"
	"gadget/internal/cli"
	"gadget/internal/config"
	"gadget/internal/registry"
	"gadget/internal/tui"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	cfg := config.NewConfig()

	// Check if adb exists
	if _, err := os.Stat(cfg.GetADBPath()); os.IsNotExist(err) {
		fmt.Printf("Error: ADB not found at %s\n", cfg.GetADBPath())
		fmt.Printf("Please check your ANDROID_HOME environment variable: %s\n", cfg.AndroidHome)
		os.Exit(1)
	}

	availableCommands := registry.GetAvailableCommandNames()
	commandHelp := fmt.Sprintf("Command to execute directly (%s)", strings.Join(availableCommands, ", "))

	command := flag.String("command", "", commandHelp)
	deviceSerial := flag.String("device", "", "Device serial for device-specific commands")
	ip := flag.String("ip", "", "IP address for WiFi commands")
	code := flag.String("code", "", "Pairing code for WiFi pairing")
	value := flag.String("value", "", "Value for setting commands (DPI, font size, screen size)")
	flag.Parse()

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
		model := tui.NewModel(cfg)
		program := tea.NewProgram(model, tea.WithAltScreen())

		if _, err := program.Run(); err != nil {
			fmt.Printf("Error running program: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Check if this is a nested command
	if isNestedCommand(cmdToExecute) {
		if err := cli.ExecuteNestedCommand(cfg, cmdToExecute, args); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
	} else {
		parsedArgs := parsePositionalArgs(cmdToExecute, args, *deviceSerial, *ip, *code, *value)
		if err := executeDirectCommand(cfg, cmdToExecute, parsedArgs.device, parsedArgs.ip, parsedArgs.code, parsedArgs.value); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
	}
}

// ParsedArgs holds the parsed command arguments
type ParsedArgs struct {
	device string
	ip     string
	code   string
	value  string
}

// ArgumentParser defines how to parse arguments for a specific command
type ArgumentParser func(args []string, flags ParsedArgs) ParsedArgs

var argumentParsers = map[string]ArgumentParser{
	"pair-wifi":            parsePairWiFiArgs,
	"connect-wifi":         parseWiFiArgs,
	"disconnect-wifi":      parseWiFiArgs,
	"dpi":                  parseSettingArgs,
	"font-size":            parseSettingArgs,
	"screen-size":          parseSettingArgs,
	"launch-emulator":      parseValueArgs,
	"configure-emulator":   parseValueArgs,
	"screenshot":           parseDeviceArgs,
	"screenshot-day-night": parseDeviceArgs,
	"screen-record":        parseDeviceArgs,
}

// parsePositionalArgs parses positional arguments based on command type
func parsePositionalArgs(command string, args []string, flagDevice, flagIP, flagCode, flagValue string) ParsedArgs {
	flags := ParsedArgs{
		device: flagDevice,
		ip:     flagIP,
		code:   flagCode,
		value:  flagValue,
	}

	parser, exists := argumentParsers[command]
	if !exists {
		return flags // Return flags as-is for unknown commands
	}

	return parser(args, flags)
}

// isNestedCommand checks if a command is a nested command by checking the registry
func isNestedCommand(command string) bool {
	_, exists := cli.NestedCommandRegistry[command]
	return exists
}

func parsePairWiFiArgs(args []string, flags ParsedArgs) ParsedArgs {
	result := flags
	if len(args) >= 1 && result.ip == "" {
		result.ip = args[0]
	}
	if len(args) >= 2 && result.code == "" {
		result.code = args[1]
	}
	return result
}

func parseWiFiArgs(args []string, flags ParsedArgs) ParsedArgs {
	result := flags
	if len(args) >= 1 && result.ip == "" {
		result.ip = args[0]
	}
	return result
}

func parseSettingArgs(args []string, flags ParsedArgs) ParsedArgs {
	result := flags
	if len(args) >= 1 && result.value == "" {
		if result.device == "" && (strings.Contains(args[0], "emulator") || strings.Contains(args[0], ":") || len(args) > 1) {
			result.device = args[0]
			if len(args) >= 2 {
				result.value = args[1]
			}
		} else {
			result.value = args[0]
		}
	}
	return result
}

func parseValueArgs(args []string, flags ParsedArgs) ParsedArgs {
	result := flags
	if len(args) >= 1 && result.value == "" {
		result.value = args[0]
	}
	return result
}

func parseDeviceArgs(args []string, flags ParsedArgs) ParsedArgs {
	result := flags
	if len(args) >= 1 && result.device == "" {
		result.device = args[0]
	}
	return result
}

// executeDirectCommand executes a command directly without the TUI
func executeDirectCommand(cfg *config.Config, command, deviceSerial, ip, code, value string) error {
	return cli.ExecuteCommand(cfg, command, deviceSerial, ip, code, value)
}
