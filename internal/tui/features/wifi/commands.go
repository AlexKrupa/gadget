package wifi

import (
	"adx/internal/commands"
	"adx/internal/config"
	"adx/internal/tui/messaging"
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// ConnectWiFiCmd returns a command to connect to a WiFi device
func ConnectWiFiCmd(cfg *config.Config, ipAndPort string) tea.Cmd {
	return executeWiFiOperation(cfg, WiFiConnect, ipAndPort, "")
}

// DisconnectWiFiCmd returns a command to disconnect from a WiFi device
func DisconnectWiFiCmd(cfg *config.Config, ipAndPort string) tea.Cmd {
	return executeWiFiOperation(cfg, WiFiDisconnect, ipAndPort, "")
}

// PairWiFiCmd returns a command to pair with a WiFi device
func PairWiFiCmd(cfg *config.Config, ipAndPort, pairingCode string) tea.Cmd {
	return executeWiFiOperation(cfg, WiFiPair, ipAndPort, pairingCode)
}

// executeWiFiOperation executes a WiFi operation asynchronously with generic handling
func executeWiFiOperation(cfg *config.Config, operation WiFiOperation, ipAndPort, pairingCode string) tea.Cmd {
	return func() tea.Msg {
		var err error
		var successMsg string

		switch operation {
		case WiFiConnect:
			err = commands.ConnectWiFi(cfg, ipAndPort)
			successMsg = fmt.Sprintf("WiFi device connected: %s", ipAndPort)
		case WiFiDisconnect:
			err = commands.DisconnectWiFi(cfg, ipAndPort)
			successMsg = fmt.Sprintf("WiFi device disconnected: %s", ipAndPort)
		case WiFiPair:
			err = commands.PairWiFiDevice(cfg, ipAndPort, pairingCode)
			successMsg = fmt.Sprintf("WiFi device paired and connected: %s", ipAndPort)
		}

		if err != nil {
			switch operation {
			case WiFiConnect:
				return messaging.WiFiConnectDoneMsg{Success: false, Message: err.Error()}
			case WiFiDisconnect:
				return messaging.WiFiDisconnectDoneMsg{Success: false, Message: err.Error()}
			case WiFiPair:
				return messaging.WiFiPairDoneMsg{Success: false, Message: err.Error()}
			}
		}

		// Small delay for connect/disconnect operations to ensure device list is updated
		if operation == WiFiConnect || operation == WiFiDisconnect {
			time.Sleep(500 * time.Millisecond)
		}

		switch operation {
		case WiFiConnect:
			return messaging.WiFiConnectDoneMsg{Success: true, Message: successMsg}
		case WiFiDisconnect:
			return messaging.WiFiDisconnectDoneMsg{Success: true, Message: successMsg}
		case WiFiPair:
			return messaging.WiFiPairDoneMsg{Success: true, Message: successMsg}
		}

		return nil // Should never reach here
	}
}
