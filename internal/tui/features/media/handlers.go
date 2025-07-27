package media

import (
	"adx/internal/tui/messaging"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

// HandleScreenshotDone handles the completion of a screenshot operation
func (m *MediaFeature) HandleScreenshotDone(msg messaging.ScreenshotDoneMsg) (tea.Model, tea.Cmd, string, string) {
	m.FinishScreenshot()

	if msg.Success {
		return nil, nil, msg.Message, ""
	}
	return nil, nil, "", fmt.Sprintf("Screenshot failed: %s", msg.Message)
}

// HandleDayNightScreenshotDone handles the completion of a day-night screenshot operation
func (m *MediaFeature) HandleDayNightScreenshotDone(msg messaging.DayNightScreenshotDoneMsg) (tea.Model, tea.Cmd, string, string) {
	m.FinishDayNightScreenshot()

	if msg.Success {
		return nil, nil, msg.Message, ""
	}
	return nil, nil, "", fmt.Sprintf("Day-night screenshots failed: %s", msg.Message)
}

// HandleRecordingStarted handles the start of a screen recording
func (m *MediaFeature) HandleRecordingStarted(msg messaging.RecordingStartedMsg) (tea.Model, tea.Cmd, string, string) {
	if msg.Err != nil {
		m.recordingScreen = false
		return nil, nil, "", fmt.Sprintf("Failed to start recording: %s", msg.Err.Error())
	}

	m.activeRecording = msg.Recording
	return nil, nil, "", ""
}

// HandleScreenRecordDone handles the completion of a screen recording
func (m *MediaFeature) HandleScreenRecordDone(msg messaging.ScreenRecordDoneMsg) (tea.Model, tea.Cmd, string, string) {
	m.FinishRecording()

	if msg.Success {
		return nil, nil, msg.Message, ""
	}
	return nil, nil, "", fmt.Sprintf("Screen recording failed: %s", msg.Message)
}

// GetStatusText returns status text for active media operations
func (m *MediaFeature) GetStatusText() string {
	if m.takingScreenshot {
		return "Taking screenshot..."
	}
	if m.takingDayNight {
		return "Taking day-night screenshots..."
	}
	if m.recordingScreen {
		return "Recording screen... (Press 'r' to stop)"
	}
	return ""
}
