package logger

import (
	"os"
	"sync"
	"testing"
)

func TestLoggerServiceEnabled(t *testing.T) {
	tests := []struct {
		name        string
		logLevel    string
		expected    bool
		description string
	}{
		{
			name:        "EmptyLogLevel",
			logLevel:    "",
			expected:    true,
			description: "Empty LOG_LEVEL should default to INFO level behavior",
		},
		{
			name:        "NoneLogLevel",
			logLevel:    "NONE",
			expected:    false,
			description: "NONE LOG_LEVEL should disable logging",
		},
		{
			name:        "InfoLogLevel",
			logLevel:    "INFO",
			expected:    true,
			description: "INFO LOG_LEVEL should enable logging",
		},
		{
			name:        "DebugLogLevel",
			logLevel:    "DEBUG",
			expected:    true,
			description: "DEBUG LOG_LEVEL should enable logging",
		},
		{
			name:        "ErrorLogLevel",
			logLevel:    "ERROR",
			expected:    true,
			description: "ERROR LOG_LEVEL should enable logging",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset the singleton for each test
			instance = nil
			once = sync.Once{}

			// Set the environment variable
			_ = os.Setenv("LOG_LEVEL", tt.logLevel)
			defer func() { _ = os.Unsetenv("LOG_LEVEL") }()

			// Get the logger service instance
			loggerService := GetInstance()

			// Test IsEnabled method for different levels
			if tt.expected {
				// When logging is enabled, check that specific levels work correctly
				if !loggerService.IsEnabled(InfoLevel) && tt.logLevel != "ERROR" {
					t.Errorf("%s: expected IsEnabled(InfoLevel) = true when logging is enabled", tt.description)
				}
			} else {
				// When logging is disabled, all levels should return false
				if loggerService.IsEnabled(DebugLevel) {
					t.Errorf("%s: expected IsEnabled(DebugLevel) = false when logging is disabled", tt.description)
				}
				if loggerService.IsEnabled(InfoLevel) {
					t.Errorf("%s: expected IsEnabled(InfoLevel) = false when logging is disabled", tt.description)
				}
				if loggerService.IsEnabled(WarnLevel) {
					t.Errorf("%s: expected IsEnabled(WarnLevel) = false when logging is disabled", tt.description)
				}
				if loggerService.IsEnabled(ErrorLevel) {
					t.Errorf("%s: expected IsEnabled(ErrorLevel) = false when logging is disabled", tt.description)
				}
			}
		})
	}
}

func TestLoggerService_LoggingMethods(t *testing.T) {
	// Reset the singleton
	instance = nil
	once = sync.Once{}

	// Set log level to DEBUG to enable all logging
	_ = os.Setenv("LOG_LEVEL", "DEBUG")
	defer func() { _ = os.Unsetenv("LOG_LEVEL") }()

	loggerService := GetInstance()

	context := &LogContext{
		EntityType:   ServiceEntity,
		EntityName:   "TestService",
		FunctionName: "TestFunction",
	}

	// These should not panic or error - we're just testing they execute
	loggerService.Debug("debug message", context)
	loggerService.Info("info message", context)
	loggerService.Warn("warn message", context)
	loggerService.Error("error message", context)
}

func TestLoggerService_LoggingMethodsWithNilContext(t *testing.T) {
	// Reset the singleton
	instance = nil
	once = sync.Once{}

	// Set log level to DEBUG to enable all logging
	_ = os.Setenv("LOG_LEVEL", "DEBUG")
	defer func() { _ = os.Unsetenv("LOG_LEVEL") }()

	loggerService := GetInstance()

	// These should not panic when context is nil
	loggerService.Debug("debug message", nil)
	loggerService.Info("info message", nil)
	loggerService.Warn("warn message", nil)
	loggerService.Error("error message", nil)
}

func TestLoggerService_LoggingMethodsDisabled(t *testing.T) {
	// Reset the singleton
	instance = nil
	once = sync.Once{}

	// Set log level to NONE to disable all logging
	_ = os.Setenv("LOG_LEVEL", "NONE")
	defer func() { _ = os.Unsetenv("LOG_LEVEL") }()

	loggerService := GetInstance()

	context := &LogContext{
		EntityType:   ServiceEntity,
		EntityName:   "TestService",
		FunctionName: "TestFunction",
	}

	// These should not panic even when logging is disabled
	loggerService.Debug("debug message", context)
	loggerService.Info("info message", context)
	loggerService.Warn("warn message", context)
	loggerService.Error("error message", context)
}

func TestLoggerService_HierarchicalLevels(t *testing.T) {
	tests := []struct {
		name     string
		logLevel string
		expected map[LogLevel]bool
	}{
		{
			name:     "EmptyLogLevel_DefaultsToInfo",
			logLevel: "",
			expected: map[LogLevel]bool{
				DebugLevel: false, // DEBUG not included at INFO level
				InfoLevel:  true,  // INFO included
				WarnLevel:  true,  // WARN included
				ErrorLevel: true,  // ERROR included
			},
		},
		{
			name:     "DebugLevel_IncludesAll",
			logLevel: "DEBUG",
			expected: map[LogLevel]bool{
				DebugLevel: true, // DEBUG included
				InfoLevel:  true, // INFO included
				WarnLevel:  true, // WARN included
				ErrorLevel: true, // ERROR included
			},
		},
		{
			name:     "InfoLevel_ExcludesDebug",
			logLevel: "INFO",
			expected: map[LogLevel]bool{
				DebugLevel: false, // DEBUG not included
				InfoLevel:  true,  // INFO included
				WarnLevel:  true,  // WARN included
				ErrorLevel: true,  // ERROR included
			},
		},
		{
			name:     "WarnLevel_OnlyWarnAndError",
			logLevel: "WARN",
			expected: map[LogLevel]bool{
				DebugLevel: false, // DEBUG not included
				InfoLevel:  false, // INFO not included
				WarnLevel:  true,  // WARN included
				ErrorLevel: true,  // ERROR included
			},
		},
		{
			name:     "ErrorLevel_OnlyError",
			logLevel: "ERROR",
			expected: map[LogLevel]bool{
				DebugLevel: false, // DEBUG not included
				InfoLevel:  false, // INFO not included
				WarnLevel:  false, // WARN not included
				ErrorLevel: true,  // ERROR included
			},
		},
		{
			name:     "NoneLevel_DisablesAll",
			logLevel: "NONE",
			expected: map[LogLevel]bool{
				DebugLevel: false, // All disabled
				InfoLevel:  false, // All disabled
				WarnLevel:  false, // All disabled
				ErrorLevel: false, // All disabled
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset the singleton for each test
			instance = nil
			once = sync.Once{}

			// Set the environment variable
			_ = os.Setenv("LOG_LEVEL", tt.logLevel)
			defer func() { _ = os.Unsetenv("LOG_LEVEL") }()

			// Get the logger service instance
			loggerService := GetInstance()

			// Test each log level
			for level, expectedEnabled := range tt.expected {
				actual := loggerService.IsEnabled(level)
				if actual != expectedEnabled {
					t.Errorf("LOG_LEVEL=%s: expected IsEnabled(%s) = %v, got %v",
						tt.logLevel, level.String(), expectedEnabled, actual)
				}
			}
		})
	}
}
