package users

import (
	"log"
	"net/http"
	"testing"

	"cribeapp.com/cribe-server/internal/routes/migrations"
	"cribeapp.com/cribe-server/internal/utils"
)

func TestUsers_PostIntegrationTests(t *testing.T) {
	log.Printf("Setting up test environment")
	utils.CleanDatabaseAndRunMigrations(migrations.Handler)

	t.Run("should create a user with valid input", func(t *testing.T) {
		// Create a valid user request
		userDTO := UserDTO{
			FirstName: "John",
			LastName:  "Doe",
			Email:     "john.doe@example.com",
			Password:  "password123",
		}

		// Send request using test utility
		resp := utils.MustSendTestRequest[UserWithoutPassword](utils.TestRequest{
			Method:      http.MethodPost,
			URL:         "/users",
			Body:        userDTO,
			HandlerFunc: Handler,
		})

		// Check status code
		if resp.StatusCode != http.StatusCreated {
			t.Errorf("Expected status code %d, got %d", http.StatusCreated, resp.StatusCode)
		}
		// Verify response fields
		if resp.Body.FirstName != userDTO.FirstName {
			t.Errorf("Expected first name %s, got %s", userDTO.FirstName, resp.Body.FirstName)
		}
		if resp.Body.LastName != userDTO.LastName {
			t.Errorf("Expected last name %s, got %s", userDTO.LastName, resp.Body.LastName)
		}
		if resp.Body.Email != userDTO.Email {
			t.Errorf("Expected email %s, got %s", userDTO.Email, resp.Body.Email)
		}
	})

	t.Run("should return error with invalid input", func(t *testing.T) {
		// Create an invalid user request (missing password)
		userDTO := UserDTO{
			FirstName: "John",
			LastName:  "Doe",
			Email:     "john.doe@example.com",
		}

		// Send request using test utility
		resp := utils.MustSendTestRequest[utils.ErrorResponse](utils.TestRequest{
			Method:      http.MethodPost,
			URL:         "/users",
			Body:        userDTO,
			HandlerFunc: Handler,
		})

		// Check status code
		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, resp.StatusCode)
		}

		// Verify error response
		if resp.Body.Message == "" {
			t.Error("Expected error message, got empty string")
		}
		if len(resp.Body.Details) == 0 {
			t.Error("Expected error details, got empty slice")
		}
	})
}
