package wifi

import (
	"gadget/internal/config"
)

// WiFiFeature handles WiFi connection operations
type WiFiFeature struct {
	config            *config.Config
	connectingWiFi    bool
	disconnectingWiFi bool
	pairingWiFi       bool
	pairingAddress    string // Store pairing address between input steps
}

// NewWiFiFeature creates a new WiFi feature instance
func NewWiFiFeature(cfg *config.Config) *WiFiFeature {
	return &WiFiFeature{
		config: cfg,
	}
}

// IsConnecting returns true if WiFi connection is in progress
func (w *WiFiFeature) IsConnecting() bool {
	return w.connectingWiFi
}

// IsDisconnecting returns true if WiFi disconnection is in progress
func (w *WiFiFeature) IsDisconnecting() bool {
	return w.disconnectingWiFi
}

// IsPairing returns true if WiFi pairing is in progress
func (w *WiFiFeature) IsPairing() bool {
	return w.pairingWiFi
}

// GetPairingAddress returns the stored pairing address
func (w *WiFiFeature) GetPairingAddress() string {
	return w.pairingAddress
}

// SetConnecting sets the connecting state
func (w *WiFiFeature) SetConnecting(connecting bool) {
	w.connectingWiFi = connecting
}

// SetDisconnecting sets the disconnecting state
func (w *WiFiFeature) SetDisconnecting(disconnecting bool) {
	w.disconnectingWiFi = disconnecting
}

// SetPairing sets the pairing state
func (w *WiFiFeature) SetPairing(pairing bool) {
	w.pairingWiFi = pairing
}

// SetPairingAddress sets the pairing address
func (w *WiFiFeature) SetPairingAddress(address string) {
	w.pairingAddress = address
}

// ClearPairingAddress clears the stored pairing address
func (w *WiFiFeature) ClearPairingAddress() {
	w.pairingAddress = ""
}
