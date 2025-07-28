package registry

// Command represents a menu command with metadata
type Command struct {
	Command     string // kebab-case command name for CLI
	Name        string
	Description string
	Category    string
}

// CommandCategory represents a group of related commands
type CommandCategory struct {
	Name     string
	Commands []Command
}

// GetAvailableCommands returns the list of all available commands
func GetAvailableCommands() []Command {
	return []Command{
		{"screenshot", "Screenshot", "Take a screenshot", "Media"},
		{"screenshot-day-night", "Screenshot day-night", "Take screenshots in day and night mode", "Media"},
		{"screen-record", "Screen record", "Record the screen", "Media"},
		{"change-dpi", "Change DPI", "Change device DPI", "Device settings"},
		{"change-font-size", "Change font size", "Change device font size", "Device settings"},
		{"change-screen-size", "Change screen size", "Change device screen size", "Device settings"},
		{"pair-wifi", "Pair WiFi device", "Pair with a new WiFi device", "WiFi"},
		{"connect-wifi", "Connect WiFi device", "Connect to a WiFi device", "WiFi"},
		{"disconnect-wifi", "Disconnect WiFi device", "Disconnect from a WiFi device", "WiFi"},
		{"launch-emulator", "Launch emulator", "Start an Android emulator", "Devices/emulators"},
		{"configure-emulator", "Configure emulator", "Edit emulator configuration", "Devices/emulators"},
		{"refresh-devices", "Refresh devices", "Refresh the device list", "Devices/emulators"},
	}
}

// GetCommandCategories returns commands grouped by category
func GetCommandCategories() []CommandCategory {
	commands := GetAvailableCommands()
	categoryMap := make(map[string][]Command)

	// Group commands by category
	for _, cmd := range commands {
		categoryMap[cmd.Category] = append(categoryMap[cmd.Category], cmd)
	}

	// Return categories in desired order
	categoryOrder := []string{"Media", "Device settings", "WiFi", "Devices/emulators"}
	var categories []CommandCategory

	for _, categoryName := range categoryOrder {
		if cmds, exists := categoryMap[categoryName]; exists {
			categories = append(categories, CommandCategory{
				Name:     categoryName,
				Commands: cmds,
			})
		}
	}

	return categories
}

// GetAvailableCommandNames returns just the command names for CLI help
func GetAvailableCommandNames() []string {
	commands := GetAvailableCommands()
	names := make([]string, len(commands))
	for i, cmd := range commands {
		names[i] = cmd.Command
	}
	return names
}
