package test

import (
	"gadget/internal/cli"
	"gadget/internal/config"
	"gadget/test/cli/util"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFontSizeCommand(t *testing.T) {
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
			name: "get current font size - default",
			setupStubs: func(f *util.GenericExecFaker, cfg *config.Config) {
				f.StubSingleDevice(cfg.GetADBPath())
				f.StubFontSizeGet(cfg.GetADBPath(), "emulator-5554", "null")
			},
			command:          "font-size",
			expectedOutput:   []string{"Default font size: 1.0", "Current font size: 1.0"},
			expectedCommands: []string{"adb devices -l", "adb -s emulator-5554 shell settings get system font_scale"},
		},
		{
			name: "get current font size - custom scale",
			setupStubs: func(f *util.GenericExecFaker, cfg *config.Config) {
				f.StubSingleDevice(cfg.GetADBPath())
				f.StubFontSizeGet(cfg.GetADBPath(), "emulator-5554", "1.5")
			},
			command:          "font-size",
			expectedOutput:   []string{"Default font size: 1.0", "Current font size: 1.5"},
			expectedCommands: []string{"adb devices -l", "adb -s emulator-5554 shell settings get system font_scale"},
		},
		{
			name: "set font size successfully",
			setupStubs: func(f *util.GenericExecFaker, cfg *config.Config) {
				f.StubSingleDevice(cfg.GetADBPath())
				f.StubFontSizeSet(cfg.GetADBPath(), "emulator-5554", "1.2", 0)
				// Also stub the follow-up get to show new values
				f.StubFontSizeGet(cfg.GetADBPath(), "emulator-5554", "1.2")
			},
			command:          "font-size",
			value:            "1.2",
			expectedOutput:   []string{"Default font size: 1.0", "Current font size: 1.2"},
			expectedCommands: []string{"adb devices -l", "adb -s emulator-5554 shell settings put system font_scale 1.2", "adb -s emulator-5554 shell settings get system font_scale"},
		},
		{
			name: "get font size fails - adb command error",
			setupStubs: func(f *util.GenericExecFaker, cfg *config.Config) {
				f.StubSingleDevice(cfg.GetADBPath())
				f.StubADBShellCommand(cfg.GetADBPath(), "emulator-5554", []string{"settings", "get", "system", "font_scale"}, "", "settings: command not found", 1)
			},
			command:          "font-size",
			expectedError:    "failed to get current font size",
			expectedCommands: []string{"adb devices -l", "adb -s emulator-5554 shell settings get system font_scale"},
		},
		{
			name: "set font size fails - adb command error",
			setupStubs: func(f *util.GenericExecFaker, cfg *config.Config) {
				f.StubSingleDevice(cfg.GetADBPath())
				f.StubFontSizeSet(cfg.GetADBPath(), "emulator-5554", "2.0", 1)
			},
			command:          "font-size",
			value:            "2.0",
			expectedError:    "failed to set font size to 2.0",
			expectedCommands: []string{"adb devices -l", "adb -s emulator-5554 shell settings put system font_scale 2.0"},
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
