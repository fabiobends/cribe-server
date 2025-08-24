package auth

import (
	"testing"
	"time"

	"cribeapp.com/cribe-server/internal/routes/users"
	"cribeapp.com/cribe-server/internal/utils"
)

var service = NewMockAuthServiceReady()

func TestAuthService_Register(t *testing.T) {
	t.Run("should register a user", func(t *testing.T) {
		user := users.UserDTO{
			FirstName: "John",
			LastName:  "Doe",
			Email:     "john.doe.auth.service@example.com",
			Password:  "password123",
		}

		result, err := service.Register(user)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if result.ID == 0 {
			t.Errorf("Expected user ID, got 0")
		}
	})

	t.Run("should not register a user with empty fields", func(t *testing.T) {
		user := users.UserDTO{
			FirstName: "John",
			LastName:  "Doe",
		}

		_, err := service.Register(user)
		if err == nil {
			t.Errorf("Expected error, got nil")
		}
	})

	t.Run("should not register a user if email is already taken", func(t *testing.T) {
		user := users.UserDTO{
			FirstName: "John",
			LastName:  "Doe",
			Email:     "john.doe.auth.service@example.com",
			Password:  "password123",
		}

		_, err := service.Register(user)
		if err == nil {
			t.Errorf("Expected error, got nil")
		}
	})
}

func TestAuthService_Login(t *testing.T) {
	t.Run("should login a user", func(t *testing.T) {
		result, err := service.Login(LoginRequest{
			Email:    "john.doe.auth.service@example.com",
			Password: "password123",
		})
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if result.AccessToken == "" {
			t.Errorf("Expected access token, got empty string")
		}

		if result.RefreshToken == "" {
			t.Errorf("Expected refresh token, got empty string")
		}
	})

	t.Run("should not login a user with invalid email", func(t *testing.T) {
		service := NewMockAuthServiceReady()
		user := users.UserDTO{
			Email:    "invalid.email@example.com",
			Password: "password123",
		}

		_, err := service.Login(LoginRequest{
			Email:    user.Email,
			Password: user.Password,
		})

		if err.Message != utils.InvalidCredentials {
			t.Errorf("Expected %s, got %s", utils.InvalidCredentials, err.Message)
		}
	})

	t.Run("should not login a user with invalid password", func(t *testing.T) {
		service := NewMockAuthServiceReady()
		user := users.UserDTO{
			Email:    "john.doe.auth.service@example.com",
			Password: "invalid.password",
		}

		_, err := service.Login(LoginRequest{
			Email:    user.Email,
			Password: user.Password,
		})

		if err.Message != utils.InvalidCredentials {
			t.Errorf("Expected %s, got %s", utils.InvalidCredentials, err.Message)
		}
	})

}

func TestAuthService_RefreshToken(t *testing.T) {
	t.Run("should refresh a token", func(t *testing.T) {
		result, err := service.RefreshToken(RefreshTokenRequest{
			RefreshToken: "refresh_token_test",
		})
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if result.AccessToken == "" {
			t.Errorf("Expected access token, got empty string")
		}
	})

	t.Run("should not refresh a token with invalid refresh token", func(t *testing.T) {
		service := NewMockAuthServiceReady()

		_, err := service.RefreshToken(RefreshTokenRequest{
			RefreshToken: "invalid.refresh.token",
		})

		if err.Message != utils.InvalidRequestBody {
			t.Errorf("Expected %s, got %s", utils.InvalidRequestBody, err.Message)
		}
	})

	t.Run("should not refresh a token if it is expired", func(t *testing.T) {
		refreshTokenExpiration := -10 * time.Second // 10 seconds ago
		tokenService := NewMockTokenService([]byte("test"), time.Second*10, refreshTokenExpiration, time.Now)
		userRepo := users.NewMockUserRepositoryReady(users.UserWithPassword{
			Email:    "john.doe.auth.service@example.com",
			Password: "password123",
		})
		service := NewAuthService(userRepo, tokenService)

		_, err := service.RefreshToken(RefreshTokenRequest{
			RefreshToken: "refresh_token_test",
		})

		expectedMessage := "refresh token expired"
		if err.Details != expectedMessage {
			t.Errorf("Expected %s, got %s", expectedMessage, err.Details)
		}
	})
}
