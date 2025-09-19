package logger

import (
	"os"
	"strings"
	"sync"
)

// LoggerService provides a singleton logger instance with feature flag integration
type LoggerService struct {
	logger     Logger
	minLevel   LogLevel
	enabled    bool
	once       sync.Once
	configured bool
}

var (
	instance *LoggerService
	once     sync.Once
)

// GetInstance returns the singleton logger service instance
func GetInstance() *LoggerService {
	once.Do(func() {
		instance = &LoggerService{}
		instance.configure()
	})
	return instance
}

// configure initializes the logger service based on environment variables
func (ls *LoggerService) configure() {
	ls.once.Do(func() {
		// Check if logging is enabled via environment variable
		logLevel := strings.ToUpper(strings.TrimSpace(os.Getenv("LOG_LEVEL")))

		if logLevel == "" || logLevel == "NONE" {
			ls.enabled = false
			ls.configured = true
			return
		}

		// Set minimum log level
		switch logLevel {
		case "ALL", "DEBUG":
			ls.minLevel = DebugLevel
		case "INFO":
			ls.minLevel = InfoLevel
		case "WARN", "WARNING":
			ls.minLevel = WarnLevel
		case "ERROR":
			ls.minLevel = ErrorLevel
		default:
			// For unknown log levels, default to INFO
			ls.minLevel = InfoLevel
		}

		// Check if colors should be disabled (useful for production logs)
		enableColors := !strings.EqualFold(os.Getenv("LOG_DISABLE_COLORS"), "true")

		// Check if emojis should be disabled
		enableEmojis := !strings.EqualFold(os.Getenv("LOG_DISABLE_EMOJIS"), "true")

		// Initialize the console logger
		ls.logger = NewConsoleLoggerWithOptions(enableColors, enableEmojis)
		ls.enabled = true
		ls.configured = true
	})
}

// Debug logs a debug message if the service is enabled and level permits
func (ls *LoggerService) Debug(message string, context *LogContext) {
	if ls.shouldLog(DebugLevel) {
		ls.logger.Debug(message, context)
	}
}

// Info logs an info message if the service is enabled and level permits
func (ls *LoggerService) Info(message string, context *LogContext) {
	if ls.shouldLog(InfoLevel) {
		ls.logger.Info(message, context)
	}
}

// Warn logs a warning message if the service is enabled and level permits
func (ls *LoggerService) Warn(message string, context *LogContext) {
	if ls.shouldLog(WarnLevel) {
		ls.logger.Warn(message, context)
	}
}

// Error logs an error message if the service is enabled and level permits
func (ls *LoggerService) Error(message string, context *LogContext) {
	if ls.shouldLog(ErrorLevel) {
		ls.logger.Error(message, context)
	}
}

// IsEnabled returns whether logging is enabled for the given level
func (ls *LoggerService) IsEnabled(level LogLevel) bool {
	return ls.shouldLog(level)
}

// shouldLog determines if a message should be logged based on level and enabled state
func (ls *LoggerService) shouldLog(level LogLevel) bool {
	if !ls.configured {
		ls.configure()
	}
	return ls.enabled && level >= ls.minLevel
}
