package tui

import (
	"adx/internal/adb"
	"adx/internal/commands"
	"adx/internal/config"
	"adx/internal/emulator"
	"adx/internal/tui/core"
	"adx/internal/tui/features/media"
	"adx/internal/tui/features/wifi"
	"fmt"
	"sort"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Aliases for core types for backward compatibility
type Mode = core.Mode
type LogEntry = core.LogEntry
type LogType = core.LogType
type Command = core.Command
type CommandCategory = core.CommandCategory

// Constants from core
const (
	ModeMenu           = core.ModeMenu
	ModeDeviceSelect   = core.ModeDeviceSelect
	ModeEmulatorSelect = core.ModeEmulatorSelect
	ModeCommand        = core.ModeCommand
	ModeTextInput      = core.ModeTextInput
)

const (
	LogTypeSuccess = core.LogTypeSuccess
	LogTypeError   = core.LogTypeError
	LogTypeInfo    = core.LogTypeInfo
)

// tickMsg is sent for progress animation - keep local for now
type tickMsg time.Time

// Delegate to core functions
func getAvailableCommands() []Command {
	return core.GetAvailableCommands()
}

func getCommandCategories() []CommandCategory {
	return core.GetCommandCategories()
}

// GetAvailableCommandNames returns a list of all available command names for CLI help
func GetAvailableCommandNames() []string {
	return core.GetAvailableCommandNames()
}

// Model represents the TUI state
type Model struct {
	config           *config.Config
	devices          []adb.Device
	avds             []emulator.AVD
	selectedDevice   int
	selectedEmulator int
	selectedCommand  int
	mode             Mode
	err              error
	successMsg       string
	quitting         bool

	// Log system
	logHistory    []LogEntry
	maxLogEntries int
	loading       bool

	// Features
	mediaFeature            *media.MediaFeature
	wifiFeature             *wifi.WiFiFeature
	textInput               string
	textInputPrompt         string
	textInputAction         string
	selectedDeviceForAction adb.Device
	currentSettingInfo      *commands.SettingInfo
	currentSettingType      commands.SettingType
	connectingWiFi          bool
	disconnectingWiFi       bool
	pairingWiFi             bool
	pairingAddress          string // Store pairing address between input steps

	// Command search fields
	searchFilter         string
	filteredCommands     []Command
	selectedCommandIndex int

	// Progress tracking
	operationStartTime time.Time
	progressTicker     int // For animated progress indicators
}

// NewModel creates a new TUI model
func NewModel(cfg *config.Config) Model {
	m := Model{
		config:               cfg,
		mode:                 ModeMenu,
		selectedDevice:       0,
		selectedEmulator:     0,
		selectedCommand:      0,
		loading:              true,
		searchFilter:         "",
		selectedCommandIndex: 0,
		logHistory:           make([]LogEntry, 0),
		maxLogEntries:        5, // Keep last 5 log entries
		mediaFeature:         media.NewMediaFeature(cfg),
		wifiFeature:          wifi.NewWiFiFeature(cfg),
	}
	m.filteredCommands = m.filterCommands()
	return m
}

// addLogEntry adds a new log entry and maintains the history limit
func (m *Model) addLogEntry(message string, logType LogType) {
	// Normalize indentation and whitespace
	normalizedMessage := strings.TrimSpace(strings.ReplaceAll(message, "\t", "  "))

	entry := LogEntry{
		Message:   normalizedMessage,
		Type:      logType,
		Timestamp: time.Now(),
	}

	m.logHistory = append(m.logHistory, entry)

	// Keep only the last maxLogEntries
	if len(m.logHistory) > m.maxLogEntries {
		m.logHistory = m.logHistory[len(m.logHistory)-m.maxLogEntries:]
	}

	// Clear old success message system for backward compatibility
	m.successMsg = ""
	m.err = nil
}

// addSuccess adds a success log entry
func (m *Model) addSuccess(message string) {
	m.addLogEntry(message, LogTypeSuccess)
}

// addError adds an error log entry
func (m *Model) addError(message string) {
	m.addLogEntry(message, LogTypeError)
}

// addInfo adds an info log entry
func (m *Model) addInfo(message string) {
	m.addLogEntry(message, LogTypeInfo)
}

// clearLogs clears all log entries
func (m *Model) clearLogs() {
	m.logHistory = make([]LogEntry, 0)
	m.successMsg = ""
	m.err = nil
}

// CommandMatch holds a command and its match score
type CommandMatch struct {
	Command Command
	Score   int
}

// filterCommands applies fuzzy search to the command list and sorts by score
func (m Model) filterCommands() []Command {
	if m.searchFilter == "" {
		return getAvailableCommands()
	}

	var matches []CommandMatch
	filter := strings.ToLower(m.searchFilter)

	for _, cmd := range getAvailableCommands() {
		if score := m.fuzzyMatchScore(cmd, filter); score > 0 {
			matches = append(matches, CommandMatch{Command: cmd, Score: score})
		}
	}

	// Sort by score (higher is better)
	sort.Slice(matches, func(i, j int) bool {
		return matches[i].Score > matches[j].Score
	})

	// Extract commands from matches
	var filtered []Command
	for _, match := range matches {
		filtered = append(filtered, match.Command)
	}

	return filtered
}

// fuzzyMatchScore calculates a score for how well a command matches the filter
// Returns 0 if no match, higher scores for better matches
func (m Model) fuzzyMatchScore(cmd Command, filter string) int {

	name := strings.ToLower(cmd.Name)
	description := strings.ToLower(cmd.Description)

	// Check both name and description, take the higher score
	nameScore := m.fuzzyMatchStringScore(name, filter)
	descScore := m.fuzzyMatchStringScore(description, filter)

	maxScore := nameScore
	if descScore > maxScore {
		maxScore = descScore
	}

	// Boost score if name matches (prefer name matches over description)
	if nameScore > 0 {
		maxScore += 50
	}

	return maxScore
}

// fuzzyMatchStringScore calculates fuzzy match score for a string
func (m Model) fuzzyMatchStringScore(str, filter string) int {
	if filter == "" {
		return 0
	}

	strRunes := []rune(str)
	filterRunes := []rune(filter)

	filterIndex := 0
	score := 0
	consecutiveBonus := 0
	matchPositions := []int{}

	for i, strChar := range strRunes {
		if filterIndex < len(filterRunes) && strChar == filterRunes[filterIndex] {
			matchPositions = append(matchPositions, i)

			// Base score for character match
			score += 10

			// Bonus for consecutive characters
			if filterIndex > 0 && i > 0 && strRunes[i-1] == filterRunes[filterIndex-1] {
				consecutiveBonus += 5
				score += consecutiveBonus
			} else {
				consecutiveBonus = 0
			}

			// Smaller bonus for matching at word start (reduced from 15 to 8)
			if i == 0 || strRunes[i-1] == ' ' || strRunes[i-1] == '-' {
				score += 8
			}

			filterIndex++
		}
	}

	// Only return score if all filter characters were matched
	if filterIndex == len(filterRunes) {
		// Early position bonus: based on average position of all matches
		// Scale from 0-50 points based on how early matches occur on average
		positionBonus := 0
		if len(matchPositions) > 0 && len(strRunes) > 0 {
			// Calculate average position of all matches
			totalPos := 0
			for _, pos := range matchPositions {
				totalPos += pos
			}
			avgPos := totalPos / len(matchPositions)
			positionBonus = 50 - (avgPos * 50 / len(strRunes))
			if positionBonus < 0 {
				positionBonus = 0
			}
		}

		// Compactness bonus: reward matches that are close together
		compactnessBonus := 0
		if len(matchPositions) > 1 {
			span := matchPositions[len(matchPositions)-1] - matchPositions[0] + 1
			// Give bonus inversely proportional to span
			// Compact matches (small span) get up to 25 points
			maxSpan := len(strRunes)
			compactnessBonus = 25 - (span * 25 / maxSpan)
			if compactnessBonus < 0 {
				compactnessBonus = 0
			}
		}

		return score + positionBonus + compactnessBonus
	}

	return 0
}

// doTick returns a command that sends a tickMsg after a short delay
func doTick() tea.Cmd {
	return tea.Tick(time.Millisecond*100, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// Init initializes the model (required by Bubble Tea)
func (m Model) Init() tea.Cmd {
	m.operationStartTime = time.Now()
	return tea.Batch(loadDevices(m.config.GetADBPath()), doTick())
}

// Update handles messages and updates the model state
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyPress(msg)
	case devicesLoadedMsg:
		m.devices = msg.Devices
		m.err = msg.Err
		m.loading = false
		if len(m.devices) == 0 && m.err == nil {
			m.err = fmt.Errorf("no devices connected")
		}
		// Don't clear success messages during auto-refresh
		return m, nil
	case avdsLoadedMsg:
		m.avds = msg.Avds
		if msg.Err != nil {
			m.err = msg.Err
			m.mode = ModeMenu
		}
		return m, nil
	case screenshotDoneMsg:
		_, _, successMsg, errorMsg := m.mediaFeature.HandleScreenshotDone(msg)
		if successMsg != "" {
			m.addSuccess(successMsg)
		}
		if errorMsg != "" {
			m.addError(errorMsg)
		}
		return m, nil
	case dayNightScreenshotDoneMsg:
		_, _, successMsg, errorMsg := m.mediaFeature.HandleDayNightScreenshotDone(msg)
		if successMsg != "" {
			m.addSuccess(successMsg)
		}
		if errorMsg != "" {
			m.addError(errorMsg)
		}
		return m, nil
	case recordingStartedMsg:
		_, _, _, errorMsg := m.mediaFeature.HandleRecordingStarted(msg)
		if errorMsg != "" {
			m.err = fmt.Errorf(errorMsg)
			m.successMsg = ""
		}
		return m, nil
	case screenRecordDoneMsg:
		_, _, successMsg, errorMsg := m.mediaFeature.HandleScreenRecordDone(msg)
		if successMsg != "" {
			m.addSuccess(successMsg)
		}
		if errorMsg != "" {
			m.addError(errorMsg)
		}
		return m, nil
	case settingLoadedMsg:
		if msg.Err != nil {
			m.err = fmt.Errorf("failed to get current setting: %s", msg.Err.Error())
			m.mode = ModeMenu
		} else {
			m.currentSettingInfo = msg.SettingInfo
			m.currentSettingType = msg.SettingInfo.Type
			m.mode = ModeTextInput

			// Show both Default/Physical and Current values
			displayInfo := fmt.Sprintf("Physical %s: %s\nCurrent %s: %s",
				msg.SettingInfo.DisplayName, msg.SettingInfo.Default,
				msg.SettingInfo.DisplayName, msg.SettingInfo.Current)

			m.textInputPrompt = fmt.Sprintf("Device: %s\n%s\n\n%s",
				m.selectedDeviceForAction.Serial, displayInfo, msg.SettingInfo.InputPrompt)
		}
		return m, nil
	case settingChangedMsg:
		if msg.Success {
			m.addSuccess(msg.Message)
			// Refresh setting info to show updated values
			return m, getCurrentSetting(m.config, m.selectedDeviceForAction, msg.SettingType)
		} else {
			m.addError(fmt.Sprintf("Setting change failed: %s", msg.Message))
		}
		return m, nil
	case wifiConnectDoneMsg:
		m.connectingWiFi = false
		_, _, successMsg, errorMsg := m.wifiFeature.HandleWiFiConnectDone(msg)
		if successMsg != "" {
			m.addSuccess(successMsg)
			m.mode = ModeMenu
			// Refresh device list after successful WiFi connection
			return m, loadDevices(m.config.GetADBPath())
		} else if errorMsg != "" {
			m.addError(errorMsg)
		}
		return m, nil
	case wifiDisconnectDoneMsg:
		m.disconnectingWiFi = false
		_, _, successMsg, errorMsg := m.wifiFeature.HandleWiFiDisconnectDone(msg)
		if successMsg != "" {
			m.addSuccess(successMsg)
			m.mode = ModeMenu
			// Refresh device list after successful WiFi disconnection
			return m, loadDevices(m.config.GetADBPath())
		} else if errorMsg != "" {
			m.addError(errorMsg)
		}
		return m, nil
	case wifiPairDoneMsg:
		m.pairingWiFi = false
		_, _, successMsg, errorMsg := m.wifiFeature.HandleWiFiPairDone(msg)
		if successMsg != "" {
			m.addSuccess(successMsg)
			m.mode = ModeMenu
			// Refresh device list after successful WiFi pairing
			return m, loadDevices(m.config.GetADBPath())
		} else if errorMsg != "" {
			m.addError(errorMsg)
		}
		return m, nil
	case tickMsg:
		m.progressTicker++
		// Continue ticking if any operation is active
		if m.loading || m.mediaFeature.IsActive() ||
			m.connectingWiFi || m.disconnectingWiFi || m.pairingWiFi {
			return m, doTick()
		}
		return m, nil
	case tea.QuitMsg:
		m.quitting = true
		return m, tea.Quit
	}

	return m, nil
}

// handleKeyPress processes keyboard input
func (m Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Global key handling for active recording
	if m.mediaFeature.IsRecording() && msg.String() == "esc" {
		return m.stopRecording()
	}

	switch m.mode {
	case ModeMenu:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "esc":
			// Clear search filter if one exists
			if m.searchFilter != "" {
				m.searchFilter = ""
				m.filteredCommands = m.filterCommands()
				m.selectedCommandIndex = 0
			}
			return m, nil
		case "up":
			if m.selectedCommandIndex > 0 {
				m.selectedCommandIndex--
			}
			return m, nil
		case "down":
			if m.selectedCommandIndex < len(m.filteredCommands)-1 {
				m.selectedCommandIndex++
			}
			return m, nil
		case "enter":
			if len(m.filteredCommands) > 0 && m.selectedCommandIndex < len(m.filteredCommands) {
				return m.executeSelectedCommand()
			}
			return m, nil
		case "backspace":
			if len(m.searchFilter) > 0 {
				m.searchFilter = m.searchFilter[:len(m.searchFilter)-1]
				m.filteredCommands = m.filterCommands()
				// Reset selection if it's out of bounds
				if m.selectedCommandIndex >= len(m.filteredCommands) {
					m.selectedCommandIndex = 0
				}
			}
			return m, nil
		default:
			// Handle typing for search
			if len(msg.String()) == 1 {
				m.searchFilter += msg.String()
				m.filteredCommands = m.filterCommands()
				m.selectedCommandIndex = 0 // Reset to first item
			}
			return m, nil
		}
	case ModeDeviceSelect:
		switch msg.String() {
		case "esc":
			m.mode = ModeMenu
			return m, nil
		case "up", "k", "h":
			if m.selectedDevice > 0 {
				m.selectedDevice--
			}
			return m, nil
		case "down", "j", "l":
			if m.selectedDevice < len(m.devices)-1 {
				m.selectedDevice++
			}
			return m, nil
		case "enter":
			if m.selectedDevice < len(m.devices) {
				return m.executeCommandForDevice(m.devices[m.selectedDevice])
			}
		}
	case ModeEmulatorSelect:
		switch msg.String() {
		case "esc":
			m.mode = ModeMenu
			return m, nil
		case "up", "k", "h":
			if m.selectedEmulator > 0 {
				m.selectedEmulator--
			}
			return m, nil
		case "down", "j", "l":
			if m.selectedEmulator < len(m.avds)-1 {
				m.selectedEmulator++
			}
			return m, nil
		case "enter":
			if m.selectedEmulator < len(m.avds) {
				return m.launchEmulator(m.avds[m.selectedEmulator])
			}
		}
	case ModeTextInput:
		switch msg.String() {
		case "enter":
			return m.handleTextInputSubmit()
		case "esc":
			m.mode = ModeMenu
			m.textInput = ""
			m.textInputPrompt = ""
			m.textInputAction = ""
			return m, nil
		case "backspace":
			if len(m.textInput) > 0 {
				m.textInput = m.textInput[:len(m.textInput)-1]
			}
			return m, nil
		default:
			// Add character to input if it's a regular character
			if len(msg.String()) == 1 {
				m.textInput += msg.String()
			}
			return m, nil
		}
	}

	return m, nil
}

// executeScreenshot runs the screenshot command
func (m Model) executeScreenshot(device adb.Device) (tea.Model, tea.Cmd) {
	m.mode = ModeMenu
	m.clearLogs()
	m.mediaFeature.StartScreenshot()
	m.operationStartTime = time.Now()

	return m, tea.Batch(takeScreenshot(m.config, device), doTick())
}

// executeDayNightScreenshots runs the day-night screenshot command
func (m Model) executeDayNightScreenshots(device adb.Device) (tea.Model, tea.Cmd) {
	m.mode = ModeMenu
	m.clearLogs()
	m.mediaFeature.StartDayNightScreenshot()
	m.operationStartTime = time.Now()

	return m, tea.Batch(takeDayNightScreenshots(m.config, device), doTick())
}

// executeSelectedCommand executes the currently selected command from the filtered list
func (m Model) executeSelectedCommand() (tea.Model, tea.Cmd) {
	if len(m.filteredCommands) == 0 || m.selectedCommandIndex >= len(m.filteredCommands) {
		return m, nil
	}

	selectedCmd := m.filteredCommands[m.selectedCommandIndex]

	// Store the selected command for device selection
	m.selectedCommand = m.selectedCommandIndex

	switch selectedCmd.Command {
	case "launch-emulator":
		m.mode = ModeEmulatorSelect
		return m, loadAVDs(m.config)
	case "connect-wifi":
		m.mode = ModeTextInput
		m.textInputPrompt = "Enter IP address or IP:port (e.g., 192.168.1.100 or 192.168.1.100:5555)\nDefaults to port 4444 if not specified"
		m.textInputAction = "wifi_connect"
		m.textInput = ""
		return m, nil
	case "pair-wifi":
		m.mode = ModeTextInput
		m.textInputPrompt = "Enter pairing address from phone (e.g., 192.168.3.30:43719)"
		m.textInputAction = "wifi_pair_address"
		m.textInput = ""
		return m, nil
	case "disconnect-wifi":
		m.mode = ModeTextInput
		m.textInputPrompt = "Enter IP address or IP:port to disconnect (e.g., 192.168.1.100 or 192.168.1.100:5555)\nDefaults to port 4444 if not specified"
		m.textInputAction = "wifi_disconnect"
		m.textInput = ""
		return m, nil
	case "refresh-devices":
		m.clearLogs()
		return m, loadDevices(m.config.GetADBPath())
	default:
		// Commands that require device selection
		if len(m.devices) == 1 {
			return m.executeCommandForDevice(m.devices[0])
		}
		m.mode = ModeDeviceSelect
		return m, nil
	}
}

// executeCommandForDevice executes the selected command for a specific device
func (m Model) executeCommandForDevice(device adb.Device) (tea.Model, tea.Cmd) {
	if len(m.filteredCommands) == 0 || m.selectedCommandIndex >= len(m.filteredCommands) {
		return m, nil
	}

	selectedCmd := m.filteredCommands[m.selectedCommandIndex]

	switch selectedCmd.Command {
	case "screenshot":
		return m.executeScreenshot(device)
	case "screenshot-day-night":
		return m.executeDayNightScreenshots(device)
	case "screen-record":
		return m.executeScreenRecord(device)
	case "change-dpi":
		return m.startSettingChange(device, commands.SettingTypeDPI)
	case "change-font-size":
		return m.startSettingChange(device, commands.SettingTypeFontSize)
	case "change-screen-size":
		return m.startSettingChange(device, commands.SettingTypeScreenSize)
	default:
		// Fallback to screenshot
		return m.executeScreenshot(device)
	}
}

// executeScreenRecord runs the screen recording command
func (m Model) executeScreenRecord(device adb.Device) (tea.Model, tea.Cmd) {
	m.mode = ModeMenu
	m.clearLogs()
	m.mediaFeature.StartRecording()
	m.operationStartTime = time.Now()

	return m, tea.Batch(startRecording(m.config, device), doTick())
}

// stopRecording stops the active recording and saves it
func (m Model) stopRecording() (tea.Model, tea.Cmd) {
	activeRecording := m.mediaFeature.GetActiveRecording()
	if activeRecording != nil {
		return m, stopAndSaveRecording(activeRecording)
	}
	m.mediaFeature.FinishRecording()
	return m, nil
}

// startSettingChange initiates setting change for the selected device
func (m Model) startSettingChange(device adb.Device, settingType commands.SettingType) (tea.Model, tea.Cmd) {
	m.selectedDeviceForAction = device
	m.textInputAction = string(settingType)

	return m, getCurrentSetting(m.config, device, settingType)
}

// handleTextInputSubmit handles submission of text input
func (m Model) handleTextInputSubmit() (tea.Model, tea.Cmd) {
	settingType := commands.SettingType(m.textInputAction)
	switch settingType {
	case commands.SettingTypeDPI, commands.SettingTypeFontSize, commands.SettingTypeScreenSize:
		return m.executeSettingChange(settingType)
	}

	// Handle WiFi actions
	switch m.textInputAction {
	case "wifi_connect":
		return m.executeWiFiConnect()
	case "wifi_disconnect":
		return m.executeWiFiDisconnect()
	case "wifi_pair_address":
		return m.handlePairingAddressInput()
	case "wifi_pair_code":
		return m.executeWiFiPair()
	}

	// Reset to menu if unknown action
	m.mode = ModeMenu
	m.textInput = ""
	m.textInputPrompt = ""
	m.textInputAction = ""
	return m, nil
}

// executeSettingChange processes the setting change
func (m Model) executeSettingChange(settingType commands.SettingType) (tea.Model, tea.Cmd) {
	// Stay in text input mode, just clear the input and send the command

	// Save input and clear it
	input := m.textInput
	m.textInput = ""

	return m, changeSetting(m.config, m.selectedDeviceForAction, settingType, input)
}

// executeWiFiConnect processes WiFi connection
func (m Model) executeWiFiConnect() (tea.Model, tea.Cmd) {
	m.mode = ModeMenu
	m.clearLogs()
	m.connectingWiFi = true
	m.operationStartTime = time.Now()

	// Save input and clear it
	input := m.textInput
	m.textInput = ""
	m.textInputPrompt = ""
	m.textInputAction = ""

	return m, tea.Batch(connectWiFi(m.config, input), doTick())
}

// executeWiFiDisconnect processes WiFi disconnection
func (m Model) executeWiFiDisconnect() (tea.Model, tea.Cmd) {
	m.mode = ModeMenu
	m.clearLogs()
	m.disconnectingWiFi = true
	m.operationStartTime = time.Now()

	// Save input and clear it
	input := m.textInput
	m.textInput = ""
	m.textInputPrompt = ""
	m.textInputAction = ""

	return m, tea.Batch(disconnectWiFi(m.config, input), doTick())
}

// handlePairingAddressInput processes the first step of pairing (address input)
func (m Model) handlePairingAddressInput() (tea.Model, tea.Cmd) {
	// Store the pairing address and ask for pairing code
	m.pairingAddress = m.textInput
	m.textInput = ""
	m.textInputPrompt = fmt.Sprintf("Enter 6-digit pairing code from phone for %s", m.pairingAddress)
	m.textInputAction = "wifi_pair_code"

	return m, nil
}

// executeWiFiPair processes the second step of pairing (code input and execution)
func (m Model) executeWiFiPair() (tea.Model, tea.Cmd) {
	m.mode = ModeMenu
	m.clearLogs()
	m.pairingWiFi = true
	m.operationStartTime = time.Now()

	// Save pairing code and clear inputs
	pairingCode := m.textInput
	pairingAddress := m.pairingAddress
	m.textInput = ""
	m.textInputPrompt = ""
	m.textInputAction = ""
	m.pairingAddress = ""

	return m, tea.Batch(pairWiFi(m.config, pairingAddress, pairingCode), doTick())
}

// launchEmulator starts the selected emulator
func (m Model) launchEmulator(avd emulator.AVD) (tea.Model, tea.Cmd) {
	m.mode = ModeMenu
	m.err = nil
	m.successMsg = ""

	err := emulator.LaunchEmulator(m.config, avd)
	if err != nil {
		m.err = fmt.Errorf("failed to launch emulator: %v", err)
		return m, nil
	} else {
		m.addSuccess(fmt.Sprintf("Launched emulator: %s (may take a moment to appear)", avd.Name))
		// Refresh device list after launching emulator (it may take time to connect)
		return m, loadDevices(m.config.GetADBPath())
	}
}

// View renders the TUI
func (m Model) View() string {
	if m.quitting {
		return "Goodbye!\n"
	}

	var s strings.Builder

	// Header
	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("86")).
		Render("Android Tools CLI")

	s.WriteString(header + "\n")

	// Status bar
	statusBar := m.renderStatusBar()
	if statusBar != "" {
		s.WriteString(statusBar + "\n")
	}
	s.WriteString("\n")

	// Error display
	if m.err != nil {
		errorStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true)
		s.WriteString(errorStyle.Render("Error: "+m.err.Error()) + "\n\n")
	}

	switch m.mode {
	case ModeMenu:
		s.WriteString(m.renderMainMenu())
	case ModeDeviceSelect:
		s.WriteString(m.renderDeviceSelection())
	case ModeEmulatorSelect:
		s.WriteString(m.renderEmulatorSelection())
	case ModeTextInput:
		s.WriteString(m.renderTextInput())
	}

	// Progress indicators at bottom
	progressIndicators := m.renderProgressIndicators()
	if progressIndicators != "" {
		s.WriteString("\n" + progressIndicators + "\n")
	}

	// Log history display at bottom (only for main menu)
	if len(m.logHistory) > 0 && m.mode == ModeMenu {
		s.WriteString("\n" + m.renderLogHistory() + "\n")
	}

	// Footer
	var footerText string
	switch m.mode {
	case ModeMenu:
		if m.searchFilter != "" {
			footerText = "Type to search • ↑↓ to navigate • Enter to select • Esc to clear filter • Ctrl+C to quit"
		} else {
			footerText = "Type to search • ↑↓ to navigate • Enter to select • Ctrl+C to quit"
		}
	default:
		footerText = "Press Ctrl+C to quit"
	}

	footer := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Render(footerText)
	s.WriteString("\n" + footer)

	return s.String()
}

// renderMainMenu renders the main menu
func (m Model) renderMainMenu() string {
	var s strings.Builder

	// Header
	s.WriteString("Available commands")
	if m.searchFilter != "" {
		s.WriteString(fmt.Sprintf(" (filter: %s)", m.searchFilter))
	}
	s.WriteString(":\n\n")

	if m.searchFilter == "" {
		// Show categorized commands when no filter
		categories := getCommandCategories()
		currentIndex := 0

		for _, category := range categories {
			// Category header
			categoryStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("86")).
				Bold(true)
			s.WriteString(categoryStyle.Render(category.Name) + "\n")

			// Commands in category
			for _, cmd := range category.Commands {
				cursor := "  "
				if currentIndex == m.selectedCommandIndex {
					cursor = "> "
				}
				s.WriteString(fmt.Sprintf("%s%s\n", cursor, cmd.Name))
				currentIndex++
			}
			s.WriteString("\n")
		}
	} else {
		// Show filtered commands
		for i, cmd := range m.filteredCommands {
			cursor := "  "
			if i == m.selectedCommandIndex {
				cursor = "> "
			}
			s.WriteString(fmt.Sprintf("%s%s\n", cursor, cmd.Name))
		}

		if len(m.filteredCommands) == 0 {
			s.WriteString("  No matching commands\n")
		}
		s.WriteString("\n")
	}

	// Show description of selected command
	if len(m.filteredCommands) > 0 && m.selectedCommandIndex < len(m.filteredCommands) {
		selectedCmd := m.filteredCommands[m.selectedCommandIndex]
		descStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("245")).
			Italic(true)
		s.WriteString(descStyle.Render(fmt.Sprintf("→ %s", selectedCmd.Description)) + "\n\n")
	}

	s.WriteString(fmt.Sprintf("Connected devices: %d\n", len(m.devices)))

	for _, device := range m.devices {
		s.WriteString(fmt.Sprintf("  %s %s\n", device.GetStatusIndicator(), device.String()))
	}

	return s.String()
}

// renderDeviceSelection renders the device selection screen
func (m Model) renderDeviceSelection() string {
	s := []string{"Select a device:", ""}

	for i, device := range m.devices {
		cursor := "  "
		if i == m.selectedDevice {
			cursor = "> "
		}
		deviceInfo := fmt.Sprintf("%s %s", device.GetStatusIndicator(), device.String())
		extendedInfo := device.GetExtendedInfo()
		if extendedInfo != "" {
			deviceInfo += fmt.Sprintf("\n    %s", extendedInfo)
		}
		s = append(s, fmt.Sprintf("%s%s", cursor, deviceInfo))
	}

	s = append(s, "", "Press Enter to select, Esc to go back")
	return strings.Join(s, "\n")
}

// renderTextInput renders the text input screen
func (m Model) renderTextInput() string {
	s := []string{m.textInputPrompt, ""}

	// Show text input with cursor
	inputStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("86")).
		Bold(true)

	input := m.textInput + "█" // Simple cursor
	s = append(s, inputStyle.Render("> "+input))

	s = append(s, "", "Press Enter to submit, Esc to cancel")

	// Show log history at bottom if available
	if len(m.logHistory) > 0 {
		s = append(s, "", m.renderLogHistory())
	}

	return strings.Join(s, "\n")
}

// renderEmulatorSelection renders the emulator selection screen
func (m Model) renderEmulatorSelection() string {
	s := []string{"Select an emulator to launch:", ""}

	if len(m.avds) == 0 {
		s = append(s, "No AVDs found. Create one with Android Studio or avdmanager.")
		s = append(s, "", "Press Esc to go back")
		return strings.Join(s, "\n")
	}

	for i, avd := range m.avds {
		cursor := "  "
		if i == m.selectedEmulator {
			cursor = "> "
		}
		s = append(s, fmt.Sprintf("%s%s", cursor, avd.String()))
	}

	s = append(s, "", "Press Enter to launch, Esc to go back")
	return strings.Join(s, "\n")
}

// renderStatusBar renders the status bar showing filter, device count, and active operations
func (m Model) renderStatusBar() string {
	var statusItems []string

	// Device count with status indicators
	if len(m.devices) > 0 {
		var deviceCounts []string
		physicalCount, emulatorCount, wifiCount := 0, 0, 0

		for _, device := range m.devices {
			switch device.GetConnectionType() {
			case adb.DeviceTypePhysical:
				physicalCount++
			case adb.DeviceTypeEmulator:
				emulatorCount++
			case adb.DeviceTypeWiFi:
				wifiCount++
			}
		}

		if physicalCount > 0 {
			deviceCounts = append(deviceCounts, fmt.Sprintf("🔵 %d", physicalCount))
		}
		if emulatorCount > 0 {
			deviceCounts = append(deviceCounts, fmt.Sprintf("🟡 %d", emulatorCount))
		}
		if wifiCount > 0 {
			deviceCounts = append(deviceCounts, fmt.Sprintf("🟢 %d", wifiCount))
		}

		if len(deviceCounts) > 0 {
			statusItems = append(statusItems, fmt.Sprintf("Devices: %s", strings.Join(deviceCounts, " ")))
		}
	} else {
		statusItems = append(statusItems, "No devices connected")
	}

	// Active filter
	if m.searchFilter != "" {
		statusItems = append(statusItems, fmt.Sprintf("Filter: '%s'", m.searchFilter))
		if len(m.filteredCommands) > 0 {
			statusItems = append(statusItems, fmt.Sprintf("Commands: %d/%d", len(m.filteredCommands), len(getAvailableCommands())))
		} else {
			statusItems = append(statusItems, "No matching commands")
		}
	}

	// Active operations
	var activeOps []string
	if m.mediaFeature.IsTakingScreenshot() {
		activeOps = append(activeOps, "📸 Screenshot")
	}
	if m.mediaFeature.IsTakingDayNight() {
		activeOps = append(activeOps, "📸 Day-Night")
	}
	if m.mediaFeature.IsRecording() {
		activeOps = append(activeOps, "🎥 Recording")
	}
	if m.connectingWiFi {
		activeOps = append(activeOps, "📶 Connecting")
	}
	if m.disconnectingWiFi {
		activeOps = append(activeOps, "📶 Disconnecting")
	}
	if m.pairingWiFi {
		activeOps = append(activeOps, "📶 Pairing")
	}

	if len(activeOps) > 0 {
		statusItems = append(statusItems, fmt.Sprintf("Active: %s", strings.Join(activeOps, ", ")))
	}

	if len(statusItems) == 0 {
		return ""
	}

	// Style the status bar
	statusStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("245")).
		Background(lipgloss.Color("236")).
		Padding(0, 1)

	return statusStyle.Render(strings.Join(statusItems, " • "))
}

// getProgressText returns animated progress text with elapsed time
func (m Model) getProgressText(operation string) string {
	// Animated spinner
	spinners := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	spinner := spinners[m.progressTicker%len(spinners)]

	// Calculate elapsed time
	elapsed := time.Since(m.operationStartTime)
	var timeStr string
	if elapsed >= time.Minute {
		timeStr = fmt.Sprintf("(%dm%02ds)", int(elapsed.Minutes()), int(elapsed.Seconds())%60)
	} else {
		timeStr = fmt.Sprintf("(%.1fs)", elapsed.Seconds())
	}

	return fmt.Sprintf("%s %s %s", spinner, operation, timeStr)
}

// renderLogHistory renders the log history with proper formatting and styling
func (m Model) renderLogHistory() string {
	if len(m.logHistory) == 0 {
		return ""
	}

	var logLines []string

	for _, entry := range m.logHistory {
		var style lipgloss.Style
		var prefix string

		switch entry.Type {
		case LogTypeSuccess:
			style = core.SuccessStyle
			prefix = "✓"
		case LogTypeError:
			style = core.ErrorStyle
			prefix = "✗"
		case LogTypeInfo:
			style = core.InfoStyle
			prefix = "•"
		}

		// Format timestamp (show only time for recent entries)
		timeStr := entry.Timestamp.Format("15:04:05")

		// Handle multi-line messages by indenting continuation lines
		lines := strings.Split(entry.Message, "\n")
		for i, line := range lines {
			if i == 0 {
				// First line with timestamp and prefix
				formattedLine := fmt.Sprintf("[%s] %s %s", timeStr, prefix, strings.TrimSpace(line))
				logLines = append(logLines, style.Render(formattedLine))
			} else if strings.TrimSpace(line) != "" {
				// Continuation lines with single space indentation
				indentedLine := fmt.Sprintf(" %s", strings.TrimSpace(line))
				logLines = append(logLines, style.Render(indentedLine))
			}
		}
	}

	// Join all lines and add some spacing
	logStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("238")).
		Padding(0, 1).
		Margin(0, 0)

	return logStyle.Render(strings.Join(logLines, "\n"))
}

// renderProgressIndicators renders all active progress indicators
func (m Model) renderProgressIndicators() string {
	var indicators []string

	loadingStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("214")).
		Bold(true)

	if m.mediaFeature.IsTakingScreenshot() {
		progressText := m.getProgressText("Taking screenshot")
		indicators = append(indicators, loadingStyle.Render(progressText))
	}

	if m.mediaFeature.IsTakingDayNight() {
		progressText := m.getProgressText("Taking day-night screenshots")
		indicators = append(indicators, loadingStyle.Render(progressText))
	}

	if m.mediaFeature.IsRecording() {
		progressText := m.getProgressText("Recording screen • Press Esc to stop")
		indicators = append(indicators, loadingStyle.Render(progressText))
	}

	if m.connectingWiFi {
		progressText := m.getProgressText("Connecting to WiFi device")
		indicators = append(indicators, loadingStyle.Render(progressText))
	}

	if m.disconnectingWiFi {
		progressText := m.getProgressText("Disconnecting from WiFi device")
		indicators = append(indicators, loadingStyle.Render(progressText))
	}

	if m.pairingWiFi {
		progressText := m.getProgressText("Pairing with WiFi device")
		indicators = append(indicators, loadingStyle.Render(progressText))
	}

	if len(indicators) == 0 {
		return ""
	}

	return strings.Join(indicators, "\n")
}
