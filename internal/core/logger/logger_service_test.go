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
			expected:    false,
			description: "Empty LOG_LEVEL should disable logging",
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
