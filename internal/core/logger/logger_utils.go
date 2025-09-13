package logger

import (
	"runtime"
	"strings"
)

// ContextualLogger provides convenient logging methods with automatic context
type ContextualLogger struct {
	entityType EntityType
	entityName string
	service    *LoggerService
}

// NewContextualLogger creates a new contextual logger for a specific entity
func NewContextualLogger(entityType EntityType, entityName string) *ContextualLogger {
	return &ContextualLogger{
		entityType: entityType,
		entityName: entityName,
		service:    GetInstance(),
	}
}

// Debug logs a debug message with automatic context
func (cl *ContextualLogger) Debug(message string, extra ...map[string]interface{}) {
	context := cl.buildContext(extra...)
	cl.service.Debug(message, context)
}

// Info logs an info message with automatic context
func (cl *ContextualLogger) Info(message string, extra ...map[string]interface{}) {
	context := cl.buildContext(extra...)
	cl.service.Info(message, context)
}

// Warn logs a warning message with automatic context
func (cl *ContextualLogger) Warn(message string, extra ...map[string]interface{}) {
	context := cl.buildContext(extra...)
	cl.service.Warn(message, context)
}

// Error logs an error message with automatic context
func (cl *ContextualLogger) Error(message string, extra ...map[string]interface{}) {
	context := cl.buildContext(extra...)
	cl.service.Error(message, context)
}

// buildContext creates a LogContext with automatic function name detection
func (cl *ContextualLogger) buildContext(extra ...map[string]interface{}) *LogContext {
	context := &LogContext{
		EntityType:   cl.entityType,
		EntityName:   cl.entityName,
		FunctionName: cl.getFunctionName(),
		Extra:        make(map[string]interface{}),
	}

	// Merge extra parameters
	for _, extraMap := range extra {
		for k, v := range extraMap {
			context.Extra[k] = v
		}
	}

	return context
}

// getFunctionName automatically detects the calling function name
func (cl *ContextualLogger) getFunctionName() string {
	pc, _, _, ok := runtime.Caller(3) // Skip 3 frames: getFunctionName -> buildContext -> Info/Debug/etc -> actual caller
	if !ok {
		return "unknown"
	}

	funcName := runtime.FuncForPC(pc).Name()

	// Extract just the function name from the full path
	parts := strings.Split(funcName, ".")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}

	return "unknown"
}

// Entity-specific logger constructors

// NewHandlerLogger creates a logger for HTTP handlers
func NewHandlerLogger(handlerName string) *ContextualLogger {
	return NewContextualLogger(HandlerEntity, handlerName)
}

// NewServiceLogger creates a logger for business services
func NewServiceLogger(serviceName string) *ContextualLogger {
	return NewContextualLogger(ServiceEntity, serviceName)
}

// NewRepositoryLogger creates a logger for data repositories
func NewRepositoryLogger(repositoryName string) *ContextualLogger {
	return NewContextualLogger(RepositoryEntity, repositoryName)
}

// NewModelLogger creates a logger for data models
func NewModelLogger(modelName string) *ContextualLogger {
	return NewContextualLogger(ModelEntity, modelName)
}

// NewMiddlewareLogger creates a logger for HTTP middlewares
func NewMiddlewareLogger(middlewareName string) *ContextualLogger {
	return NewContextualLogger(MiddlewareEntity, middlewareName)
}

// NewUtilLogger creates a logger for utility functions
func NewUtilLogger(utilName string) *ContextualLogger {
	return NewContextualLogger(UtilEntity, utilName)
}

// NewCoreLogger creates a logger for core system components
func NewCoreLogger(componentName string) *ContextualLogger {
	return NewContextualLogger(CoreEntity, componentName)
}

// Convenience functions for quick logging without creating a logger instance

// LogDebug logs a debug message with minimal context
func LogDebug(entityType EntityType, entityName, message string) {
	context := &LogContext{
		EntityType: entityType,
		EntityName: entityName,
	}
	GetInstance().Debug(message, context)
}

// LogInfo logs an info message with minimal context
func LogInfo(entityType EntityType, entityName, message string) {
	context := &LogContext{
		EntityType: entityType,
		EntityName: entityName,
	}
	GetInstance().Info(message, context)
}

// LogWarn logs a warning message with minimal context
func LogWarn(entityType EntityType, entityName, message string) {
	context := &LogContext{
		EntityType: entityType,
		EntityName: entityName,
	}
	GetInstance().Warn(message, context)
}

// LogError logs an error message with minimal context
func LogError(entityType EntityType, entityName, message string) {
	context := &LogContext{
		EntityType: entityType,
		EntityName: entityName,
	}
	GetInstance().Error(message, context)
}
