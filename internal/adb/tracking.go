package adb

import (
	"bufio"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// DeviceChangeEvent represents a device connection/disconnection event
type DeviceChangeEvent struct {
	Serial string
	Status string // "device", "offline", "disconnected"
}

// StartDeviceTracking starts monitoring device changes using adb track-devices
// Returns a channel that emits DeviceChangeEvent when devices connect/disconnect
func StartDeviceTracking(adbPath string) (<-chan DeviceChangeEvent, error) {
	eventChan := make(chan DeviceChangeEvent, 10)

	cmd := exec.Command(adbPath, "track-devices")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		close(eventChan)
		return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		close(eventChan)
		return nil, fmt.Errorf("failed to start adb track-devices: %w", err)
	}

	go func() {
		defer close(eventChan)
		defer cmd.Process.Kill() // Cleanup when goroutine exits

		scanner := bufio.NewScanner(stdout)

		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" {
				continue
			}

			// Parse device status line: "SERIAL\tSTATUS"
			parts := strings.Split(line, "\t")
			if len(parts) >= 2 {
				serial := strings.TrimSpace(parts[0])
				status := strings.TrimSpace(parts[1])

				// Send event if device status is meaningful
				if serial != "" && (status == "device" || status == "offline" || status == "disconnected") {
					select {
					case eventChan <- DeviceChangeEvent{Serial: serial, Status: status}:
					case <-time.After(100 * time.Millisecond):
						// Skip if channel is full to prevent blocking
					}
				}
			}
		}

		// Handle scanner errors
		if err := scanner.Err(); err != nil {
			// Don't send error events to avoid noise, just log internally
			// The tracking will be restarted by the caller if needed
		}
	}()

	return eventChan, nil
}
