package users

import (
	"log"
	"net/http"
	"testing"

	"cribeapp.com/cribe-server/internal/errors"
	"cribeapp.com/cribe-server/internal/routes/migrations"
	"cribeapp.com/cribe-server/internal/utils"
)

func TestUsers_IntegrationTests(t *testing.T) {
	log.Printf("Setting up test environment")
	utils.CleanDatabaseAndRunMigrations(migrations.HandleHTTPRequests)

	t.Run("shouldn't get a user since the database is empty", func(t *testing.T) {
		resp := utils.MustSendTestRequest[errors.ErrorResponse](utils.TestRequest{
			Method:      http.MethodGet,
			URL:         "/users/1",
			HandlerFunc: HandleHTTPRequests,
		})

		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("Expected status code %d, got %d", http.StatusNotFound, resp.StatusCode)
		}

		if resp.Body.Message != errors.UserNotFound {
			t.Errorf("Expected message %s, got %s", errors.UserNotFound, resp.Body.Message)
		}

	})

	t.Run("shouldn't get a user with invalid id", func(t *testing.T) {
		resp := utils.MustSendTestRequest[errors.ErrorResponse](utils.TestRequest{
			Method:      http.MethodGet,
			URL:         "/users/invalid",
			HandlerFunc: HandleHTTPRequests,
		})

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, resp.StatusCode)
		}

		if resp.Body.Message != errors.InvalidIdParameter {
			t.Errorf("Expected message %s, got %s", errors.InvalidIdParameter, resp.Body.Message)
		}
	})

	t.Run("shouldn't create a user with invalid input", func(t *testing.T) {
		userDTO := UserDTO{
			FirstName: "John",
			LastName:  "Doe",
			Email:     "john.doe@example.com",
		}

		resp := utils.MustSendTestRequest[errors.ErrorResponse](utils.TestRequest{
			Method:      http.MethodPost,
			URL:         "/users",
			Body:        userDTO,
			HandlerFunc: HandleHTTPRequests,
		})

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, resp.StatusCode)
		}

		if resp.Body.Message != errors.ValidationError {
			t.Errorf("Expected message %s, got %s", errors.ValidationError, resp.Body.Message)
		}
	})

	t.Run("should create a user with valid input", func(t *testing.T) {
		userDTO := UserDTO{
			FirstName: "John",
			LastName:  "Doe",
			Email:     "john.doe@example.com",
			Password:  "password123",
		}

		resp := utils.MustSendTestRequest[User](utils.TestRequest{
			Method:      http.MethodPost,
			URL:         "/users/",
			Body:        userDTO,
			HandlerFunc: HandleHTTPRequests,
		})

		if resp.StatusCode != http.StatusCreated {
			t.Errorf("Expected status code %d, got %d", http.StatusCreated, resp.StatusCode)
		}

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

	t.Run("should get a user by id", func(t *testing.T) {
		resp := utils.MustSendTestRequest[User](utils.TestRequest{
			Method:      http.MethodGet,
			URL:         "/users/1", // First user created in the database
			HandlerFunc: HandleHTTPRequests,
		})

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
		}

		if resp.Body.ID != 1 {
			t.Errorf("Expected id %d, got %d", 1, resp.Body.ID)
		}

		if resp.Body.FirstName != "John" {
			t.Errorf("Expected first name %s, got %s", "John", resp.Body.FirstName)
		}

		if resp.Body.LastName != "Doe" {
			t.Errorf("Expected last name %s, got %s", "Doe", resp.Body.LastName)
		}
	})

	t.Run("shouldn't get a user if the user doesn't exist", func(t *testing.T) {
		resp := utils.MustSendTestRequest[errors.ErrorResponse](utils.TestRequest{
			Method:      http.MethodGet,
			URL:         "/users/2",
			HandlerFunc: HandleHTTPRequests,
		})

		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("Expected status code %d, got %d", http.StatusNotFound, resp.StatusCode)
		}
	})

	t.Run("should get all users", func(t *testing.T) {
		resp := utils.MustSendTestRequest[[]User](utils.TestRequest{
			Method:      http.MethodGet,
			URL:         "/users",
			HandlerFunc: HandleHTTPRequests,
		})

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
		}

		if len(resp.Body) < 1 {
			t.Errorf("Expected at least %d users, got %d", 1, len(resp.Body))
		}
	})
}
