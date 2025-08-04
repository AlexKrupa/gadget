package core

import "gadget/internal/registry"

// Delegate to registry package for command definitions
func GetAvailableCommands() []Command {
	registryCommands := registry.GetAvailableCommands()
	commands := make([]Command, len(registryCommands))
	for i, cmd := range registryCommands {
		commands[i] = Command{
			Command:     cmd.Command,
			Name:        cmd.Name,
			Description: cmd.Description,
			Category:    cmd.Category,
		}
	}
	return commands
}

func GetCommandCategories() []CommandCategory {
	registryCategories := registry.GetCommandCategories()
	categories := make([]CommandCategory, len(registryCategories))
	for i, cat := range registryCategories {
		commands := make([]Command, len(cat.Commands))
		for j, cmd := range cat.Commands {
			commands[j] = Command{
				Command:     cmd.Command,
				Name:        cmd.Name,
				Description: cmd.Description,
				Category:    cmd.Category,
			}
		}
		categories[i] = CommandCategory{
			Name:     cat.Name,
			Commands: commands,
		}
	}
	return categories
}
