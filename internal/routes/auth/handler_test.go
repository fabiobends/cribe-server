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

		response, err := utils.DecodeResponse[users.UserWithPassword](w.Body.String())
		if err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

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

		response, err := utils.DecodeResponse[LoginResponse](w.Body.String())
		if err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

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

		response, err := utils.DecodeResponse[RefreshTokenResponse](w.Body.String())
		if err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if response.AccessToken == "" {
			t.Errorf("Expected access token, got empty string")
		}
	})

	t.Run("should return 404 for unknown route", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/auth/unknown", nil)
		handler.HandleRequest(w, r)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status code %v, got %v", http.StatusNotFound, w.Code)
		}
	})

	t.Run("should return 405 for invalid method", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodDelete, "/auth/login", nil)
		handler.HandleRequest(w, r)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("Expected status code %v, got %v", http.StatusMethodNotAllowed, w.Code)
		}
	})

	t.Run("should return 400 for invalid JSON in login request", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer([]byte("invalid json")))
		handler.HandleRequest(w, r)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status code %v, got %v", http.StatusBadRequest, w.Code)
		}

		response, err := utils.DecodeResponse[map[string]interface{}](w.Body.String())
		if err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if response["message"] != "Invalid request body" {
			t.Errorf("Expected message 'Invalid request body', got %v", response["message"])
		}
	})

	t.Run("should return 400 for missing email in login request", func(t *testing.T) {
		loginRequest := LoginRequest{
			Password: "password123",
		}
		body, _ := json.Marshal(loginRequest)
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(body))
		handler.HandleRequest(w, r)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status code %v, got %v", http.StatusBadRequest, w.Code)
		}

		response, err := utils.DecodeResponse[map[string]interface{}](w.Body.String())
		if err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if response["message"] != "Validation error" {
			t.Errorf("Expected message 'Validation error', got %v", response["message"])
		}
	})

	t.Run("should return 400 for invalid email format in login request", func(t *testing.T) {
		loginRequest := LoginRequest{
			Email:    "invalid-email",
			Password: "password123",
		}
		body, _ := json.Marshal(loginRequest)
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(body))
		handler.HandleRequest(w, r)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status code %v, got %v", http.StatusBadRequest, w.Code)
		}

		response, err := utils.DecodeResponse[map[string]interface{}](w.Body.String())
		if err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if response["message"] != "Validation error" {
			t.Errorf("Expected message 'Validation error', got %v", response["message"])
		}
	})

	t.Run("should return 400 for missing password in login request", func(t *testing.T) {
		loginRequest := LoginRequest{
			Email: "test@example.com",
		}
		body, _ := json.Marshal(loginRequest)
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(body))
		handler.HandleRequest(w, r)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status code %v, got %v", http.StatusBadRequest, w.Code)
		}

		response, err := utils.DecodeResponse[map[string]interface{}](w.Body.String())
		if err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if response["message"] != "Validation error" {
			t.Errorf("Expected message 'Validation error', got %v", response["message"])
		}
	})

	t.Run("should return 400 for login with invalid credentials", func(t *testing.T) {
		loginRequest := LoginRequest{
			Email:    "nonexistent@example.com",
			Password: "wrongpassword",
		}
		body, _ := json.Marshal(loginRequest)
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(body))
		handler.HandleRequest(w, r)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status code %v, got %v", http.StatusBadRequest, w.Code)
		}

		response, err := utils.DecodeResponse[map[string]interface{}](w.Body.String())
		if err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if response["message"] != "Invalid credentials" {
			t.Errorf("Expected message 'Invalid credentials', got %v", response["message"])
		}
	})

	// Error case tests for register
	t.Run("should return 400 for invalid JSON in register request", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBuffer([]byte("invalid json")))
		handler.HandleRequest(w, r)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status code %v, got %v", http.StatusBadRequest, w.Code)
		}

		response, err := utils.DecodeResponse[map[string]interface{}](w.Body.String())
		if err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if response["message"] != "Invalid request body" {
			t.Errorf("Expected message 'Invalid request body', got %v", response["message"])
		}
	})

	t.Run("should return 400 for missing fields in register request", func(t *testing.T) {
		userDTO := users.UserDTO{
			FirstName: "John",
			// Missing LastName, Email, Password
		}
		body, _ := json.Marshal(userDTO)
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBuffer(body))
		handler.HandleRequest(w, r)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status code %v, got %v", http.StatusBadRequest, w.Code)
		}

		response, err := utils.DecodeResponse[map[string]interface{}](w.Body.String())
		if err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if response["message"] != "Validation error" {
			t.Errorf("Expected message 'Validation error', got %v", response["message"])
		}
	})

	// Error case tests for refresh token
	t.Run("should return 400 for invalid JSON in refresh token request", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/auth/refresh", bytes.NewBuffer([]byte("invalid json")))
		handler.HandleRequest(w, r)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status code %v, got %v", http.StatusBadRequest, w.Code)
		}

		response, err := utils.DecodeResponse[map[string]interface{}](w.Body.String())
		if err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if response["message"] != "Invalid request body" {
			t.Errorf("Expected message 'Invalid request body', got %v", response["message"])
		}
	})

	t.Run("should return 400 for missing refresh token in request", func(t *testing.T) {
		refreshRequest := RefreshTokenRequest{
			// Missing RefreshToken
		}
		body, _ := json.Marshal(refreshRequest)
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/auth/refresh", bytes.NewBuffer(body))
		handler.HandleRequest(w, r)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status code %v, got %v", http.StatusBadRequest, w.Code)
		}

		response, err := utils.DecodeResponse[map[string]interface{}](w.Body.String())
		if err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if response["message"] != "Validation error" {
			t.Errorf("Expected message 'Validation error', got %v", response["message"])
		}
	})

	// Test GET method returns 405 for all auth endpoints
	t.Run("should return 405 for GET method on login endpoint", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/auth/login", nil)
		handler.HandleRequest(w, r)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("Expected status code %v, got %v", http.StatusMethodNotAllowed, w.Code)
		}
	})

	t.Run("should return 405 for GET method on register endpoint", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/auth/register", nil)
		handler.HandleRequest(w, r)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("Expected status code %v, got %v", http.StatusMethodNotAllowed, w.Code)
		}
	})

	t.Run("should return 405 for GET method on refresh endpoint", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/auth/refresh", nil)
		handler.HandleRequest(w, r)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("Expected status code %v, got %v", http.StatusMethodNotAllowed, w.Code)
		}
	})
}
