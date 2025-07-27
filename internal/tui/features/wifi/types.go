package wifi

import (
	"adx/internal/config"
)

// WiFiOperation defines the type of WiFi operation
type WiFiOperation int

const (
	WiFiConnect WiFiOperation = iota
	WiFiDisconnect
	WiFiPair
)

// WiFiFeature handles WiFi connection operations
type WiFiFeature struct {
	config *config.Config
}

// NewWiFiFeature creates a new WiFi feature instance
func NewWiFiFeature(cfg *config.Config) *WiFiFeature {
	return &WiFiFeature{
		config: cfg,
	}
}
