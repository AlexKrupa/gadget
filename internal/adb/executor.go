package adb

import (
	"os/exec"
)

// CommandExecutor interface allows dependency injection for testing
type CommandExecutor interface {
	Command(name string, arg ...string) *exec.Cmd
}

// RealCommandExecutor is the production implementation using exec.Command
type RealCommandExecutor struct{}

func (r *RealCommandExecutor) Command(name string, arg ...string) *exec.Cmd {
	return exec.Command(name, arg...)
}

// Global executor that can be replaced in tests
var globalExecutor CommandExecutor = &RealCommandExecutor{}

// SetCommandExecutor allows tests to inject a fake executor
func SetCommandExecutor(executor CommandExecutor) {
	globalExecutor = executor
}

// ResetCommandExecutor resets to the default real executor
func ResetCommandExecutor() {
	globalExecutor = &RealCommandExecutor{}
}

// execCommand is a wrapper that uses the global executor
func execCommand(name string, arg ...string) *exec.Cmd {
	return globalExecutor.Command(name, arg...)
}
