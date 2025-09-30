package middlewares

import (
	"net/http/httptest"
	"testing"
)

func TestPublicMiddleware(t *testing.T) {
	t.Run("should return true for public routes", func(t *testing.T) {
		request := httptest.NewRequest("GET", "/status", nil)
		response := httptest.NewRecorder()
		result := PrivateCheckerMiddleware(response, request)

		expected := false
		if result != expected {
			t.Errorf("Expected %v, got %v", expected, result)
		}
	})

	t.Run("should return false for non-public routes", func(t *testing.T) {
		request := httptest.NewRequest("GET", "/users", nil)
		response := httptest.NewRecorder()
		result := PrivateCheckerMiddleware(response, request)

		expected := true
		if result != expected {
			t.Errorf("Expected %v, got %v", expected, result)
		}
	})

	t.Run("should return true for unknown routes", func(t *testing.T) {
		request := httptest.NewRequest("GET", "/unknown", nil)
		response := httptest.NewRecorder()
		result := PrivateCheckerMiddleware(response, request)

		expected := false
		if result != expected {
			t.Errorf("Expected %v, got %v", expected, result)
		}
	})

	t.Run("should treat /migrations a public route with appropriate header is set", func(t *testing.T) {
		request := httptest.NewRequest("GET", "/migrations", nil)
		request.Header.Set("x-migration-run", "true")
		response := httptest.NewRecorder()
		result := PrivateCheckerMiddleware(response, request)

		expected := false
		if result != expected {
			t.Errorf("Expected %v, got %v", expected, result)
		}
	})
}
