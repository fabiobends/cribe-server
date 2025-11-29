package logger

import (
	"strings"
	"testing"
)

func TestNewConsoleLogger(t *testing.T) {
	logger := NewConsoleLogger()

	if logger == nil {
		t.Error("NewConsoleLogger should not return nil")
		return
	}

	if !logger.enableColors {
		t.Error("Default logger should have colors enabled")
	}

	if !logger.enableEmojis {
		t.Error("Default logger should have emojis enabled")
	}
}

func TestNewConsoleLoggerWithOptions(t *testing.T) {
	tests := []struct {
		name         string
		enableColors bool
		enableEmojis bool
	}{
		{"Colors and emojis enabled", true, true},
		{"Colors enabled, emojis disabled", true, false},
		{"Colors disabled, emojis enabled", false, true},
		{"Both disabled", false, false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			logger := NewConsoleLoggerWithOptions(test.enableColors, test.enableEmojis)

			if logger.enableColors != test.enableColors {
				t.Errorf("Expected enableColors %v, got %v", test.enableColors, logger.enableColors)
			}

			if logger.enableEmojis != test.enableEmojis {
				t.Errorf("Expected enableEmojis %v, got %v", test.enableEmojis, logger.enableEmojis)
			}
		})
	}
}

func TestConsoleLogger_LoggingMethods(t *testing.T) {
	logger := NewConsoleLogger()
	context := &LogContext{
		EntityType:   ServiceEntity,
		EntityName:   "TestService",
		FunctionName: "TestFunction",
	}

	// These should not panic or error - we're just testing they execute
	logger.Debug("debug message", context)
	logger.Info("info message", context)
	logger.Warn("warn message", context)
	logger.Error("error message", context)
}

func TestConsoleLogger_LoggingWithNilContext(t *testing.T) {
	logger := NewConsoleLogger()

	// These should not panic when context is nil
	logger.Debug("debug message", nil)
	logger.Info("info message", nil)
	logger.Warn("warn message", nil)
	logger.Error("error message", nil)
}

func TestConsoleLogger_MaskSensitiveData(t *testing.T) {
	logger := NewConsoleLogger()

	tests := []struct {
		name        string
		input       string
		shouldMask  bool
		description string
	}{
		{
			name:        "Email masking",
			input:       "User email is test@example.com",
			shouldMask:  true,
			description: "should mask email addresses",
		},
		{
			name:        "Password masking",
			input:       "password: secret123",
			shouldMask:  true,
			description: "should mask password fields",
		},
		{
			name:        "Normal text",
			input:       "This is just normal text",
			shouldMask:  false,
			description: "should not mask normal text",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := logger.maskSensitiveData(test.input)

			if test.shouldMask {
				if result == test.input {
					t.Errorf("Input should have been masked: %s", test.input)
				}
				if strings.Contains(result, "****") {
					// Good - masking occurred
				} else {
					t.Errorf("Expected masking with ****, got: %s", result)
				}
			} else if result != test.input {
				t.Errorf("Input should not have been changed: expected %s, got %s", test.input, result)
			}
		})
	}
}

func TestConsoleLogger_MaskStringFixedLength(t *testing.T) {
	logger := NewConsoleLogger()

	tests := []struct {
		input     string
		showChars int
		expected  string
	}{
		{"short", 2, "sh****"},        // Long string should show prefix + ****
		{"longstring", 4, "long****"}, // Long string should show prefix + ****
		{"test", 4, "****"},           // String equal to showChars should return ****
		{"ab", 5, "****"},             // Very short string
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			result := logger.maskStringFixedLength(test.input, test.showChars)
			if result != test.expected {
				t.Errorf("Expected %s, got %s", test.expected, result)
			}
		})
	}
}

func TestConsoleLogger_FormatExtra(t *testing.T) {
	logger := NewConsoleLogger()

	tests := []struct {
		name     string
		context  *LogContext
		expected string
	}{
		{
			name:     "Nil context",
			context:  nil,
			expected: "",
		},
		{
			name: "Empty extra",
			context: &LogContext{
				EntityType: ServiceEntity,
				EntityName: "TestService",
				Extra:      map[string]any{},
			},
			expected: "",
		},
		{
			name: "Nil extra",
			context: &LogContext{
				EntityType: ServiceEntity,
				EntityName: "TestService",
				Extra:      nil,
			},
			expected: "",
		},
		{
			name: "With extra data",
			context: &LogContext{
				EntityType: ServiceEntity,
				EntityName: "TestService",
				Extra: map[string]any{
					"userId": 123,
				},
			},
			expected: "[userId=123]",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := logger.formatExtra(test.context)

			if test.expected == "" {
				if result != "" {
					t.Errorf("Expected empty string, got %s", result)
				}
			} else {
				if !strings.Contains(result, "userId") || !strings.Contains(result, "123") {
					t.Errorf("Expected result to contain userId and 123, got %s", result)
				}
			}
		})
	}
}
