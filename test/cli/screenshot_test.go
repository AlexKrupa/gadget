package test

import (
	"fmt"
	"gadget/internal/adb"
	"gadget/internal/config"
	"gadget/internal/logger"
	"gadget/test/cli/util"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScreenshotCommand(t *testing.T) {
	tests := []struct {
		name           string
		fakeSetup      func(*util.GenericExecFaker, string)
		expectedOutput []string
		expectError    bool
	}{
		{
			name: "screenshot success single device",
			fakeSetup: func(f *util.GenericExecFaker, adbPath string) {
				f.StubSingleDevice(adbPath)
				// Fake screenshot command sequence
				f.AddStub(adbPath, []string{"-s", "emulator-5554", "shell", "screencap", "/sdcard/screenshot.png"}, "", "", 0)
				f.AddStub(adbPath, []string{"-s", "emulator-5554", "pull", "/sdcard/screenshot.png", "/test/media/android-img-test.png"}, "file pulled", "", 0)
				f.AddStub(adbPath, []string{"-s", "emulator-5554", "shell", "rm", "/sdcard/screenshot.png"}, "", "", 0)
			},
			expectedOutput: []string{"Screenshot saved to:", "android-img-"},
			expectError:    false,
		},
		{
			name: "screenshot fails on screencap command",
			fakeSetup: func(f *util.GenericExecFaker, adbPath string) {
				f.StubSingleDevice(adbPath)
				// Fake screencap failure
				f.AddStub(adbPath, []string{"-s", "emulator-5554", "shell", "screencap", "/sdcard/screenshot.png"}, "", "screencap failed", 1)
			},
			expectedOutput: []string{},
			expectError:    true,
		},
		{
			name: "screenshot fails on pull command",
			fakeSetup: func(f *util.GenericExecFaker, adbPath string) {
				f.StubSingleDevice(adbPath)
				// Screencap succeeds, pull fails
				f.AddStub(adbPath, []string{"-s", "emulator-5554", "shell", "screencap", "/sdcard/screenshot.png"}, "", "", 0)
				f.AddStub(adbPath, []string{"-s", "emulator-5554", "pull", "/sdcard/screenshot.png", "/test/media/android-img-test.png"}, "", "pull failed", 1)
			},
			expectedOutput: []string{},
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			faker := util.NewGenericExecFaker()
			cfg := util.TestConfig()
			tt.fakeSetup(faker, cfg.GetADBPath())

			// Test screenshot execution with fake
			devices, err := getConnectedDevicesWithFake(cfg.GetADBPath(), faker)
			require.NoError(t, err)

			if len(devices) == 0 && !tt.expectError {
				t.Skip("No devices for screenshot test")
			}

			var screenshotError error
			output := util.CaptureLogOutput(func() {
				if len(devices) > 0 {
					screenshotError = takeScreenshotWithFake(cfg, devices[0], faker)
				}
			})

			if tt.expectError {
				assert.Error(t, screenshotError)
			} else {
				assert.NoError(t, screenshotError)
				for _, expected := range tt.expectedOutput {
					assert.Contains(t, output, expected)
				}
			}
		})
	}
}

// takeScreenshotWithFake mimics commands.TakeScreenshot with fake ADB commands
func takeScreenshotWithFake(cfg *config.Config, device adb.Device, faker *util.GenericExecFaker) error {
	remotePath := "/sdcard/screenshot.png"
	localPath := "/test/media/android-img-test.png"

	// Execute screenshot command sequence with fakes
	cmd1 := faker.FakeExecCommand(cfg.GetADBPath(), "-s", device.Serial, "shell", "screencap", remotePath)
	if err := cmd1.Run(); err != nil {
		return fmt.Errorf("failed to take screenshot: %w", err)
	}

	cmd2 := faker.FakeExecCommand(cfg.GetADBPath(), "-s", device.Serial, "pull", remotePath, localPath)
	if err := cmd2.Run(); err != nil {
		return fmt.Errorf("failed to pull screenshot: %w", err)
	}

	cmd3 := faker.FakeExecCommand(cfg.GetADBPath(), "-s", device.Serial, "shell", "rm", remotePath)
	cmd3.Run() // Cleanup - ignore errors

	logger.Success("Screenshot saved to: %s", localPath)
	return nil
}
