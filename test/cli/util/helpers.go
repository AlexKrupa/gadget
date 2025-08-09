package util

import (
	"gadget/internal/adb"
	"gadget/internal/config"
)

// WithFakeExec runs a function with a fake command executor
func WithFakeExec(faker *GenericExecFaker, fn func()) {
	// Inject the fake executor into the adb package
	testExecutor := NewTestCommandExecutor(faker)
	adb.SetCommandExecutor(testExecutor)
	defer adb.ResetCommandExecutor()

	fn()
}

// MatchesCommandPattern checks if an executed command matches the expected pattern
func MatchesCommandPattern(executed ExecutionRecord, pattern string) bool {
	// For simplicity, convert to full command string and check if pattern matches
	fullCommand := executed.Command
	if len(executed.Args) > 0 {
		fullCommand += " " + joinArgs(executed.Args)
	}

	// Check if the pattern is contained in the full command
	return containsPattern(fullCommand, pattern)
}

// containsPattern checks if a command contains the expected pattern
func containsPattern(fullCommand, pattern string) bool {
	return len(fullCommand) >= len(pattern) &&
		(fullCommand == pattern ||
			fullCommand[len(fullCommand)-len(pattern):] == pattern ||
			fullCommand[:len(pattern)] == pattern ||
			containsSubstring(fullCommand, pattern))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// joinArgs joins arguments with spaces
func joinArgs(args []string) string {
	if len(args) == 0 {
		return ""
	}
	result := args[0]
	for i := 1; i < len(args); i++ {
		result += " " + args[i]
	}
	return result
}

// FormatExecutedCommands formats executed commands for error messages
func FormatExecutedCommands(commands []ExecutionRecord) []string {
	var formatted []string
	for _, cmd := range commands {
		fullCmd := cmd.Command
		if len(cmd.Args) > 0 {
			fullCmd += " " + joinArgs(cmd.Args)
		}
		formatted = append(formatted, fullCmd)
	}
	return formatted
}

// TestConfig creates a config suitable for testing
func TestConfig() *config.Config {
	return &config.Config{
		AndroidHome: "/test/android/sdk",
		MediaPath:   "/test/media",
	}
}
