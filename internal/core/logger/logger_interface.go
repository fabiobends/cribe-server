package logger

// LogLevel represents the severity level of a log message
type LogLevel int

const (
	DebugLevel LogLevel = iota
	InfoLevel
	WarnLevel
	ErrorLevel
)

// String returns the string representation of the log level
func (l LogLevel) String() string {
	switch l {
	case DebugLevel:
		return "DEBUG"
	case InfoLevel:
		return "INFO"
	case WarnLevel:
		return "WARN"
	case ErrorLevel:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// EntityType represents the type of entity that is logging
type EntityType string

const (
	HandlerEntity    EntityType = "HANDLER"
	ServiceEntity    EntityType = "SERVICE"
	RepositoryEntity EntityType = "REPOSITORY"
	ModelEntity      EntityType = "MODEL"
	MiddlewareEntity EntityType = "MIDDLEWARE"
	UtilEntity       EntityType = "UTIL"
	CoreEntity       EntityType = "CORE"
)

// LogContext contains metadata for a log message
type LogContext struct {
	EntityType   EntityType
	EntityName   string
	FunctionName string
	Extra        map[string]interface{}
}

// Logger defines the interface for all logger implementations
// This interface is platform-agnostic and can be implemented by
// console loggers, Sentry, structured loggers, etc.
type Logger interface {
	// Debug logs debug-level messages (lowest priority)
	Debug(message string, context *LogContext)

	// Info logs informational messages
	Info(message string, context *LogContext)

	// Warn logs warning messages
	Warn(message string, context *LogContext)

	// Error logs error messages (highest priority)
	Error(message string, context *LogContext)

	// IsEnabled checks if a specific log level is enabled
	IsEnabled(level LogLevel) bool
}
