package utils

import "net/http"

// NewErrorResponse creates a new ErrorResponse with the given message and details
func NewErrorResponse(statusCode int, message string, details ...string) *ErrorResponse {
	return &ErrorResponse{
		StatusCode: statusCode,
		Message:    message,
		Details:    details,
	}
}

// NewValidationError creates a new ErrorResponse for validation errors
func NewValidationError(missingFields ...string) *ErrorResponse {
	return NewErrorResponse(http.StatusBadRequest, "Missing required fields", missingFields...)
}

// NewDatabaseError creates a new ErrorResponse for database errors
func NewDatabaseError(err error) *ErrorResponse {
	return NewErrorResponse(http.StatusInternalServerError, "Database error", err.Error())
}

// NewInternalError creates a new ErrorResponse for internal server errors
func NewInternalError(err error) *ErrorResponse {
	return NewErrorResponse(http.StatusInternalServerError, "Internal server error", err.Error())
}
