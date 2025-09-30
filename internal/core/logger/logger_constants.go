package logger

// ANSI color codes for terminal output
const (
	// Reset
	ColorReset = "\033[0m"

	// Text colors
	ColorWhite     = "\033[0;37m"
	ColorLightBlue = "\033[0;36m"
	ColorYellow    = "\033[0;33m"
	ColorRed       = "\033[0;31m"

	// Text styles
	StyleBold = "\033[1m"
	StyleDim  = "\033[2m"
)

// Log level emojis for visual identification
const (
	DebugEmoji = "🔍"
	InfoEmoji  = "ℹ️"
	WarnEmoji  = "⚠️"
	ErrorEmoji = "❌"
)

// Default log format template
const LogFormat = "%s%s %s %s %s%s"

// GetLevelColor returns the ANSI color code for a log level
func GetLevelColor(level LogLevel) string {
	switch level {
	case DebugLevel:
		return ColorLightBlue
	case InfoLevel:
		return ColorWhite
	case WarnLevel:
		return ColorYellow
	case ErrorLevel:
		return ColorRed
	default:
		return ColorReset
	}
}

// GetLevelEmoji returns the emoji for a log level
func GetLevelEmoji(level LogLevel) string {
	switch level {
	case DebugLevel:
		return DebugEmoji
	case InfoLevel:
		return InfoEmoji
	case WarnLevel:
		return WarnEmoji
	case ErrorLevel:
		return ErrorEmoji
	default:
		return "❓"
	}
}

// GetEntityColor returns a consistent color for an entity type
func GetEntityColor(entityType EntityType) string {
	switch entityType {
	case HandlerEntity:
		return "\033[1;35m" // Magenta
	case ServiceEntity:
		return "\033[1;32m" // Green
	case RepositoryEntity:
		return "\033[1;34m" // Blue
	case ModelEntity:
		return "\033[1;93m" // Bright yellow
	case MiddlewareEntity:
		return "\033[1;36m" // Cyan
	case UtilEntity:
		return "\033[1;90m" // Bright black
	case CoreEntity:
		return "\033[1;37m" // White
	default:
		return ColorReset
	}
}
