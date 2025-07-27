package wifi

import (
	"adx/internal/tui/messaging"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

// HandleWiFiConnectDone handles the completion of a WiFi connect operation
func (w *WiFiFeature) HandleWiFiConnectDone(msg messaging.WiFiConnectDoneMsg) (tea.Model, tea.Cmd, string, string) {
	if msg.Success {
		return nil, nil, msg.Message, ""
	}
	return nil, nil, "", fmt.Sprintf("WiFi connect failed: %s", msg.Message)
}

// HandleWiFiDisconnectDone handles the completion of a WiFi disconnect operation
func (w *WiFiFeature) HandleWiFiDisconnectDone(msg messaging.WiFiDisconnectDoneMsg) (tea.Model, tea.Cmd, string, string) {
	if msg.Success {
		return nil, nil, msg.Message, ""
	}
	return nil, nil, "", fmt.Sprintf("WiFi disconnect failed: %s", msg.Message)
}

// HandleWiFiPairDone handles the completion of a WiFi pair operation
func (w *WiFiFeature) HandleWiFiPairDone(msg messaging.WiFiPairDoneMsg) (tea.Model, tea.Cmd, string, string) {
	if msg.Success {
		return nil, nil, msg.Message, ""
	}
	return nil, nil, "", fmt.Sprintf("WiFi pair failed: %s", msg.Message)
}
