package commands

import (
	"fmt"
	"gadget/internal/adb"
	"gadget/internal/config"
	"gadget/internal/logger"
	"strconv"
	"strings"
	"time"
)

const DefaultWiFiPort = 4444

// ConnectWiFi attempts to connect to a device over WiFi
// For modern Android (11+), this requires pairing first
func ConnectWiFi(cfg *config.Config, ipAndPort string) error {
	adbPath := cfg.GetADBPath()
	ip, port, err := ParseIPAndPort(ipAndPort)
	if err != nil {
		return err
	}

	// If no port specified, default to our static port 4444
	if port == 0 {
		port = DefaultWiFiPort
		ipAndPort = fmt.Sprintf("%s:%d", ip, port)
	}

	// Try connecting to the specified address
	logger.Info("Attempting to connect to %s...", ipAndPort)
	output, err := adb.ExecuteGlobalCommandWithOutput(adbPath, "connect", ipAndPort)
	if err == nil && strings.Contains(output, "connected to") {
		logger.Success("Successfully connected to %s", ipAndPort)

		// If we connected to a non-standard port, try to switch to our standard port
		if port != DefaultWiFiPort {
			logger.Info("Switching device to standard port %d...", DefaultWiFiPort)
			switchErr := adb.ExecuteCommand(adbPath, ipAndPort, "tcpip", fmt.Sprintf("%d", DefaultWiFiPort))
			if switchErr != nil {
				logger.Error("Warning: failed to switch to standard port: %v", switchErr)
				logger.Info("Device will remain on port %d", port)
				return nil
			}

			time.Sleep(2 * time.Second)

			// Try connecting to the standard port
			standardAddress := fmt.Sprintf("%s:%d", ip, DefaultWiFiPort)
			logger.Info("Connecting to standard port %s...", standardAddress)

			standardOutput, standardErr := adb.ExecuteGlobalCommandWithOutput(adbPath, "connect", standardAddress)
			if standardErr == nil && strings.Contains(standardOutput, "connected to") {
				logger.Success("Successfully switched to standard port %s", standardAddress)

				// Disconnect from the original port
				logger.Info("Disconnecting from temporary port %s...", ipAndPort)
				adb.ExecuteGlobalCommand(adbPath, "disconnect", ipAndPort)

				return nil
			} else {
				logger.Error("Warning: failed to connect to standard port, keeping original connection")
			}
		}

		// Clean up any stale mDNS WiFi connections after successful connection
		CleanupStaleWiFiConnections(cfg)

		return nil
	}

	// Log the actual error for debugging
	if err != nil {
		logger.Error("Connection command failed: %v", err)
	} else {
		logger.Error("Connection rejected: %s", strings.TrimSpace(output))
	}

	return fmt.Errorf("failed to connect to %s. Device may need pairing first", ipAndPort)
}

// DisconnectWiFi disconnects from a WiFi device
func DisconnectWiFi(cfg *config.Config, ipAndPort string) error {
	adbPath := cfg.GetADBPath()
	ip, port, err := ParseIPAndPort(ipAndPort)
	if err != nil {
		return err
	}

	// If no port specified, default to our static port 4444
	if port == 0 {
		port = DefaultWiFiPort
		ipAndPort = fmt.Sprintf("%s:%d", ip, port)
	}

	logger.Info("Disconnecting from %s...", ipAndPort)

	// Check what devices are currently connected first
	output, err := adb.ExecuteGlobalCommandWithOutput(adbPath, "devices")
	if err == nil {
		logger.Info("Currently connected devices:\n%s", output)
	}

	err = adb.ExecuteGlobalCommand(adbPath, "disconnect", ipAndPort)
	if err != nil {
		// Check if the error is because the device wasn't connected
		if strings.Contains(err.Error(), "exit status 1") {
			return fmt.Errorf("device %s was not connected", ipAndPort)
		}
		return fmt.Errorf("failed to disconnect from %s: %w", ipAndPort, err)
	}

	logger.Success("Disconnected from %s", ipAndPort)

	// Clean up any stale mDNS WiFi connections
	CleanupStaleWiFiConnections(cfg)

	return nil
}

// CleanupStaleWiFiConnections removes stale mDNS WiFi connections
func CleanupStaleWiFiConnections(cfg *config.Config) {
	adbPath := cfg.GetADBPath()

	// Get current devices list
	output, err := adb.ExecuteGlobalCommandWithOutput(adbPath, "devices")
	if err != nil {
		return
	}

	// Find and disconnect from mDNS WiFi entries
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "_adb-tls-connect._tcp") && strings.Contains(line, "device") {
			// Extract the device identifier (everything before the tab/spaces)
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				deviceId := parts[0]
				logger.Info("Cleaning up stale WiFi connection: %s", deviceId)
				adb.ExecuteGlobalCommand(adbPath, "disconnect", deviceId)
			}
		}
	}
}

// ParseIPAndPort parses an input string that may contain IP:port or just IP
// Returns IP and port, with port defaulting to 0 if not provided
func ParseIPAndPort(input string) (string, int, error) {
	parts := strings.Split(input, ":")
	if len(parts) == 1 {
		// Just IP provided
		return parts[0], 0, nil
	} else if len(parts) == 2 {
		// IP:port provided
		port, err := strconv.Atoi(parts[1])
		if err != nil {
			return "", 0, fmt.Errorf("invalid port number: %s", parts[1])
		}
		if port < 1 || port > 65535 {
			return "", 0, fmt.Errorf("port number out of range: %d", port)
		}
		return parts[0], port, nil
	}

	return "", 0, fmt.Errorf("invalid IP address format: %s", input)
}
