package display

const (
	// Display icons for device information
	IconCPU        = "🔧"
	IconScreen     = "📱"
	IconNetwork    = "🌐"
	IconBattery    = "🔋"
	IconBatteryLow = "🪫"
)

// NormalizeCPUArchitecture converts technical CPU architecture names to user-friendly display names
func NormalizeCPUArchitecture(arch string) string {
	switch arch {
	case "arm64-v8a":
		return "ARM64"
	case "armeabi-v7a":
		return "ARM32"
	case "x86_64":
		return "x64"
	case "x86":
		return "x86"
	default:
		return arch
	}
}

// FormatExtendedInfoWithIndent formats extended info with consistent indentation
func FormatExtendedInfoWithIndent(mainInfo, extendedInfo string) string {
	if extendedInfo == "" {
		return mainInfo
	}
	return mainInfo + "\n    " + extendedInfo
}
