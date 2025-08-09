package wifi

import (
	"gadget/internal/commands"
	"gadget/internal/config"
	"gadget/internal/tui/features/media"

	tea "github.com/charmbracelet/bubbletea"
)

// ConnectWiFiCmd returns a command to connect to a WiFi device
func ConnectWiFiCmd(cfg *config.Config, ipAndPort string) tea.Cmd {
	return media.StreamCommand(func() error {
		return commands.ConnectWiFi(cfg, ipAndPort)
	})
}

// DisconnectWiFiCmd returns a command to disconnect from a WiFi device
func DisconnectWiFiCmd(cfg *config.Config, ipAndPort string) tea.Cmd {
	return media.StreamCommand(func() error {
		return commands.DisconnectWiFi(cfg, ipAndPort)
	})
}

// PairWiFiCmd returns a command to pair with a WiFi device
func PairWiFiCmd(cfg *config.Config, ipAndPort, pairingCode string) tea.Cmd {
	return media.StreamCommand(func() error {
		return commands.PairWiFiDevice(cfg, ipAndPort, pairingCode)
	})
}
