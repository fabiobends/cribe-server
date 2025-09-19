package utils

import (
	"strings"
	"sync"

	"cribeapp.com/cribe-server/internal/errors"
	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate
var initOnce sync.Once

// ValidateStruct validates a struct using validator tags and returns formatted errors
func ValidateStruct(s interface{}) *errors.ErrorResponse {
	initOnce.Do(func() {
		validate = validator.New()
	})

	err := validate.Struct(s)
	if err == nil {
		return nil
	}

	var validationErrors []string

	// Extract validation errors and format them nicely
	for _, err := range err.(validator.ValidationErrors) {
		errorMsg := formatValidationError(err)
		validationErrors = append(validationErrors, errorMsg)
	}

	return &errors.ErrorResponse{
		Message: errors.ValidationError,
		Details: strings.Join(validationErrors, ", "),
	}
}

// formatValidationError converts validator errors to human-readable messages
func formatValidationError(err validator.FieldError) string {
	field := strings.ToLower(err.Field())

	switch err.Tag() {
	case "required":
		return field + " is required"
	case "email":
		return field + " must be a valid email address"
	case "min":
		return field + " must be at least " + err.Param() + " characters long"
	case "max":
		return field + " must be at most " + err.Param() + " characters long"
	case "len":
		return field + " must be exactly " + err.Param() + " characters long"
	case "alpha":
		return field + " must contain only alphabetic characters"
	case "alphanum":
		return field + " must contain only alphanumeric characters"
	case "numeric":
		return field + " must be numeric"
	case "url":
		return field + " must be a valid URL"
	case "uuid":
		return field + " must be a valid UUID"
	case "hexadecimal":
		return field + " must be a valid hexadecimal"
	case "base64":
		return field + " must be valid base64"
	case "contains":
		return field + " must contain '" + err.Param() + "'"
	case "containsany":
		return field + " must contain at least one of '" + err.Param() + "'"
	case "excludes":
		return field + " must not contain '" + err.Param() + "'"
	case "startswith":
		return field + " must start with '" + err.Param() + "'"
	case "endswith":
		return field + " must end with '" + err.Param() + "'"
	case "oneof":
		return field + " must be one of: " + err.Param()
	case "regex":
		return field + " format is invalid"
	default:
		return field + " is invalid"
	}
}

// IsValidEmail checks if an email address is valid using the validator package
func IsValidEmail(email string) bool {
	initOnce.Do(func() {
		validate = validator.New()
	})

	err := validate.Var(email, "required,email")
	return err == nil
}

// IsValidPassword checks if a password meets basic requirements
func IsValidPassword(password string) bool {
	initOnce.Do(func() {
		validate = validator.New()
	})

	// Basic password validation: minimum 8 characters
	err := validate.Var(password, "required,min=8")
	return err == nil
}

// IsValidName checks if a name is valid (required, alphabetic characters and spaces)
func IsValidName(name string) bool {
	initOnce.Do(func() {
		validate = validator.New()
	})

	// Allow alphabetic characters, spaces, hyphens, and apostrophes for names
	err := validate.Var(name, "required,min=1,max=100")
	return err == nil && strings.TrimSpace(name) != ""
}

// Legacy validation functions for backward compatibility
// TODO: Consider removing these in favor of struct validation

// RequiredField represents a field that needs to be validated
type RequiredField struct {
	Name  string
	Value string
}

// ValidateRequiredFields checks if any of the required fields are empty and returns their names
func ValidateRequiredFields(fields ...RequiredField) []string {
	var missingFields []string
	for _, field := range fields {
		if field.Value == "" {
			missingFields = append(missingFields, field.Name)
		}
	}
	return missingFields
}

// ValidateRequiredField is a helper function to create a RequiredField
func ValidateRequiredField(name, value string) RequiredField {
	return RequiredField{
		Name:  name,
		Value: value,
	}
}
