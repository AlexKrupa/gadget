package media

import (
	"adx/internal/commands"
	"adx/internal/config"
)

// ScreenshotOperation defines the type of screenshot operation
type ScreenshotOperation int

const (
	ScreenshotSingle ScreenshotOperation = iota
	ScreenshotDayNight
)

// MediaFeature handles screenshot and screen recording operations
type MediaFeature struct {
	config           *config.Config
	takingScreenshot bool
	takingDayNight   bool
	recordingScreen  bool
	activeRecording  *commands.ScreenRecording
}

// NewMediaFeature creates a new media feature instance
func NewMediaFeature(cfg *config.Config) *MediaFeature {
	return &MediaFeature{
		config: cfg,
	}
}

// IsActive returns true if any media operation is in progress
func (m *MediaFeature) IsActive() bool {
	return m.takingScreenshot || m.takingDayNight || m.recordingScreen
}

// IsTakingScreenshot returns true if a screenshot operation is in progress
func (m *MediaFeature) IsTakingScreenshot() bool {
	return m.takingScreenshot
}

// IsTakingDayNight returns true if a day-night screenshot operation is in progress
func (m *MediaFeature) IsTakingDayNight() bool {
	return m.takingDayNight
}

// IsRecording returns true if screen recording is in progress
func (m *MediaFeature) IsRecording() bool {
	return m.recordingScreen
}

// GetActiveRecording returns the current recording session if any
func (m *MediaFeature) GetActiveRecording() *commands.ScreenRecording {
	return m.activeRecording
}

// StartScreenshot marks screenshot operation as started
func (m *MediaFeature) StartScreenshot() {
	m.takingScreenshot = true
}

// StartDayNightScreenshot marks day-night screenshot operation as started
func (m *MediaFeature) StartDayNightScreenshot() {
	m.takingDayNight = true
}

// StartRecording marks screen recording as started
func (m *MediaFeature) StartRecording() {
	m.recordingScreen = true
}

// FinishScreenshot marks screenshot operation as completed
func (m *MediaFeature) FinishScreenshot() {
	m.takingScreenshot = false
}

// FinishDayNightScreenshot marks day-night screenshot operation as completed
func (m *MediaFeature) FinishDayNightScreenshot() {
	m.takingDayNight = false
}

// FinishRecording marks screen recording as completed and clears active recording
func (m *MediaFeature) FinishRecording() {
	m.recordingScreen = false
	m.activeRecording = nil
}

// SetActiveRecording sets the current recording session
func (m *MediaFeature) SetActiveRecording(recording *commands.ScreenRecording) {
	m.activeRecording = recording
}
