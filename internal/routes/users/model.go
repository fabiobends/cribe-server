package users

import (
	"strings"
	"time"

	"cribeapp.com/cribe-server/internal/errors"
)

type User struct {
	ID        int       `json:"id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type UserWithPassword struct {
	ID        int       `json:"id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Email     string    `json:"email"`
	Password  string    `json:"password"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type UserDTO struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Password  string `json:"password"`
}

// Validate performs domain validation on the user DTO
func (dto UserDTO) Validate() *errors.ErrorResponse {
	var validationErrors []string

	// Required fields validation
	if strings.TrimSpace(dto.FirstName) == "" {
		validationErrors = append(validationErrors, "First name is required")
	}
	if strings.TrimSpace(dto.LastName) == "" {
		validationErrors = append(validationErrors, "Last name is required")
	}
	if strings.TrimSpace(dto.Email) == "" {
		validationErrors = append(validationErrors, "Email is required")
	}
	if strings.TrimSpace(dto.Password) == "" {
		validationErrors = append(validationErrors, "Password is required")
	}

	// Email format validation
	if strings.TrimSpace(dto.Email) != "" && !strings.Contains(dto.Email, "@") {
		validationErrors = append(validationErrors, "Invalid email format")
	}

	// Password length validation
	if strings.TrimSpace(dto.Password) != "" && len(dto.Password) < 8 {
		validationErrors = append(validationErrors, "Password must be at least 8 characters long")
	}

	if len(validationErrors) > 0 {
		return &errors.ErrorResponse{
			Message: errors.ValidationError,
			Details: strings.Join(validationErrors, ", "),
		}
	}

	return nil
}
