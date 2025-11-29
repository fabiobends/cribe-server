package logger

import (
	"testing"
)

func TestLogLevel_String(t *testing.T) {
	tests := []struct {
		level    LogLevel
		expected string
	}{
		{DebugLevel, "DEBUG"},
		{InfoLevel, "INFO"},
		{WarnLevel, "WARN"},
		{ErrorLevel, "ERROR"},
		{LogLevel(99), "UNKNOWN"}, // Test unknown level
	}

	for _, test := range tests {
		t.Run(test.expected, func(t *testing.T) {
			result := test.level.String()
			if result != test.expected {
				t.Errorf("Expected %s, got %s", test.expected, result)
			}
		})
	}
}

func TestEntityType_Constants(t *testing.T) {
	tests := []struct {
		entityType EntityType
		expected   string
	}{
		{HandlerEntity, "HANDLER"},
		{ServiceEntity, "SERVICE"},
		{RepositoryEntity, "REPOSITORY"},
		{ModelEntity, "MODEL"},
		{MiddlewareEntity, "MIDDLEWARE"},
		{UtilEntity, "UTIL"},
		{CoreEntity, "CORE"},
	}

	for _, test := range tests {
		t.Run(string(test.entityType), func(t *testing.T) {
			if string(test.entityType) != test.expected {
				t.Errorf("Expected %s, got %s", test.expected, string(test.entityType))
			}
		})
	}
}

func TestLogContext_Creation(t *testing.T) {
	extra := map[string]any{
		"userId": 123,
		"action": "test",
	}

	context := &LogContext{
		EntityType:   ServiceEntity,
		EntityName:   "TestService",
		FunctionName: "TestFunction",
		Extra:        extra,
	}

	if context.EntityType != ServiceEntity {
		t.Errorf("Expected EntityType %s, got %s", ServiceEntity, context.EntityType)
	}

	if context.EntityName != "TestService" {
		t.Errorf("Expected EntityName 'TestService', got %s", context.EntityName)
	}

	if context.FunctionName != "TestFunction" {
		t.Errorf("Expected FunctionName 'TestFunction', got %s", context.FunctionName)
	}

	if context.Extra["userId"] != 123 {
		t.Errorf("Expected Extra userId 123, got %v", context.Extra["userId"])
	}
}

// Test that the Logger interface can be implemented
type MockLogger struct {
	lastMessage string
	lastContext *LogContext
}

func (m *MockLogger) Debug(message string, context *LogContext) {
	m.lastMessage = message
	m.lastContext = context
}

func (m *MockLogger) Info(message string, context *LogContext) {
	m.lastMessage = message
	m.lastContext = context
}

func (m *MockLogger) Warn(message string, context *LogContext) {
	m.lastMessage = message
	m.lastContext = context
}

func (m *MockLogger) Error(message string, context *LogContext) {
	m.lastMessage = message
	m.lastContext = context
}

func TestLogger_Interface(t *testing.T) {
	mock := &MockLogger{}
	var logger Logger = mock // This tests that MockLogger implements Logger interface

	context := &LogContext{
		EntityType: HandlerEntity,
		EntityName: "TestHandler",
	}

	// Test each method
	logger.Debug("debug message", context)
	if mock.lastMessage != "debug message" {
		t.Errorf("Expected 'debug message', got %s", mock.lastMessage)
	}

	logger.Info("info message", context)
	if mock.lastMessage != "info message" {
		t.Errorf("Expected 'info message', got %s", mock.lastMessage)
	}

	logger.Warn("warn message", context)
	if mock.lastMessage != "warn message" {
		t.Errorf("Expected 'warn message', got %s", mock.lastMessage)
	}

	logger.Error("error message", context)
	if mock.lastMessage != "error message" {
		t.Errorf("Expected 'error message', got %s", mock.lastMessage)
	}
}
