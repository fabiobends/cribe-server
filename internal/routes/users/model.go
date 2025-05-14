package users

import (
	"strings"
	"time"

	"cribeapp.com/cribe-server/internal/utils"
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
func (dto UserDTO) Validate() *utils.ErrorResponse {
	var errors []string

	// Required fields validation
	if strings.TrimSpace(dto.FirstName) == "" {
		errors = append(errors, "First name is required")
	}
	if strings.TrimSpace(dto.LastName) == "" {
		errors = append(errors, "Last name is required")
	}
	if strings.TrimSpace(dto.Email) == "" {
		errors = append(errors, "Email is required")
	}
	if strings.TrimSpace(dto.Password) == "" {
		errors = append(errors, "Password is required")
	}

	// Email format validation
	if strings.TrimSpace(dto.Email) != "" && !strings.Contains(dto.Email, "@") {
		errors = append(errors, "Invalid email format")
	}

	// Password length validation
	if strings.TrimSpace(dto.Password) != "" && len(dto.Password) < 8 {
		errors = append(errors, "Password must be at least 8 characters long")
	}

	if len(errors) > 0 {
		return &utils.ErrorResponse{
			Message: utils.ValidationError,
			Details: strings.Join(errors, ", "),
		}
	}

	return nil
}
