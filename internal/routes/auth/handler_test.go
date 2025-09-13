package auth

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"cribeapp.com/cribe-server/internal/routes/users"
	"cribeapp.com/cribe-server/internal/utils"
)

var handler = NewMockAuthHandlerReady()

func TestAuthHandler_HandleRequest(t *testing.T) {
	t.Run("should register a user", func(t *testing.T) {
		userDTO := users.UserDTO{
			FirstName: "John",
			LastName:  "Doe",
			Email:     "john.doe.auth.handler@example.com",
			Password:  "password123",
		}

		body, _ := json.Marshal(userDTO)
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBuffer(body))
		handler.HandleRequest(w, r)

		if w.Code != http.StatusCreated {
			t.Errorf("Expected status code %v, got %v", http.StatusCreated, w.Code)
		}

		response := utils.DecodeResponse[users.UserWithPassword](w.Body.String())

		if response.ID != 1 {
			t.Errorf("Expected ID %v, got %v", 1, response.ID)
		}
	})

	t.Run("should login a user", func(t *testing.T) {
		loginRequest := LoginRequest{
			Email:    "john.doe.auth.handler@example.com",
			Password: "password123",
		}
		body, _ := json.Marshal(loginRequest)
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(body))
		handler.HandleRequest(w, r)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status code %v, got %v", http.StatusOK, w.Code)
		}

		response := utils.DecodeResponse[LoginResponse](w.Body.String())

		if response.AccessToken == "" {
			t.Errorf("Expected access token, got empty string")
		}
	})

	t.Run("should refresh a token", func(t *testing.T) {
		refreshRequest := RefreshTokenRequest{
			RefreshToken: "refresh_token_test",
		}
		body, _ := json.Marshal(refreshRequest)
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/auth/refresh", bytes.NewBuffer(body))

		handler.HandleRequest(w, r)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status code %v, got %v", http.StatusOK, w.Code)
		}

		response := utils.DecodeResponse[RefreshTokenResponse](w.Body.String())

		if response.AccessToken == "" {
			t.Errorf("Expected access token, got empty string")
		}
	})
}
