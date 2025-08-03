package wifi

import (
	"fmt"
	"gadget/internal/tui/messaging"

	tea "github.com/charmbracelet/bubbletea"
)

// StartWiFiConnect starts a WiFi connection operation
func (w *WiFiFeature) StartWiFiConnect(input string) tea.Cmd {
	w.SetConnecting(true)
	return ConnectWiFiCmd(w.config, input)
}

// StartWiFiDisconnect starts a WiFi disconnection operation
func (w *WiFiFeature) StartWiFiDisconnect(input string) tea.Cmd {
	w.SetDisconnecting(true)
	return DisconnectWiFiCmd(w.config, input)
}

// StartWiFiPair starts a WiFi pairing operation
func (w *WiFiFeature) StartWiFiPair(address, code string) tea.Cmd {
	w.SetPairing(true)
	return PairWiFiCmd(w.config, address, code)
}

// HandleWiFiConnectDone handles the completion of a WiFi connect operation
func (w *WiFiFeature) HandleWiFiConnectDone(msg messaging.WiFiConnectDoneMsg) (tea.Model, tea.Cmd, string, string) {
	w.SetConnecting(false)
	if msg.Success {
		return nil, nil, msg.Message, ""
	}
	return nil, nil, "", fmt.Sprintf("WiFi connect failed: %s", msg.Message)
}

// HandleWiFiDisconnectDone handles the completion of a WiFi disconnect operation
func (w *WiFiFeature) HandleWiFiDisconnectDone(msg messaging.WiFiDisconnectDoneMsg) (tea.Model, tea.Cmd, string, string) {
	w.SetDisconnecting(false)
	if msg.Success {
		return nil, nil, msg.Message, ""
	}
	return nil, nil, "", fmt.Sprintf("WiFi disconnect failed: %s", msg.Message)
}

// HandleWiFiPairDone handles the completion of a WiFi pair operation
func (w *WiFiFeature) HandleWiFiPairDone(msg messaging.WiFiPairDoneMsg) (tea.Model, tea.Cmd, string, string) {
	w.SetPairing(false)
	if msg.Success {
		return nil, nil, msg.Message, ""
	}
	return nil, nil, "", fmt.Sprintf("WiFi pair failed: %s", msg.Message)
}
