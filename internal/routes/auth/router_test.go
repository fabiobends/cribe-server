package auth

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"cribeapp.com/cribe-server/internal/routes/migrations"
	"cribeapp.com/cribe-server/internal/routes/users"
	"cribeapp.com/cribe-server/internal/utils"
)

func TestAuthRouter_IntegrationTests(t *testing.T) {
	log.Printf("Setting up test environment")
	utils.CleanDatabaseAndRunMigrations(migrations.HandleHTTPRequests)
	var refreshToken string

	t.Run("should register a user", func(t *testing.T) {
		user := users.UserDTO{
			FirstName: "John",
			LastName:  "Doe",
			Email:     "john.doe.auth.router@example.com",
			Password:  "password123",
		}

		body, _ := json.Marshal(user)
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBuffer(body))
		handler.HandleRequest(w, r)

		if w.Code != http.StatusCreated {
			t.Errorf("Expected status code %v, got %v", http.StatusCreated, w.Code)
		}

		response := utils.DecodeResponse[users.UserWithPassword](w.Body.String())
		if response.ID == 0 {
			t.Errorf("Expected user ID to be non-zero, got %v", response.ID)
		}
	})

	t.Run("should login a user", func(t *testing.T) {
		loginRequest := LoginRequest{
			Email:    "john.doe.auth.router@example.com",
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
			t.Errorf("Expected access token to be non-empty, got %v", response.AccessToken)
		}

		if response.RefreshToken == "" {
			t.Errorf("Expected refresh token to be non-empty, got %v", response.RefreshToken)
		}

		refreshToken = response.RefreshToken
	})

	t.Run("should refresh a token", func(t *testing.T) {
		refreshRequest := RefreshTokenRequest{
			RefreshToken: refreshToken,
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
			t.Errorf("Expected access token to be non-empty, got %v", response.AccessToken)
		}
	})

	t.Run("should not refresh a token with invalid refresh token", func(t *testing.T) {
		refreshRequest := RefreshTokenRequest{
			RefreshToken: "invalid_refresh_token",
		}

		body, _ := json.Marshal(refreshRequest)
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/auth/refresh", bytes.NewBuffer(body))
		handler.HandleRequest(w, r)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected status code %v, got %v", http.StatusUnauthorized, w.Code)
		}

		response := utils.DecodeResponse[utils.ErrorResponse](w.Body.String())
		if response.Message != utils.InvalidRequestBody {
			t.Errorf("Expected message to be %s, got %v", utils.InvalidRequestBody, response.Message)
		}
	})

}
