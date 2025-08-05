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

// GetAvailableCommands returns the list of all available CLI commands
func GetAvailableCommands() []Command {
	return []Command{
		{"screenshot", "Screenshot", "Take a screenshot", "Media"},
		{"screenshot-day-night", "Screenshot day-night", "Take screenshots in day and night mode", "Media"},
		{"screen-record", "Screen record", "Record the screen", "Media"},
		{"dpi", "DPI", "View or change device DPI", "Device settings"},
		{"font-size", "Font size", "View or change device font size", "Device settings"},
		{"screen-size", "Screen size", "View or change device screen size", "Device settings"},
		{"wifi", "WiFi", "Manage WiFi device connections", "WiFi"},
		{"emulator", "Emulator", "Manage Android emulators", "Devices/emulators"},
		{"refresh-devices", "Refresh devices", "Refresh the device list", "Devices/emulators"},
	}
}

// GetTUICommands returns the list of commands for TUI display (more granular than CLI)
func GetTUICommands() []Command {
	return []Command{
		{"screenshot", "Screenshot", "Take a screenshot", "Media"},
		{"screenshot-day-night", "Screenshot day-night", "Take screenshots in day and night mode", "Media"},
		{"screen-record", "Screen record", "Record the screen", "Media"},
		{"dpi", "DPI", "View or change device DPI", "Device settings"},
		{"font-size", "Font size", "View or change device font size", "Device settings"},
		{"screen-size", "Screen size", "View or change device screen size", "Device settings"},
		{"pair-wifi", "Pair WiFi device", "Pair with a new WiFi device", "WiFi"},
		{"connect-wifi", "Connect WiFi device", "Connect to a WiFi device", "WiFi"},
		{"disconnect-wifi", "Disconnect WiFi device", "Disconnect from a WiFi device", "WiFi"},
		{"launch-emulator", "Launch emulator", "Start an Android emulator", "Devices/emulators"},
		{"configure-emulator", "Configure emulator", "Edit emulator configuration", "Devices/emulators"},
		{"refresh-devices", "Refresh devices", "Refresh the device list", "Devices/emulators"},
	}
}

// GetTUICommandCategories returns TUI commands grouped by category
func GetTUICommandCategories() []CommandCategory {
	commands := GetTUICommands()
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
