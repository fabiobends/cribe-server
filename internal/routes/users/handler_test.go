package users

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"cribeapp.com/cribe-server/internal/errors"
	"cribeapp.com/cribe-server/internal/utils"
)

var handler = NewMockUserHandlerReady()

func TestUserHandler_HandleRequest(t *testing.T) {
	t.Run("should create a user with valid input and return the user", func(t *testing.T) {
		userDTO := UserDTO{
			FirstName: "John",
			LastName:  "Doe",
			Email:     "john.doe.user.handler@example.com",
			Password:  "password123",
		}

		body, _ := json.Marshal(userDTO)
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(body))

		handler.HandleRequest(w, r)

		if w.Code != http.StatusCreated {
			t.Errorf("Expected status code %v, got %v", http.StatusCreated, w.Code)
		}

		result := utils.DecodeResponse[User](w.Body.String())

		if result.FirstName != userDTO.FirstName {
			t.Errorf("Expected first name %v, got %v", userDTO.FirstName, result.FirstName)
		}
	})

	t.Run("should return validation errors for invalid input", func(t *testing.T) {
		userDTO := UserDTO{
			FirstName: "John",
			// Missing LastName
			Email:    "invalid-email",
			Password: "short",
		}

		body, _ := json.Marshal(userDTO)
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(body))

		handler.HandleRequest(w, r)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status code %v, got %v", http.StatusBadRequest, w.Code)
		}

		result := utils.DecodeResponse[errors.ErrorResponse](w.Body.String())

		if len(result.Details) == 0 {
			t.Error("Expected validation errors, got none")
		}
	})

	t.Run("should get a user by id", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/users/1", nil)

		handler.HandleRequest(w, r)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status code %v, got %v", http.StatusOK, w.Code)
		}

		result := utils.DecodeResponse[User](w.Body.String())

		if result.ID != 1 {
			t.Errorf("Expected user id %v, got %v", 1, result.ID)
		}
	})

	t.Run("should get all users", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/users", nil)

		handler.HandleRequest(w, r)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status code %v, got %v", http.StatusOK, w.Code)
		}

		result := utils.DecodeResponse[[]User](w.Body.String())

		if len(result) < 1 {
			t.Errorf("Expected at least 1 user, got %v", len(result))
		}
	})

	t.Run("should return 405 for unsupported method", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPut, "/users", nil)

		handler.HandleRequest(w, r)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("Expected status code %v, got %v", http.StatusMethodNotAllowed, w.Code)
		}
	})
}
