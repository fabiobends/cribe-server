package logger

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"
)

// ConsoleLogger implements the Logger interface for console output
// with ANSI colors and emoji indicators
type ConsoleLogger struct {
	enableColors bool
	enableEmojis bool
}

// NewConsoleLogger creates a new console logger instance with default options
func NewConsoleLogger() *ConsoleLogger {
	return &ConsoleLogger{
		enableColors: true,
		enableEmojis: true,
	}
}

// NewConsoleLoggerWithOptions creates a console logger with custom options
func NewConsoleLoggerWithOptions(enableColors, enableEmojis bool) *ConsoleLogger {
	return &ConsoleLogger{
		enableColors: enableColors,
		enableEmojis: enableEmojis,
	}
}

// Debug logs debug-level messages
func (c *ConsoleLogger) Debug(message string, context *LogContext) {
	c.log(DebugLevel, message, context)
}

// Info logs informational messages
func (c *ConsoleLogger) Info(message string, context *LogContext) {
	c.log(InfoLevel, message, context)
}

// Warn logs warning messages
func (c *ConsoleLogger) Warn(message string, context *LogContext) {
	c.log(WarnLevel, message, context)
}

// Error logs error messages
func (c *ConsoleLogger) Error(message string, context *LogContext) {
	c.log(ErrorLevel, message, context)
}

// IsEnabled checks if a specific log level is enabled
func (c *ConsoleLogger) IsEnabled(level LogLevel) bool {
	return true
}

// log is the internal logging method that handles formatting and output
func (c *ConsoleLogger) log(level LogLevel, message string, context *LogContext) {
	timestamp := time.Now().Format("15:04:05")

	// Mask sensitive data in the message
	maskedMessage := c.maskSensitiveData(message)

	// Get colors and emoji
	levelColor := ""
	resetColor := ""
	emoji := ""

	if c.enableColors {
		levelColor = GetLevelColor(level)
		resetColor = ColorReset
	}

	if c.enableEmojis {
		emoji = GetLevelEmoji(level)
	}

	// Build the log message components
	levelStr := level.String()

	var entityStr string
	if context != nil {
		entityName := context.EntityName
		if context.FunctionName != "" {
			entityName += "." + context.FunctionName
		}
		// Make entity type bold while keeping the level color for the whole line
		entityStr = fmt.Sprintf("%s%s%s:%s", StyleBold, context.EntityType, ColorReset+levelColor, entityName)
	} else {
		entityStr = fmt.Sprintf("%sUNKNOWN%s", StyleBold, ColorReset+levelColor)
	}

	// Format the main log line using the LogFormat constant
	logLine := fmt.Sprintf(LogFormat,
		levelColor, // Start level color
		timestamp,
		emoji+levelStr,
		entityStr,
		maskedMessage,
		resetColor, // Reset color at the end
	)

	// Add extra info on a separate line if present
	extraInfo := c.formatExtra(context)
	if extraInfo != "" {
		logLine += fmt.Sprintf("\n%s    %s%s", levelColor, extraInfo, resetColor)
	}

	// Output to appropriate stream
	if level >= ErrorLevel {
		fmt.Fprintln(os.Stderr, logLine)
	} else {
		fmt.Fprintln(os.Stdout, logLine)
	}
}

// maskSensitiveData masks sensitive information in log messages
func (c *ConsoleLogger) maskSensitiveData(message string) string {
	// Email masking with fixed length to prevent length guessing
	emailRegex := regexp.MustCompile(`\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}\b`)
	message = emailRegex.ReplaceAllStringFunc(message, func(email string) string {
		parts := strings.Split(email, "@")
		if len(parts) != 2 {
			return "****@****"
		}

		localPart := parts[0]
		domainPart := parts[1]

		// Fixed length masking to prevent length guessing
		maskedLocal := c.maskStringFixedLength(localPart, 4)
		maskedDomain := c.maskStringFixedLength(domainPart, 4)

		return maskedLocal + "@" + maskedDomain
	})

	// Password masking
	passwordRegex := regexp.MustCompile(`(?i)(password|pwd|pass|secret|token|key)["']?\s*[:=]\s*["']?[^\s,"']+`)
	message = passwordRegex.ReplaceAllString(message, "${1}: ****")

	// JWT Token masking
	jwtRegex := regexp.MustCompile(`[A-Za-z0-9_-]+\.[A-Za-z0-9_-]+\.[A-Za-z0-9_-]+`)
	message = jwtRegex.ReplaceAllString(message, "****")

	return message
}

// maskStringFixedLength masks a string with fixed length to prevent length guessing
func (c *ConsoleLogger) maskStringFixedLength(s string, showChars int) string {
	if len(s) <= showChars {
		return strings.Repeat("*", 4) // Always show 4 asterisks for short strings
	}
	return s[:showChars] + "****" // Always show exactly 4 asterisks regardless of actual length
}

// formatExtra formats the extra context information with sensitive data masking
func (c *ConsoleLogger) formatExtra(context *LogContext) string {
	if context == nil || len(context.Extra) == 0 {
		return ""
	}

	var parts []string
	for key, value := range context.Extra {
		// Convert value to string and mask sensitive data
		valueStr := fmt.Sprintf("%v", value)
		maskedValue := c.maskSensitiveData(valueStr)
		parts = append(parts, fmt.Sprintf("%s=%s", key, maskedValue))
	}

	return fmt.Sprintf("[%s]", strings.Join(parts, ", "))
}
