package test

import (
	"fmt"
	"gadget/internal/config"
	"os"
	"os/exec"
	"strings"
)

// FakedExecCommand represents a command response to fake
type FakedExecCommand struct {
	Command  string   // Command name (e.g., "adb", "/path/to/adb")
	Args     []string // Arguments to match
	Stdout   string   // Standard output to return
	Stderr   string   // Standard error to return
	ExitCode int      // Exit code to return (0 for success)
}

// ExecFaker provides command faking using the helper process pattern
type ExecFaker struct {
	fakes []FakedExecCommand
}

// NewExecFaker creates a new command faker
func NewExecFaker() *ExecFaker {
	return &ExecFaker{
		fakes: []FakedExecCommand{},
	}
}

// AddFake adds a command fake
func (f *ExecFaker) AddFake(command string, args []string, stdout, stderr string, exitCode int) {
	f.fakes = append(f.fakes, FakedExecCommand{
		Command:  command,
		Args:     args,
		Stdout:   stdout,
		Stderr:   stderr,
		ExitCode: exitCode,
	})
}

// FakeADBDevicesList fakes the "adb devices -l" command
func (f *ExecFaker) FakeADBDevicesList(response string) {
	// Fake both full path and just "adb" command
	f.AddFake("adb", []string{"devices", "-l"}, response, "", 0)
	// Also fake with potential full paths
	f.AddFake("/test/android/sdk/platform-tools/adb", []string{"devices", "-l"}, response, "", 0)
}

// FakeADBDevicesEmpty fakes empty device list
func (f *ExecFaker) FakeADBDevicesEmpty() {
	f.FakeADBDevicesList("List of devices attached\n")
}

// FakeADBDevicesSingle fakes single device response
func (f *ExecFaker) FakeADBDevicesSingle() {
	response := `List of devices attached
emulator-5554	device product:sdk_gphone64_x86_64 model:sdk_gphone64_x86_64 device:generic_x86_64 transport_id:1
`
	f.FakeADBDevicesList(response)
}

// FakeADBDevicesMultiple fakes multiple devices response
func (f *ExecFaker) FakeADBDevicesMultiple() {
	response := `List of devices attached
emulator-5554	device product:sdk_gphone64_x86_64 model:sdk_gphone64_x86_64 device:generic_x86_64 transport_id:1
192.168.1.100:5555	device product:OnePlus7Pro model:GM1913 device:OnePlus7Pro transport_id:2
`
	f.FakeADBDevicesList(response)
}

// FakeADBError fakes ADB command failure
func (f *ExecFaker) FakeADBError() {
	f.AddFake("adb", []string{"devices", "-l"}, "", "adb: command not found", 1)
	f.AddFake("/test/android/sdk/platform-tools/adb", []string{"devices", "-l"}, "", "adb: command not found", 1)
}

// FakeExecCommand returns a fake exec.Cmd that will run our test helper
func (f *ExecFaker) FakeExecCommand(command string, args ...string) *exec.Cmd {
	cs := []string{"-test.run=TestHelperProcess", "--", command}
	cs = append(cs, args...)
	cmd := exec.Command(os.Args[0], cs...)

	// Pass the fake data via environment variables
	env := []string{"GO_TEST_HELPER_PROCESS=1"}

	// Find matching fake and encode it in environment
	for _, fake := range f.fakes {
		if f.commandMatches(fake, command, args) {
			env = append(env, fmt.Sprintf("TEST_STDOUT=%s", fake.Stdout))
			env = append(env, fmt.Sprintf("TEST_STDERR=%s", fake.Stderr))
			env = append(env, fmt.Sprintf("TEST_EXIT_CODE=%d", fake.ExitCode))
			break
		}
	}

	cmd.Env = env
	return cmd
}

// commandMatches checks if a command and args match a fake
func (f *ExecFaker) commandMatches(fake FakedExecCommand, command string, args []string) bool {
	// Check if command ends with fake command (handles full paths)
	if !strings.HasSuffix(command, fake.Command) && command != fake.Command {
		return false
	}

	// Check args match exactly
	if len(args) != len(fake.Args) {
		return false
	}

	for i, arg := range args {
		if arg != fake.Args[i] {
			return false
		}
	}

	return true
}

// TestConfig creates a config suitable for testing
func TestConfig() *config.Config {
	return &config.Config{
		AndroidHome: "/test/android/sdk",
		MediaPath:   "/test/media",
	}
}
