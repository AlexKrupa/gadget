package commands

import (
	"adx/internal/adb"
	"adx/internal/config"
	"fmt"
	"strings"
)

// PairWiFiDevice pairs with a WiFi device using a pairing code
func PairWiFiDevice(cfg *config.Config, ipAndPort, pairingCode string) error {
	adbPath := cfg.GetADBPath()
	
	fmt.Printf("Pairing with %s using code %s...\n", ipAndPort, pairingCode)
	
	// Use adb pair command for newer Android versions
	output, err := adb.ExecuteGlobalCommandWithOutput(adbPath, "pair", ipAndPort, pairingCode)
	if err != nil {
		return fmt.Errorf("pairing command failed: %w", err)
	}
	
	// Check if pairing was successful
	if strings.Contains(output, "Successfully paired") {
		fmt.Printf("Successfully paired with %s\n", ipAndPort)
		
		// After pairing, we need to check what port the device is actually listening on
		// On modern Android, this is shown in the main "IP address & Port" section
		fmt.Printf("\nPairing successful! Now you need to:\n")
		fmt.Printf("1. Check the main 'IP address & Port' on your phone (not the pairing section)\n")
		fmt.Printf("2. Use the Connect WiFi command (menu 8) with that address\n")
		fmt.Printf("3. The tool will then set it to use port %d permanently\n\n", DefaultWiFiPort)
		
		// Clean up any stale mDNS WiFi connections after pairing
		CleanupStaleWiFiConnections(cfg)
		
		// For now, return success since pairing worked
		// The user will need to use the connect command with the service port
		return nil
	}
	
	return fmt.Errorf("pairing failed: %s", strings.TrimSpace(output))
}