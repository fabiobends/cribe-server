package logger

import (
	"os"
	"sync"
	"testing"
)

func TestNewContextualLogger(t *testing.T) {
	logger := NewContextualLogger(ServiceEntity, "TestService")

	if logger == nil {
		t.Error("NewContextualLogger should not return nil")
		return
	}

	if logger.entityType != ServiceEntity {
		t.Errorf("Expected entityType %v, got %v", ServiceEntity, logger.entityType)
	}

	if logger.entityName != "TestService" {
		t.Errorf("Expected entityName %s, got %s", "TestService", logger.entityName)
	}

	if logger.service == nil {
		t.Error("ContextualLogger service should not be nil")
	}
}

func TestContextualLogger_LoggingMethods(t *testing.T) {
	// Reset the singleton and enable logging
	instance = nil
	once = sync.Once{}
	_ = os.Setenv("LOG_LEVEL", "DEBUG")
	defer func() { _ = os.Unsetenv("LOG_LEVEL") }()

	logger := NewContextualLogger(ServiceEntity, "TestService")

	// These should not panic or error - we're just testing they execute
	logger.Debug("debug message")
	logger.Info("info message")
	logger.Warn("warn message")
	logger.Error("error message")
}

func TestContextualLogger_LoggingMethodsWithExtra(t *testing.T) {
	// Reset the singleton and enable logging
	instance = nil
	once = sync.Once{}
	_ = os.Setenv("LOG_LEVEL", "DEBUG")
	defer func() { _ = os.Unsetenv("LOG_LEVEL") }()

	logger := NewContextualLogger(ServiceEntity, "TestService")

	extra := map[string]any{
		"userId": 123,
		"action": "test",
	}

	// These should not panic or error with extra parameters
	logger.Debug("debug message", extra)
	logger.Info("info message", extra)
	logger.Warn("warn message", extra)
	logger.Error("error message", extra)
}

func TestContextualLogger_BuildContext(t *testing.T) {
	logger := NewContextualLogger(ServiceEntity, "TestService")

	// Test with no extra parameters
	context := logger.buildContext()

	if context.EntityType != ServiceEntity {
		t.Errorf("Expected EntityType %v, got %v", ServiceEntity, context.EntityType)
	}

	if context.EntityName != "TestService" {
		t.Errorf("Expected EntityName %s, got %s", "TestService", context.EntityName)
	}

	if context.Extra == nil {
		t.Error("Context.Extra should not be nil")
	}

	// Test with extra parameters
	extra1 := map[string]any{"key1": "value1"}
	extra2 := map[string]any{"key2": "value2"}

	contextWithExtra := logger.buildContext(extra1, extra2)

	if contextWithExtra.Extra["key1"] != "value1" {
		t.Errorf("Expected key1=value1, got %v", contextWithExtra.Extra["key1"])
	}

	if contextWithExtra.Extra["key2"] != "value2" {
		t.Errorf("Expected key2=value2, got %v", contextWithExtra.Extra["key2"])
	}
}

func TestContextualLogger_GetFunctionName(t *testing.T) {
	logger := NewContextualLogger(ServiceEntity, "TestService")

	functionName := logger.getFunctionName()

	// The function name should not be empty and should be a string
	if functionName == "" {
		t.Error("Function name should not be empty")
	}

	// It should return a valid function name (could be "unknown" if detection fails)
	if functionName != "unknown" && len(functionName) == 0 {
		t.Error("Function name should be either 'unknown' or a valid name")
	}
}

// Test entity-specific logger constructors

func TestNewHandlerLogger(t *testing.T) {
	logger := NewHandlerLogger("TestHandler")

	if logger.entityType != HandlerEntity {
		t.Errorf("Expected HandlerEntity, got %v", logger.entityType)
	}

	if logger.entityName != "TestHandler" {
		t.Errorf("Expected TestHandler, got %s", logger.entityName)
	}
}

func TestNewServiceLogger(t *testing.T) {
	logger := NewServiceLogger("TestService")

	if logger.entityType != ServiceEntity {
		t.Errorf("Expected ServiceEntity, got %v", logger.entityType)
	}

	if logger.entityName != "TestService" {
		t.Errorf("Expected TestService, got %s", logger.entityName)
	}
}

func TestNewRepositoryLogger(t *testing.T) {
	logger := NewRepositoryLogger("TestRepository")

	if logger.entityType != RepositoryEntity {
		t.Errorf("Expected RepositoryEntity, got %v", logger.entityType)
	}

	if logger.entityName != "TestRepository" {
		t.Errorf("Expected TestRepository, got %s", logger.entityName)
	}
}

func TestNewModelLogger(t *testing.T) {
	logger := NewModelLogger("TestModel")

	if logger.entityType != ModelEntity {
		t.Errorf("Expected ModelEntity, got %v", logger.entityType)
	}

	if logger.entityName != "TestModel" {
		t.Errorf("Expected TestModel, got %s", logger.entityName)
	}
}

func TestNewMiddlewareLogger(t *testing.T) {
	logger := NewMiddlewareLogger("TestMiddleware")

	if logger.entityType != MiddlewareEntity {
		t.Errorf("Expected MiddlewareEntity, got %v", logger.entityType)
	}

	if logger.entityName != "TestMiddleware" {
		t.Errorf("Expected TestMiddleware, got %s", logger.entityName)
	}
}

func TestNewUtilLogger(t *testing.T) {
	logger := NewUtilLogger("TestUtil")

	if logger.entityType != UtilEntity {
		t.Errorf("Expected UtilEntity, got %v", logger.entityType)
	}

	if logger.entityName != "TestUtil" {
		t.Errorf("Expected TestUtil, got %s", logger.entityName)
	}
}

func TestNewCoreLogger(t *testing.T) {
	logger := NewCoreLogger("TestCore")

	if logger.entityType != CoreEntity {
		t.Errorf("Expected CoreEntity, got %v", logger.entityType)
	}

	if logger.entityName != "TestCore" {
		t.Errorf("Expected TestCore, got %s", logger.entityName)
	}
}

// Test convenience functions

func TestLogDebug(t *testing.T) {
	// Reset the singleton and enable logging
	instance = nil
	once = sync.Once{}
	_ = os.Setenv("LOG_LEVEL", "DEBUG")
	defer func() { _ = os.Unsetenv("LOG_LEVEL") }()

	// Should not panic
	LogDebug(ServiceEntity, "TestService", "debug message")
}

func TestLogInfo(t *testing.T) {
	// Reset the singleton and enable logging
	instance = nil
	once = sync.Once{}
	_ = os.Setenv("LOG_LEVEL", "DEBUG")
	defer func() { _ = os.Unsetenv("LOG_LEVEL") }()

	// Should not panic
	LogInfo(ServiceEntity, "TestService", "info message")
}

func TestLogWarn(t *testing.T) {
	// Reset the singleton and enable logging
	instance = nil
	once = sync.Once{}
	_ = os.Setenv("LOG_LEVEL", "DEBUG")
	defer func() { _ = os.Unsetenv("LOG_LEVEL") }()

	// Should not panic
	LogWarn(ServiceEntity, "TestService", "warn message")
}

func TestLogError(t *testing.T) {
	// Reset the singleton and enable logging
	instance = nil
	once = sync.Once{}
	_ = os.Setenv("LOG_LEVEL", "DEBUG")
	defer func() { _ = os.Unsetenv("LOG_LEVEL") }()

	// Should not panic
	LogError(ServiceEntity, "TestService", "error message")
}

func TestConvenienceFunctions_Disabled(t *testing.T) {
	// Reset the singleton and disable logging
	instance = nil
	once = sync.Once{}
	_ = os.Setenv("LOG_LEVEL", "NONE")
	defer func() { _ = os.Unsetenv("LOG_LEVEL") }()

	// Should not panic even when logging is disabled
	LogDebug(ServiceEntity, "TestService", "debug message")
	LogInfo(ServiceEntity, "TestService", "info message")
	LogWarn(ServiceEntity, "TestService", "warn message")
	LogError(ServiceEntity, "TestService", "error message")
}
