# ADX - Android debugging experience

A command-line Android debugging tool built in Go that provides both TUI (terminal user interface) and direct command-line interfaces for common Android development tasks.

## Usage

### TUI mode (interactive)

Run the TUI interface for interactive device and emulator management:

```bash
./adx
```

### CLI mode (direct commands)

Arguments can be passed positionally (in order) or with named flags:

```bash
# Positional arguments (in order)
./adx pair-wifi "192.168.1.100:5555" "123456"
./adx change-dpi "480"
./adx launch-emulator "Pixel_6_API_34"

# Named flags (any order)
./adx pair-wifi -ip "192.168.1.100:5555" -code "123456"
./adx change-dpi -value "480" -device "emulator-5554"
./adx launch-emulator -value "Pixel_6_API_34"

# Alternative command syntax
./adx -command pair-wifi -ip "192.168.1.100:5555" -code "123456"
./adx -command change-dpi -value "480"
```

## CLI commands

| Command | Description | Parameters |
|---------|-------------|------------|
| `screenshot` | Take device screenshot | `-device` (optional) |
| `screenshot-day-night` | Take screenshots in both light and dark themes | `-device` (optional) |
| `screen-record` | Record device screen (Ctrl+C to stop) | `-device` (optional) |
| `change-dpi` | Modify device DPI | `-value` (required), `-device` (optional) |
| `change-font-size` | Adjust system font scaling | `-value` (required), `-device` (optional) |
| `change-screen-size` | Change display resolution | `-value` (required), `-device` (optional) |
| `launch-emulator` | Start Android emulator | `-value` (AVD name, optional) |
| `configure-emulator` | Edit emulator configuration in $EDITOR | `-value` (AVD name, optional) |
| `pair-wifi` | Pair device over WiFi | `-ip` (required), `-code` (required) |
| `connect-wifi` | Connect to WiFi ADB device | `-ip` (required) |
| `disconnect-wifi` | Disconnect from WiFi ADB device | `-ip` (required) |
| `refresh-devices` | List connected devices with extended info | None |

## Setup

### Prerequisites

- Android SDK installed with `ANDROID_HOME` or `ANDROID_SDK_ROOT` environment variable set
- ADB (Android Debug Bridge) available in PATH or SDK tools directory
- Go 1.24+ for building from source

### Build

```bash
go build -o adx
```

### Configuration

The tool automatically detects your Android SDK installation:
- Environment variables: `ANDROID_HOME` or `ANDROID_SDK_ROOT`
- Default macOS location: `~/Library/Android/sdk`
- Media files saved to: `~/Downloads` (configurable)

## Development

### External libraries

| Library | Purpose |
|---------|---------|
| [Bubble Tea](https://github.com/charmbracelet/bubbletea) | Terminal UI framework for interactive TUI interface |
| [Lip Gloss](https://github.com/charmbracelet/lipgloss) | Terminal styling and layout for consistent UI rendering |
| [Cobra](https://github.com/spf13/cobra) | CLI framework for command parsing and flag handling |

### Architecture

Built with a feature-based architecture prioritizing vertical slices over horizontal layers:

- **CLI/TUI dual interface**: Commands work in both interactive TUI and direct CLI modes
- **Device management**: Unified handling of physical devices and emulators with extended info
- **Command system**: Modular, registry-based command execution with validation
- **Configuration**: Environment-aware Android SDK detection with sensible defaults

### Development commands

```bash
# Install dependencies
go mod tidy

# Format code
gofmt -s -w .

# Build binary
go build -o adx

# Run tests (when implemented)
go test ./...
```
