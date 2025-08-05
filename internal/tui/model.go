package tui

import (
	"errors"
	"fmt"
	"gadget/internal/adb"
	"gadget/internal/commands"
	"gadget/internal/config"
	"gadget/internal/tui/core"
	"gadget/internal/tui/features/devices"
	"gadget/internal/tui/features/media"
	"gadget/internal/tui/features/settings"
	"gadget/internal/tui/features/wifi"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
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

// Delegate to core functions
func getAvailableCommands() []Command {
	return core.GetAvailableCommands()
}

func getCommandCategories() []CommandCategory {
	return core.GetCommandCategories()
}

// Model represents the TUI state
type Model struct {
	config          *config.Config
	selectedCommand int
	mode            Mode
	err             error
	quitting        bool

	logHistory    []LogEntry
	maxLogEntries int
	loading       bool

	devicesFeature          *devices.DevicesFeature
	mediaFeature            *media.MediaFeature
	wifiFeature             *wifi.WiFiFeature
	settingsFeature         *settings.SettingsFeature
	selectedDeviceForAction adb.Device

	textInputPrompt string
	textInputAction string

	searchFilter         string
	filteredCommands     []Command
	selectedCommandIndex int
	searchMode           bool

	operationStartTime time.Time

	keys      KeyMap
	help      help.Model
	textInput textinput.Model
	spinner   spinner.Model
}

// NewModel creates a new TUI model
func NewModel(cfg *config.Config) Model {
	m := Model{
		config:               cfg,
		mode:                 ModeMenu,
		selectedCommand:      0,
		loading:              true,
		searchFilter:         "",
		selectedCommandIndex: 0,
		searchMode:           false,
		logHistory:           make([]LogEntry, 0),
		maxLogEntries:        5, // Keep last 5 log entries
		operationStartTime:   time.Now(),
		devicesFeature:       devices.NewDevicesFeature(cfg),
		mediaFeature:         media.NewMediaFeature(cfg),
		wifiFeature:          wifi.NewWiFiFeature(cfg),
		settingsFeature:      settings.NewSettingsFeature(cfg),
	}

	m.keys = DefaultKeyMap()
	m.help = help.New()
	m.textInput = newTextInput()
	m.spinner = newSpinner()

	m.filteredCommands = m.filterCommands()
	return m
}

// newTextInput creates and configures a new text input component
func newTextInput() textinput.Model {
	ti := textinput.New()
	ti.Placeholder = "Enter value..."
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 50
	return ti
}

// newSpinner creates and configures a new spinner component
func newSpinner() spinner.Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	return s
}

// addLogEntry adds a new log entry and maintains the history limit
func (m *Model) addLogEntry(message string, logType LogType) {
	normalizedMessage := strings.TrimSpace(strings.ReplaceAll(message, "\t", "  "))

	entry := LogEntry{
		Message:   normalizedMessage,
		Type:      logType,
		Timestamp: time.Now(),
	}

	m.logHistory = append(m.logHistory, entry)

	if len(m.logHistory) > m.maxLogEntries {
		m.logHistory = m.logHistory[len(m.logHistory)-m.maxLogEntries:]
	}

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

// clearLogs clears all log entries
func (m *Model) clearLogs() {
	m.logHistory = make([]LogEntry, 0)
	m.err = nil
}

// CommandMatch holds a command and its match score
type CommandMatch struct {
	Command Command
	Score   int
}

// filterCommands applies fuzzy search to the command list and sorts by score
func (m Model) filterCommands() []Command {
	if !m.searchMode || m.searchFilter == "" || m.searchFilter == "/" {
		return getAvailableCommands()
	}

	var matches []CommandMatch
	filter := strings.ToLower(strings.TrimPrefix(m.searchFilter, "/"))

	for _, cmd := range getAvailableCommands() {
		if score := m.fuzzyMatchScore(cmd, filter); score > 0 {
			matches = append(matches, CommandMatch{Command: cmd, Score: score})
		}
	}

	sort.Slice(matches, func(i, j int) bool {
		return matches[i].Score > matches[j].Score
	})

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

// Init initializes the model (required by Bubble Tea)
func (m Model) Init() tea.Cmd {
	return tea.Batch(loadDevices(m.config), m.spinner.Tick)
}

// Update handles messages and updates the model state
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyPress(msg)
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	case devicesLoadedMsg:
		_, _, _, errorMsg := m.devicesFeature.HandleDevicesLoaded(msg)
		m.loading = false
		if errorMsg != "" {
			m.err = errors.New(errorMsg)
		} else {
			m.err = nil
		}
		// Don't clear success messages during auto-refresh
		return m, nil
	case avdsLoadedMsg:
		_, _, _, errorMsg := m.devicesFeature.HandleAvdsLoaded(msg)
		if errorMsg != "" {
			m.err = errors.New(errorMsg)
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
			m.err = errors.New(errorMsg)
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
		_, _, _, errorMsg := m.settingsFeature.HandleSettingLoaded(msg)
		if errorMsg != "" {
			m.err = fmt.Errorf("failed to get current setting: %s", errorMsg)
			m.mode = ModeMenu
		} else {
			m.mode = ModeTextInput
			m.textInput.Focus()

			// Show both Default/Physical and Current values
			settingInfo := m.settingsFeature.GetCurrentSettingInfo()
			displayInfo := fmt.Sprintf("Physical %s: %s\nCurrent %s: %s",
				settingInfo.DisplayName, settingInfo.Default,
				settingInfo.DisplayName, settingInfo.Current)

			// Set contextual placeholder based on setting type
			var placeholder string
			switch settingInfo.Type {
			case commands.SettingTypeDPI:
				placeholder = settingInfo.Current
			case commands.SettingTypeFontSize:
				placeholder = settingInfo.Current
			case commands.SettingTypeScreenSize:
				placeholder = settingInfo.Current
			default:
				placeholder = "Enter new value..."
			}
			m.textInput.Placeholder = placeholder

			m.textInputPrompt = fmt.Sprintf("Device: %s\n%s\n\n%s:",
				m.selectedDeviceForAction.Serial, displayInfo, settingInfo.DisplayName)
		}
		return m, nil
	case settingChangedMsg:
		_, cmd, successMsg, errorMsg := m.settingsFeature.HandleSettingChanged(msg, m.selectedDeviceForAction)
		if successMsg != "" {
			m.addSuccess(successMsg)
			// Refresh setting info to show updated values
			return m, cmd
		} else if errorMsg != "" {
			m.addError(errorMsg)
		}
		return m, nil
	case wifiConnectDoneMsg:
		_, _, successMsg, errorMsg := m.wifiFeature.HandleWiFiConnectDone(msg)
		if successMsg != "" {
			m.addSuccess(successMsg)
			m.mode = ModeMenu
			// Refresh device list after successful WiFi connection
			return m, loadDevices(m.config)
		} else if errorMsg != "" {
			m.addError(errorMsg)
		}
		return m, nil
	case wifiDisconnectDoneMsg:
		_, _, successMsg, errorMsg := m.wifiFeature.HandleWiFiDisconnectDone(msg)
		if successMsg != "" {
			m.addSuccess(successMsg)
			m.mode = ModeMenu
			// Refresh device list after successful WiFi disconnection
			return m, loadDevices(m.config)
		} else if errorMsg != "" {
			m.addError(errorMsg)
		}
		return m, nil
	case wifiPairDoneMsg:
		_, _, successMsg, errorMsg := m.wifiFeature.HandleWiFiPairDone(msg)
		if successMsg != "" {
			m.addSuccess(successMsg)
			m.mode = ModeMenu
			// Refresh device list after successful WiFi pairing
			return m, loadDevices(m.config)
		} else if errorMsg != "" {
			m.addError(errorMsg)
		}
		return m, nil
	case emulatorConfigureDoneMsg:
		if msg.Success {
			m.addSuccess(msg.Message)
		} else {
			m.addError(msg.Message)
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
	// Global key handling for quit
	if key.Matches(msg, m.keys.Quit) {
		return m, tea.Quit
	}

	// Global key handling for active recording
	if m.mediaFeature.IsRecording() && key.Matches(msg, m.keys.StopRecording) {
		return m.stopRecording()
	}

	switch m.mode {
	case ModeMenu:
		if key.Matches(msg, m.keys.Escape) {
			// Clear search mode and filter if active
			if m.searchMode {
				m.searchMode = false
				m.searchFilter = ""
				m.filteredCommands = m.filterCommands()
				m.selectedCommandIndex = 0
			}
			return m, nil
		} else if key.Matches(msg, m.keys.Up) {
			if m.selectedCommandIndex > 0 {
				m.selectedCommandIndex--
			}
			return m, nil
		} else if key.Matches(msg, m.keys.Down) {
			if m.selectedCommandIndex < len(m.filteredCommands)-1 {
				m.selectedCommandIndex++
			}
			return m, nil
		} else if key.Matches(msg, m.keys.Enter) {
			if len(m.filteredCommands) > 0 && m.selectedCommandIndex < len(m.filteredCommands) {
				return m.executeSelectedCommand()
			}
			return m, nil
		} else if key.Matches(msg, m.keys.Backspace) {
			if m.searchMode && len(m.searchFilter) > 0 {
				m.searchFilter = m.searchFilter[:len(m.searchFilter)-1]
				// Exit search mode if only "/" remains or filter becomes empty
				if m.searchFilter == "/" || m.searchFilter == "" {
					m.searchMode = false
					m.searchFilter = ""
				}
				m.filteredCommands = m.filterCommands()
				// Reset selection if it's out of bounds
				if m.selectedCommandIndex >= len(m.filteredCommands) {
					m.selectedCommandIndex = 0
				}
			}
			return m, nil
		} else {
			// Handle typing for search
			if len(msg.String()) == 1 {
				char := msg.String()
				if key.Matches(msg, m.keys.Search) && !m.searchMode {
					// Enter search mode
					m.searchMode = true
					m.searchFilter = "/"
					m.filteredCommands = m.filterCommands()
					m.selectedCommandIndex = 0
				} else if m.searchMode {
					// Add character to search filter when in search mode
					m.searchFilter += char
					m.filteredCommands = m.filterCommands()
					m.selectedCommandIndex = 0 // Reset to first item
				}
			}
			return m, nil
		}
	case ModeDeviceSelect:
		if key.Matches(msg, m.keys.Escape) {
			m.mode = ModeMenu
			return m, nil
		} else if key.Matches(msg, m.keys.VimUp) {
			selectedDevice := m.devicesFeature.GetSelectedDevice()
			if selectedDevice > 0 {
				m.devicesFeature.SetSelectedDevice(selectedDevice - 1)
			}
			return m, nil
		} else if key.Matches(msg, m.keys.VimDown) {
			selectedDevice := m.devicesFeature.GetSelectedDevice()
			if selectedDevice < len(m.devicesFeature.GetDevices())-1 {
				m.devicesFeature.SetSelectedDevice(selectedDevice + 1)
			}
			return m, nil
		} else if key.Matches(msg, m.keys.Enter) {
			selectedDevice := m.devicesFeature.GetSelectedDeviceInstance()
			if selectedDevice != nil {
				return m.executeCommandForDevice(*selectedDevice)
			}
		}
	case ModeEmulatorSelect:
		if key.Matches(msg, m.keys.Escape) {
			m.mode = ModeMenu
			return m, nil
		} else if key.Matches(msg, m.keys.VimUp) {
			selectedEmulator := m.devicesFeature.GetSelectedEmulator()
			if selectedEmulator > 0 {
				m.devicesFeature.SetSelectedEmulator(selectedEmulator - 1)
			}
			return m, nil
		} else if key.Matches(msg, m.keys.VimDown) {
			selectedEmulator := m.devicesFeature.GetSelectedEmulator()
			if selectedEmulator < len(m.devicesFeature.GetAvds())-1 {
				m.devicesFeature.SetSelectedEmulator(selectedEmulator + 1)
			}
			return m, nil
		} else if key.Matches(msg, m.keys.Enter) {
			return m.executeEmulatorCommand()
		}
	case ModeTextInput:
		if key.Matches(msg, m.keys.Submit) {
			return m.handleTextInputSubmit()
		} else if key.Matches(msg, m.keys.Cancel) {
			m.mode = ModeMenu
			m.textInput.SetValue("")
			m.textInputPrompt = ""
			m.textInputAction = ""
			return m, nil
		} else {
			// Delegate to textinput component
			var cmd tea.Cmd
			m.textInput, cmd = m.textInput.Update(msg)
			return m, cmd
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

	return m, tea.Batch(takeScreenshot(m.config, device), m.spinner.Tick)
}

// executeDayNightScreenshots runs the day-night screenshot command
func (m Model) executeDayNightScreenshots(device adb.Device) (tea.Model, tea.Cmd) {
	m.mode = ModeMenu
	m.clearLogs()
	m.mediaFeature.StartDayNightScreenshot()
	m.operationStartTime = time.Now()

	return m, tea.Batch(takeDayNightScreenshots(m.config, device), m.spinner.Tick)
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
	case "configure-emulator":
		m.mode = ModeEmulatorSelect
		return m, loadAVDs(m.config)
	case "connect-wifi":
		m.mode = ModeTextInput
		m.textInput.Focus()
		m.textInput.Placeholder = "192.168.1.100 or 192.168.1.100:5555 (defaults to port 4444)"
		m.textInputPrompt = "Connect to WiFi device"
		m.textInputAction = "wifi_connect"
		m.textInput.SetValue("")
		return m, nil
	case "pair-wifi":
		m.mode = ModeTextInput
		m.textInput.Focus()
		m.textInput.Placeholder = "192.168.3.30:43719 (from phone's pairing dialog)"
		m.textInputPrompt = "Pair with WiFi device"
		m.textInputAction = "wifi_pair_address"
		m.textInput.SetValue("")
		return m, nil
	case "disconnect-wifi":
		m.mode = ModeTextInput
		m.textInput.Focus()
		m.textInput.Placeholder = "192.168.1.100 or 192.168.1.100:5555 (defaults to port 4444)"
		m.textInputPrompt = "Disconnect from WiFi device"
		m.textInputAction = "wifi_disconnect"
		m.textInput.SetValue("")
		return m, nil
	case "refresh-devices":
		m.clearLogs()
		return m, loadDevices(m.config)
	default:
		// Commands that require device selection
		devices := m.devicesFeature.GetDevices()
		if len(devices) == 1 {
			return m.executeCommandForDevice(devices[0])
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
	case "dpi":
		return m.startSettingChange(device, commands.SettingTypeDPI)
	case "font-size":
		return m.startSettingChange(device, commands.SettingTypeFontSize)
	case "screen-size":
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

	return m, tea.Batch(startRecording(m.config, device), m.spinner.Tick)
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
	m.textInput.SetValue("")
	m.textInputPrompt = ""
	m.textInputAction = ""
	return m, nil
}

// executeSettingChange processes the setting change
func (m Model) executeSettingChange(settingType commands.SettingType) (tea.Model, tea.Cmd) {
	// Stay in text input mode, just clear the input and send the command

	// Save input and clear it
	input := m.textInput.Value()
	m.textInput.SetValue("")

	return m, changeSetting(m.config, m.selectedDeviceForAction, settingType, input)
}

// executeWiFiConnect processes WiFi connection
func (m Model) executeWiFiConnect() (tea.Model, tea.Cmd) {
	m.mode = ModeMenu
	m.clearLogs()
	m.operationStartTime = time.Now()

	// Save input and clear it
	input := m.textInput.Value()
	m.textInput.SetValue("")
	m.textInputPrompt = ""
	m.textInputAction = ""

	cmd := m.wifiFeature.StartWiFiConnect(input)
	return m, tea.Batch(cmd, m.spinner.Tick)
}

// executeWiFiDisconnect processes WiFi disconnection
func (m Model) executeWiFiDisconnect() (tea.Model, tea.Cmd) {
	m.mode = ModeMenu
	m.clearLogs()
	m.operationStartTime = time.Now()

	// Save input and clear it
	input := m.textInput.Value()
	m.textInput.SetValue("")
	m.textInputPrompt = ""
	m.textInputAction = ""

	cmd := m.wifiFeature.StartWiFiDisconnect(input)
	return m, tea.Batch(cmd, m.spinner.Tick)
}

// handlePairingAddressInput processes the first step of pairing (address input)
func (m Model) handlePairingAddressInput() (tea.Model, tea.Cmd) {
	// Store the pairing address and ask for pairing code
	m.wifiFeature.SetPairingAddress(m.textInput.Value())
	m.textInput.SetValue("")
	m.textInput.Focus()
	m.textInput.Placeholder = "123456 (6-digit code from phone)"
	m.textInputPrompt = fmt.Sprintf("Enter pairing code for %s", m.wifiFeature.GetPairingAddress())
	m.textInputAction = "wifi_pair_code"

	return m, nil
}

// executeWiFiPair processes the second step of pairing (code input and execution)
func (m Model) executeWiFiPair() (tea.Model, tea.Cmd) {
	m.mode = ModeMenu
	m.clearLogs()
	m.operationStartTime = time.Now()

	// Save pairing code and clear inputs
	pairingCode := m.textInput.Value()
	pairingAddress := m.wifiFeature.GetPairingAddress()
	m.textInput.SetValue("")
	m.textInputPrompt = ""
	m.textInputAction = ""
	m.wifiFeature.ClearPairingAddress()

	cmd := m.wifiFeature.StartWiFiPair(pairingAddress, pairingCode)
	return m, tea.Batch(cmd, m.spinner.Tick)
}

// launchEmulator starts the selected emulator
func (m Model) launchEmulator() (tea.Model, tea.Cmd) {
	m.mode = ModeMenu
	m.err = nil

	_, cmd, successMsg, errorMsg := m.devicesFeature.LaunchSelectedEmulator()
	if errorMsg != "" {
		m.err = errors.New(errorMsg)
		return m, nil
	} else if successMsg != "" {
		m.addSuccess(successMsg)
		return m, cmd
	}

	return m, nil
}

// executeEmulatorCommand executes the appropriate emulator command based on selection
func (m Model) executeEmulatorCommand() (tea.Model, tea.Cmd) {
	if len(m.filteredCommands) == 0 || m.selectedCommandIndex >= len(m.filteredCommands) {
		return m, nil
	}

	selectedCmd := m.filteredCommands[m.selectedCommandIndex]

	switch selectedCmd.Command {
	case "launch-emulator":
		return m.launchEmulator()
	case "configure-emulator":
		return m.configureEmulator()
	default:
		// Fallback to launch for unknown commands
		return m.launchEmulator()
	}
}

// configureEmulator opens the selected emulator's config in editor
func (m Model) configureEmulator() (tea.Model, tea.Cmd) {
	m.mode = ModeMenu
	m.err = nil

	selectedAVD := m.devicesFeature.GetSelectedEmulatorInstance()
	if selectedAVD == nil {
		m.err = fmt.Errorf("no emulator selected")
		return m, nil
	}

	return m, configureEmulatorCmd(m.config, *selectedAVD)
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
		Render("Go-go Gadget…")

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

	// Footer with help (only for modes that don't handle their own help)
	var helpKeys []key.Binding
	switch m.mode {
	case ModeMenu:
		helpKeys = m.keys.MenuKeys(m.searchMode)
	case ModeTextInput:
		helpKeys = m.keys.TextInputKeys()
	case ModeDeviceSelect, ModeEmulatorSelect:
		// These modes handle their own help display, skip global footer
		// But still show logs below everything
		if len(m.logHistory) > 0 {
			s.WriteString("\n" + m.renderLogHistory())
		}
		return s.String()
	default:
		helpKeys = []key.Binding{m.keys.Quit}
	}

	// Add recording-specific help if recording
	if m.mediaFeature.IsRecording() {
		helpKeys = m.keys.RecordingKeys()
	}

	footer := m.renderHelp(helpKeys)
	s.WriteString("\n\n" + footer)

	// Log history display at bottom (persistent across all screens)
	if len(m.logHistory) > 0 {
		s.WriteString("\n\n" + m.renderLogHistory())
	}

	return s.String()
}

// renderMainMenu renders the main menu
func (m Model) renderMainMenu() string {
	var s strings.Builder

	// Header - show search status if active
	if m.searchMode && m.searchFilter != "" {
		displayFilter := strings.TrimPrefix(m.searchFilter, "/")
		if displayFilter == "" {
			s.WriteString("(search mode: type to filter)\n\n")
		} else {
			s.WriteString(fmt.Sprintf("(filter: %s)\n\n", displayFilter))
		}
	}

	if !m.searchMode || m.searchFilter == "" || m.searchFilter == "/" {
		// Show categorized commands when not in search mode or no effective filter
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

	devices := m.devicesFeature.GetDevices()
	s.WriteString(fmt.Sprintf("Connected devices: %d\n", len(devices)))

	for _, device := range devices {
		s.WriteString(fmt.Sprintf("  %s %s", device.GetStatusIndicator(), device.String()))
		if extendedInfo := device.GetExtendedInfo(); extendedInfo != "" {
			s.WriteString(fmt.Sprintf("\n    %s", extendedInfo))
		}
		s.WriteString("\n")
	}

	return s.String()
}

// renderDeviceSelection renders the device selection screen
func (m Model) renderDeviceSelection() string {
	s := []string{"Select a device:", ""}

	devices := m.devicesFeature.GetDevices()
	selectedDevice := m.devicesFeature.GetSelectedDevice()

	for i, device := range devices {
		cursor := "  "
		if i == selectedDevice {
			cursor = "> "
		}
		deviceInfo := fmt.Sprintf("%s %s", device.GetStatusIndicator(), device.String())
		extendedInfo := device.GetExtendedInfo()
		if extendedInfo != "" {
			deviceInfo += fmt.Sprintf("\n    %s", extendedInfo)
		}
		s = append(s, fmt.Sprintf("%s%s", cursor, deviceInfo))
	}

	// Add help using bubbles help component
	s = append(s, "", "", m.renderHelp(m.keys.DeviceSelectKeys()))
	return strings.Join(s, "\n")
}

// renderTextInput renders the text input screen
func (m Model) renderTextInput() string {
	var s []string

	// Split prompt into lines
	promptLines := strings.Split(m.textInputPrompt, "\n")

	// Add all lines except the last one (which ends with ":")
	if len(promptLines) > 1 {
		s = append(s, promptLines[:len(promptLines)-1]...)
		s = append(s, "")
	}

	// Add the last line (the label) and input on separate lines with spacing
	lastLine := promptLines[len(promptLines)-1]
	s = append(s, lastLine)
	s = append(s, "")
	s = append(s, m.textInput.View())

	// Help is handled by global footer, don't duplicate it here
	// Log history is now handled globally at the bottom

	return strings.Join(s, "\n")
}

// renderEmulatorSelection renders the emulator selection screen
func (m Model) renderEmulatorSelection() string {
	s := []string{"Select an emulator to launch:", ""}

	avds := m.devicesFeature.GetAvds()
	selectedEmulator := m.devicesFeature.GetSelectedEmulator()

	if len(avds) == 0 {
		s = append(s, "No AVDs found. Create one with Android Studio or avdmanager.")
		// For no AVDs case, just show escape key
		s = append(s, "", "", m.renderHelp([]key.Binding{m.keys.EscapeBack}))
		return strings.Join(s, "\n")
	}

	for i, avd := range avds {
		cursor := "  "
		if i == selectedEmulator {
			cursor = "> "
		}
		avdInfo := fmt.Sprintf("%s%s", cursor, avd.String())
		if extendedInfo := avd.GetExtendedInfo(); extendedInfo != "" {
			avdInfo += fmt.Sprintf("\n    %s", extendedInfo)
		}
		s = append(s, avdInfo)
	}

	// Add help using bubbles help component
	s = append(s, "", "", m.renderHelp(m.keys.EmulatorSelectKeys()))
	return strings.Join(s, "\n")
}

// renderStatusBar renders the status bar showing filter, device count, and active operations
func (m Model) renderStatusBar() string {
	var statusItems []string

	// Device count with status indicators
	devices := m.devicesFeature.GetDevices()
	if len(devices) > 0 {
		var deviceCounts []string
		physicalCount, emulatorCount, wifiCount := 0, 0, 0

		for _, device := range devices {
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
	if m.searchMode {
		if m.searchFilter == "/" {
			statusItems = append(statusItems, "Search mode active")
		} else if len(m.searchFilter) > 1 {
			displayFilter := strings.TrimPrefix(m.searchFilter, "/")
			statusItems = append(statusItems, fmt.Sprintf("Filter: '%s'", displayFilter))
			if len(m.filteredCommands) > 0 {
				statusItems = append(statusItems, fmt.Sprintf("Commands: %d/%d", len(m.filteredCommands), len(getAvailableCommands())))
			} else {
				statusItems = append(statusItems, "No matching commands")
			}
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
	if m.wifiFeature.IsConnecting() {
		activeOps = append(activeOps, "📶 Connecting")
	}
	if m.wifiFeature.IsDisconnecting() {
		activeOps = append(activeOps, "📶 Disconnecting")
	}
	if m.wifiFeature.IsPairing() {
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
	// Calculate elapsed time
	elapsed := time.Since(m.operationStartTime)
	var timeStr string
	if elapsed >= time.Minute {
		timeStr = fmt.Sprintf("(%dm%02ds)", int(elapsed.Minutes()), int(elapsed.Seconds())%60)
	} else {
		timeStr = fmt.Sprintf("(%.1fs)", elapsed.Seconds())
	}

	return fmt.Sprintf("%s %s %s", m.spinner.View(), operation, timeStr)
}

// renderHelp renders help text using the bubbles help component with consistent styling
func (m Model) renderHelp(keys []key.Binding) string {
	helpView := m.help.ShortHelpView(keys)
	return lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render(helpView)
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

	if m.wifiFeature.IsConnecting() {
		progressText := m.getProgressText("Connecting to WiFi device")
		indicators = append(indicators, loadingStyle.Render(progressText))
	}

	if m.wifiFeature.IsDisconnecting() {
		progressText := m.getProgressText("Disconnecting from WiFi device")
		indicators = append(indicators, loadingStyle.Render(progressText))
	}

	if m.wifiFeature.IsPairing() {
		progressText := m.getProgressText("Pairing with WiFi device")
		indicators = append(indicators, loadingStyle.Render(progressText))
	}

	if len(indicators) == 0 {
		return ""
	}

	return strings.Join(indicators, "\n")
}
