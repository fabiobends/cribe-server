package middlewares

import (
	"net/http/httptest"
	"testing"
	"time"

	"cribeapp.com/cribe-server/internal/routes/auth"
	"cribeapp.com/cribe-server/internal/utils"
)

func TestAuthMiddleware(t *testing.T) {
	t.Run("should allow access in private routes when token is valid", func(t *testing.T) {
		tokenService := auth.NewMockTokenService([]byte("test"), time.Hour, time.Hour*24*30, utils.MockGetCurrentTime)
		accessToken, _ := tokenService.GetAccessToken(1)

		request := httptest.NewRequest("GET", "/users", nil)
		request.Header.Set("Authorization", "Bearer "+accessToken)
		response := httptest.NewRecorder()

		token, _ := AuthMiddleware(response, request, tokenService)
		if token == nil {
			t.Errorf("Expected a token object, got nil")
		}
	})

	t.Run("should not allow access in private routes when token is invalid", func(t *testing.T) {
		tokenService := auth.NewMockTokenService([]byte("test"), time.Hour, time.Hour*24*30, utils.MockGetCurrentTime)
		accessToken, _ := tokenService.GetAccessToken(1)

		request := httptest.NewRequest("GET", "/users", nil)
		request.Header.Set("Authorization", "Bearer "+accessToken+"invalid")
		response := httptest.NewRecorder()

		token, _ := AuthMiddleware(response, request, tokenService)
		if token != nil {
			t.Errorf("Expected nil, got a token object")
		}
	})

	t.Run("should not allow access in private routes when token is expired", func(t *testing.T) {
		accessTokenExpiration := -10 * time.Second // 10 seconds ago
		tokenService := auth.NewMockTokenService([]byte("test"), accessTokenExpiration, time.Hour*24*30, time.Now)
		accessToken, _ := tokenService.GetAccessToken(1)

		request := httptest.NewRequest("GET", "/users", nil)
		request.Header.Set("Authorization", "Bearer "+accessToken)
		response := httptest.NewRecorder()

		token, _ := AuthMiddleware(response, request, tokenService)
		if token != nil {
			t.Errorf("Expected nil, got a token object")
		}
	})
}
