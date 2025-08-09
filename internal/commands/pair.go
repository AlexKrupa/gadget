package commands

import (
	"fmt"
	"gadget/internal/adb"
	"gadget/internal/config"
	"gadget/internal/logger"
	"strings"
)

// PairWiFiDevice pairs with a WiFi device using a pairing code
func PairWiFiDevice(cfg *config.Config, ipAndPort, pairingCode string) error {
	adbPath := cfg.GetADBPath()

	logger.Info("Pairing with %s using code %s...", ipAndPort, pairingCode)

	output, err := adb.ExecuteGlobalCommandWithOutput(adbPath, "pair", ipAndPort, pairingCode)
	if err != nil {
		return fmt.Errorf("pairing command failed: %w", err)
	}

	if strings.Contains(output, "Successfully paired") {
		logger.Success("Successfully paired with %s", ipAndPort)

		// After pairing, we need to check what port the device is actually listening on
		logger.Info("")
		logger.Info("Pairing successful! Now you need to:")
		logger.Info("1. Check the main 'IP address & Port' on your phone (not the pairing section)")
		logger.Info("2. Use the Connect WiFi command (menu 8) with that address")
		logger.Info("3. The tool will then set it to use port %d permanently", DefaultWiFiPort)
		logger.Info("")

		CleanupStaleWiFiConnections(cfg)

		// For now, return success since pairing worked
		return nil
	}

	return fmt.Errorf("pairing failed: %s", strings.TrimSpace(output))
}

// PairWiFiDeviceForTUI pairs with a WiFi device using progress callback
func PairWiFiDeviceForTUI(cfg *config.Config, ipAndPort, pairingCode string, progress func(string)) error {
	adbPath := cfg.GetADBPath()

	progress(fmt.Sprintf("Pairing with %s using code %s...", ipAndPort, pairingCode))

	output, err := adb.ExecuteGlobalCommandWithOutput(adbPath, "pair", ipAndPort, pairingCode)
	if err != nil {
		progress(fmt.Sprintf("Pairing command failed: %v", err))
		return fmt.Errorf("pairing command failed: %w", err)
	}

	if strings.Contains(output, "Successfully paired") {
		progress(fmt.Sprintf("Successfully paired with %s", ipAndPort))
		progress("Pairing successful! Now you need to:")
		progress("1. Check the main 'IP address & Port' on your phone (not the pairing section)")
		progress("2. Use the Connect WiFi command with that address")
		progress(fmt.Sprintf("3. The tool will then set it to use port %d permanently", DefaultWiFiPort))

		CleanupStaleWiFiConnections(cfg)
		return nil
	}

	errorMsg := fmt.Sprintf("Pairing failed: %s", strings.TrimSpace(output))
	progress(errorMsg)
	return fmt.Errorf("%s", errorMsg)
}
