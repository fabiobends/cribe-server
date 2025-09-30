package logger

import (
	"testing"
)

func TestGetLevelColor(t *testing.T) {
	tests := []struct {
		level    LogLevel
		expected string
	}{
		{DebugLevel, ColorLightBlue},
		{InfoLevel, ColorWhite},
		{WarnLevel, ColorYellow},
		{ErrorLevel, ColorRed},
		{LogLevel(99), ColorReset}, // Test unknown level
	}

	for _, test := range tests {
		t.Run(test.level.String(), func(t *testing.T) {
			result := GetLevelColor(test.level)
			if result != test.expected {
				t.Errorf("Expected %s, got %s", test.expected, result)
			}
		})
	}
}

func TestGetLevelEmoji(t *testing.T) {
	tests := []struct {
		level    LogLevel
		expected string
	}{
		{DebugLevel, DebugEmoji},
		{InfoLevel, InfoEmoji},
		{WarnLevel, WarnEmoji},
		{ErrorLevel, ErrorEmoji},
		{LogLevel(99), "❓"}, // Test unknown level
	}

	for _, test := range tests {
		t.Run(test.level.String(), func(t *testing.T) {
			result := GetLevelEmoji(test.level)
			if result != test.expected {
				t.Errorf("Expected %s, got %s", test.expected, result)
			}
		})
	}
}

func TestGetEntityColor(t *testing.T) {
	tests := []struct {
		entityType EntityType
		expected   string
	}{
		{HandlerEntity, "\033[1;35m"},       // Magenta
		{ServiceEntity, "\033[1;32m"},       // Green
		{RepositoryEntity, "\033[1;34m"},    // Blue
		{ModelEntity, "\033[1;93m"},         // Bright yellow
		{MiddlewareEntity, "\033[1;36m"},    // Cyan
		{UtilEntity, "\033[1;90m"},          // Bright black
		{CoreEntity, "\033[1;37m"},          // White
		{EntityType("UNKNOWN"), ColorReset}, // Test unknown entity
	}

	for _, test := range tests {
		t.Run(string(test.entityType), func(t *testing.T) {
			result := GetEntityColor(test.entityType)
			if result != test.expected {
				t.Errorf("Expected %s, got %s", test.expected, result)
			}
		})
	}
}

func TestColorConstants(t *testing.T) {
	// Test that color constants are defined and not empty
	constants := map[string]string{
		"ColorReset":     ColorReset,
		"ColorWhite":     ColorWhite,
		"ColorLightBlue": ColorLightBlue,
		"ColorYellow":    ColorYellow,
		"ColorRed":       ColorRed,
		"StyleBold":      StyleBold,
		"StyleDim":       StyleDim,
	}

	for name, value := range constants {
		t.Run(name, func(t *testing.T) {
			if value == "" {
				t.Errorf("Constant %s should not be empty", name)
			}
		})
	}
}

func TestEmojiConstants(t *testing.T) {
	// Test that emoji constants are defined and not empty
	emojis := map[string]string{
		"DebugEmoji": DebugEmoji,
		"InfoEmoji":  InfoEmoji,
		"WarnEmoji":  WarnEmoji,
		"ErrorEmoji": ErrorEmoji,
	}

	for name, value := range emojis {
		t.Run(name, func(t *testing.T) {
			if value == "" {
				t.Errorf("Emoji constant %s should not be empty", name)
			}
		})
	}
}

func TestLogFormatConstant(t *testing.T) {
	if LogFormat == "" {
		t.Error("LogFormat constant should not be empty")
	}

	// Test that LogFormat contains expected placeholders
	expectedPlaceholders := []string{"%s"}
	for _, placeholder := range expectedPlaceholders {
		if !contains(LogFormat, placeholder) {
			t.Errorf("LogFormat should contain placeholder %s", placeholder)
		}
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			containsInMiddle(s, substr))))
}

func containsInMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
