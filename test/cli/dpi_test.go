package test

import (
	"gadget/internal/cli"
	"gadget/internal/config"
	"gadget/test/cli/util"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDPICommand(t *testing.T) {
	tests := []struct {
		name             string
		setupStubs       func(*util.GenericExecFaker, *config.Config)
		command          string
		deviceSerial     string
		value            string
		expectedOutput   []string
		expectedError    string
		expectedCommands []string // Expected command patterns that should be executed
	}{
		{
			name: "get current DPI - physical only",
			setupStubs: func(f *util.GenericExecFaker, cfg *config.Config) {
				f.StubSingleDevice(cfg.GetADBPath())
				f.StubDPIGet(cfg.GetADBPath(), "emulator-5554", "Physical density: 420")
			},
			command:          "dpi",
			expectedOutput:   []string{"Physical DPI: 420", "Current DPI: 420"},
			expectedCommands: []string{"adb devices -l", "adb -s emulator-5554 shell wm density"},
		},
		{
			name: "get current DPI - with override",
			setupStubs: func(f *util.GenericExecFaker, cfg *config.Config) {
				f.StubSingleDevice(cfg.GetADBPath())
				response := `Physical density: 420
Override density: 480`
				f.StubDPIGet(cfg.GetADBPath(), "emulator-5554", response)
			},
			command:          "dpi",
			expectedOutput:   []string{"Physical DPI: 420", "Current DPI: 480"},
			expectedCommands: []string{"adb devices -l", "adb -s emulator-5554 shell wm density"},
		},
		{
			name: "set DPI successfully",
			setupStubs: func(f *util.GenericExecFaker, cfg *config.Config) {
				f.StubSingleDevice(cfg.GetADBPath())
				f.StubDPISet(cfg.GetADBPath(), "emulator-5554", "480", 0)
				// Also stub the follow-up get to show new values
				f.StubDPIGet(cfg.GetADBPath(), "emulator-5554", "Physical density: 420\nOverride density: 480")
			},
			command:          "dpi",
			value:            "480",
			expectedOutput:   []string{"Physical DPI: 420", "Current DPI: 480"},
			expectedCommands: []string{"adb devices -l", "adb -s emulator-5554 shell wm density 480", "adb -s emulator-5554 shell wm density"},
		},
		{
			name: "set DPI with device serial",
			setupStubs: func(f *util.GenericExecFaker, cfg *config.Config) {
				// Stub multiple devices to test device selection
				multiDeviceResponse := `List of devices attached
emulator-5554	device product:sdk_gphone64_x86_64 model:sdk_gphone64_x86_64 device:generic_x86_64 transport_id:1
192.168.1.100:5555	device product:OnePlus7Pro model:GM1913 device:OnePlus7Pro transport_id:2
`
				f.StubADBDevicesCommand(cfg.GetADBPath(), multiDeviceResponse)
				f.StubDPISet(cfg.GetADBPath(), "192.168.1.100:5555", "320", 0)
				f.StubDPIGet(cfg.GetADBPath(), "192.168.1.100:5555", "Physical density: 280\nOverride density: 320")
			},
			command:          "dpi",
			deviceSerial:     "192.168.1.100:5555",
			value:            "320",
			expectedOutput:   []string{"Physical DPI: 280", "Current DPI: 320"},
			expectedCommands: []string{"adb devices -l", "adb -s 192.168.1.100:5555 shell wm density 320", "adb -s 192.168.1.100:5555 shell wm density"},
		},
		{
			name: "get DPI fails - adb command error",
			setupStubs: func(f *util.GenericExecFaker, cfg *config.Config) {
				f.StubSingleDevice(cfg.GetADBPath())
				f.StubADBShellCommand(cfg.GetADBPath(), "emulator-5554", []string{"wm", "density"}, "", "wm: command not found", 1)
			},
			command:          "dpi",
			expectedError:    "failed to get current DPI",
			expectedCommands: []string{"adb devices -l", "adb -s emulator-5554 shell wm density"},
		},
		{
			name: "set DPI fails - adb command error",
			setupStubs: func(f *util.GenericExecFaker, cfg *config.Config) {
				f.StubSingleDevice(cfg.GetADBPath())
				f.StubDPISet(cfg.GetADBPath(), "emulator-5554", "600", 1)
			},
			command:          "dpi",
			value:            "600",
			expectedError:    "failed to set DPI to 600",
			expectedCommands: []string{"adb devices -l", "adb -s emulator-5554 shell wm density 600"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			faker := util.NewGenericExecFaker()
			cfg := util.TestConfig()
			tt.setupStubs(faker, cfg)

			var cmdError error
			output := util.CaptureLogOutput(func() {
				util.WithFakeExec(faker, func() {
					cmdError = cli.ExecuteCommand(cfg, tt.command, tt.deviceSerial, "", "", tt.value)
				})
			})

			// Check error expectations
			if tt.expectedError != "" {
				require.Error(t, cmdError)
				assert.Contains(t, cmdError.Error(), tt.expectedError)
			} else {
				require.NoError(t, cmdError)
			}

			// Check output expectations
			for _, expectedOut := range tt.expectedOutput {
				assert.Contains(t, output, expectedOut, "Expected output not found")
			}

			// Verify expected commands were executed
			executedCommands := faker.GetExecutedCommands()
			for _, expectedCmd := range tt.expectedCommands {
				found := false
				for _, executed := range executedCommands {
					if util.MatchesCommandPattern(executed, expectedCmd) {
						found = true
						break
					}
				}
				assert.True(t, found, "Expected command not executed: %s\nActual commands: %v", expectedCmd, util.FormatExecutedCommands(executedCommands))
			}
		})
	}
}
