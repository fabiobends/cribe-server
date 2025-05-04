package middlewares

import (
	"net/http/httptest"
	"testing"
)

func TestPublicMiddleware(t *testing.T) {
	t.Run("should return true for public routes", func(t *testing.T) {
		request := httptest.NewRequest("GET", "/status", nil)
		response := httptest.NewRecorder()
		result := PublicMiddleware(response, request)

		if result != true {
			t.Errorf("Expected true, got %v", result)
		}
	})

	t.Run("should return false for non-public routes", func(t *testing.T) {
		request := httptest.NewRequest("GET", "/users", nil)
		response := httptest.NewRecorder()
		result := PublicMiddleware(response, request)

		if result != false {
			t.Errorf("Expected false, got %v", result)
		}
	})
}
