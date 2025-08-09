package util

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// CommandStub represents a stubbed command with its expected response
type CommandStub struct {
	Command  string   // Full command path (e.g., "/test/android/sdk/platform-tools/adb")
	Args     []string // Arguments to match exactly
	Stdout   string   // Standard output to return
	Stderr   string   // Standard error to return
	ExitCode int      // Exit code to return (0 for success)
}

// ExecutionRecord tracks commands that were executed during tests
type ExecutionRecord struct {
	Command string   // Full command path
	Args    []string // Arguments passed
}

// GenericExecFaker provides generic exec.Command faking using the helper process pattern
type GenericExecFaker struct {
	stubs            []CommandStub
	executedCommands []ExecutionRecord
}

// NewGenericExecFaker creates a new generic command faker
func NewGenericExecFaker() *GenericExecFaker {
	return &GenericExecFaker{
		stubs:            []CommandStub{},
		executedCommands: []ExecutionRecord{},
	}
}

// AddStub adds a command stub for exact command and args matching
func (f *GenericExecFaker) AddStub(command string, args []string, stdout, stderr string, exitCode int) {
	f.stubs = append(f.stubs, CommandStub{
		Command:  command,
		Args:     args,
		Stdout:   stdout,
		Stderr:   stderr,
		ExitCode: exitCode,
	})
}

// GetExecutedCommands returns all commands that were executed during the test
func (f *GenericExecFaker) GetExecutedCommands() []ExecutionRecord {
	return f.executedCommands
}

// GetStubs returns all configured stubs (for debugging)
func (f *GenericExecFaker) GetStubs() []CommandStub {
	return f.stubs
}

// FindExecutedCommand finds the first executed command matching the pattern
func (f *GenericExecFaker) FindExecutedCommand(commandSuffix string, args ...string) *ExecutionRecord {
	for _, record := range f.executedCommands {
		if strings.HasSuffix(record.Command, commandSuffix) {
			if len(args) == 0 || f.argsMatch(record.Args, args) {
				return &record
			}
		}
	}
	return nil
}

// FakeExecCommand returns a fake exec.Cmd that will run our test helper
func (f *GenericExecFaker) FakeExecCommand(command string, args ...string) *exec.Cmd {
	// Record this command execution
	f.executedCommands = append(f.executedCommands, ExecutionRecord{
		Command: command,
		Args:    args,
	})

	cs := []string{"-test.run=TestHelperProcess", "--", command}
	cs = append(cs, args...)
	cmd := exec.Command(os.Args[0], cs...)

	// Pass the fake data via environment variables
	env := []string{"GO_TEST_HELPER_PROCESS=1"}

	// Find matching stub and encode it in environment
	for _, stub := range f.stubs {
		if f.commandMatches(stub, command, args) {
			env = append(env, fmt.Sprintf("TEST_STDOUT=%s", stub.Stdout))
			env = append(env, fmt.Sprintf("TEST_STDERR=%s", stub.Stderr))
			env = append(env, fmt.Sprintf("TEST_EXIT_CODE=%d", stub.ExitCode))
			break
		}
	}

	cmd.Env = env
	return cmd
}

// commandMatches checks if a command and args match a stub exactly
func (f *GenericExecFaker) commandMatches(stub CommandStub, command string, args []string) bool {
	// Check if command ends with stub command (handles full paths)
	if !strings.HasSuffix(command, stub.Command) && command != stub.Command {
		return false
	}

	return f.argsMatch(args, stub.Args)
}

// argsMatch checks if two argument lists match exactly
func (f *GenericExecFaker) argsMatch(actual, expected []string) bool {
	if len(actual) != len(expected) {
		return false
	}

	for i, arg := range actual {
		if arg != expected[i] {
			return false
		}
	}

	return true
}

// Helper methods for common command patterns

// StubADBDevicesCommand stubs the "adb devices -l" command
func (f *GenericExecFaker) StubADBDevicesCommand(adbPath, response string) {
	f.AddStub(adbPath, []string{"devices", "-l"}, response, "", 0)
}

// StubADBShellCommand stubs an "adb -s <device> shell <command>"
func (f *GenericExecFaker) StubADBShellCommand(adbPath, deviceSerial string, shellCommand []string, stdout, stderr string, exitCode int) {
	args := []string{"-s", deviceSerial, "shell"}
	args = append(args, shellCommand...)
	f.AddStub(adbPath, args, stdout, stderr, exitCode)
}

// Convenience methods for common ADB commands

// StubSingleDevice stubs adb devices to return a single emulator
func (f *GenericExecFaker) StubSingleDevice(adbPath string) {
	response := `List of devices attached
emulator-5554	device product:sdk_gphone64_x86_64 model:sdk_gphone64_x86_64 device:generic_x86_64 transport_id:1
`
	f.StubADBDevicesCommand(adbPath, response)
}

// StubDPIGet stubs getting DPI with specific response
func (f *GenericExecFaker) StubDPIGet(adbPath, deviceSerial, response string) {
	f.StubADBShellCommand(adbPath, deviceSerial, []string{"wm", "density"}, response, "", 0)
}

// StubDPISet stubs setting DPI
func (f *GenericExecFaker) StubDPISet(adbPath, deviceSerial, dpiValue string, exitCode int) {
	f.StubADBShellCommand(adbPath, deviceSerial, []string{"wm", "density", dpiValue}, "", "", exitCode)
}

// StubScreenSizeGet stubs getting screen size with specific response
func (f *GenericExecFaker) StubScreenSizeGet(adbPath, deviceSerial, response string) {
	f.StubADBShellCommand(adbPath, deviceSerial, []string{"wm", "size"}, response, "", 0)
}

// StubScreenSizeSet stubs setting screen size
func (f *GenericExecFaker) StubScreenSizeSet(adbPath, deviceSerial, sizeValue string, exitCode int) {
	f.StubADBShellCommand(adbPath, deviceSerial, []string{"wm", "size", sizeValue}, "", "", exitCode)
}

// StubFontSizeGet stubs getting font size with specific response
func (f *GenericExecFaker) StubFontSizeGet(adbPath, deviceSerial, response string) {
	f.StubADBShellCommand(adbPath, deviceSerial, []string{"settings", "get", "system", "font_scale"}, response, "", 0)
}

// StubFontSizeSet stubs setting font size
func (f *GenericExecFaker) StubFontSizeSet(adbPath, deviceSerial, scaleValue string, exitCode int) {
	f.StubADBShellCommand(adbPath, deviceSerial, []string{"settings", "put", "system", "font_scale", scaleValue}, "", "", exitCode)
}

// Additional convenience methods for compatibility with legacy tests

// StubMultipleDevices stubs adb devices to return multiple devices
func (f *GenericExecFaker) StubMultipleDevices(adbPath string) {
	response := `List of devices attached
emulator-5554	device product:sdk_gphone64_x86_64 model:sdk_gphone64_x86_64 device:generic_x86_64 transport_id:1
192.168.1.100:5555	device product:OnePlus7Pro model:GM1913 device:OnePlus7Pro transport_id:2
`
	f.StubADBDevicesCommand(adbPath, response)
}

// StubEmptyDevices stubs adb devices to return no devices
func (f *GenericExecFaker) StubEmptyDevices(adbPath string) {
	response := "List of devices attached\n"
	f.StubADBDevicesCommand(adbPath, response)
}

// StubADBError stubs adb devices command to fail
func (f *GenericExecFaker) StubADBError(adbPath string) {
	f.AddStub(adbPath, []string{"devices", "-l"}, "", "adb: command not found", 1)
}

// TestCommandExecutor implements the CommandExecutor interface for testing
type TestCommandExecutor struct {
	faker *GenericExecFaker
}

// NewTestCommandExecutor creates a new test command executor
func NewTestCommandExecutor(faker *GenericExecFaker) *TestCommandExecutor {
	return &TestCommandExecutor{faker: faker}
}

// Command implements the CommandExecutor interface
func (t *TestCommandExecutor) Command(name string, arg ...string) *exec.Cmd {
	return t.faker.FakeExecCommand(name, arg...)
}
