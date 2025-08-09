package test

import (
	"gadget/internal/cli"
	"gadget/internal/config"
	"gadget/test/cli/util"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScreenSizeCommand(t *testing.T) {
	tests := []struct {
		name             string
		setupStubs       func(*util.GenericExecFaker, *config.Config)
		command          string
		deviceSerial     string
		value            string
		expectedOutput   []string
		expectedError    string
		expectedCommands []string
	}{
		{
			name: "get current screen size - physical only",
			setupStubs: func(f *util.GenericExecFaker, cfg *config.Config) {
				f.StubSingleDevice(cfg.GetADBPath())
				f.StubScreenSizeGet(cfg.GetADBPath(), "emulator-5554", "Physical size: 1080x1920")
			},
			command:          "screen-size",
			expectedOutput:   []string{"Physical screen size: 1080x1920", "Current screen size: 1080x1920"},
			expectedCommands: []string{"adb devices -l", "adb -s emulator-5554 shell wm size"},
		},
		{
			name: "get current screen size - with override",
			setupStubs: func(f *util.GenericExecFaker, cfg *config.Config) {
				f.StubSingleDevice(cfg.GetADBPath())
				response := `Physical size: 1080x1920
Override size: 1080x1800`
				f.StubScreenSizeGet(cfg.GetADBPath(), "emulator-5554", response)
			},
			command:          "screen-size",
			expectedOutput:   []string{"Physical screen size: 1080x1920", "Current screen size: 1080x1800"},
			expectedCommands: []string{"adb devices -l", "adb -s emulator-5554 shell wm size"},
		},
		{
			name: "set screen size successfully",
			setupStubs: func(f *util.GenericExecFaker, cfg *config.Config) {
				f.StubSingleDevice(cfg.GetADBPath())
				f.StubScreenSizeSet(cfg.GetADBPath(), "emulator-5554", "1080x1800", 0)
				// Also stub the follow-up get to show new values
				f.StubScreenSizeGet(cfg.GetADBPath(), "emulator-5554", "Physical size: 1080x1920\nOverride size: 1080x1800")
			},
			command:          "screen-size",
			value:            "1080x1800",
			expectedOutput:   []string{"Physical screen size: 1080x1920", "Current screen size: 1080x1800"},
			expectedCommands: []string{"adb devices -l", "adb -s emulator-5554 shell wm size 1080x1800", "adb -s emulator-5554 shell wm size"},
		},
		{
			name: "get screen size fails - adb command error",
			setupStubs: func(f *util.GenericExecFaker, cfg *config.Config) {
				f.StubSingleDevice(cfg.GetADBPath())
				f.StubADBShellCommand(cfg.GetADBPath(), "emulator-5554", []string{"wm", "size"}, "", "wm: command not found", 1)
			},
			command:          "screen-size",
			expectedError:    "failed to get current screen size",
			expectedCommands: []string{"adb devices -l", "adb -s emulator-5554 shell wm size"},
		},
		{
			name: "set screen size fails - adb command error",
			setupStubs: func(f *util.GenericExecFaker, cfg *config.Config) {
				f.StubSingleDevice(cfg.GetADBPath())
				f.StubScreenSizeSet(cfg.GetADBPath(), "emulator-5554", "1080x1600", 1)
			},
			command:          "screen-size",
			value:            "1080x1600",
			expectedError:    "failed to set screen size to 1080x1600",
			expectedCommands: []string{"adb devices -l", "adb -s emulator-5554 shell wm size 1080x1600"},
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
