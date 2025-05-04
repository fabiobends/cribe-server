package middlewares

import (
	"net/http/httptest"
	"testing"
)

func TestNotFoundMiddleware(t *testing.T) {
	t.Run("should return true for unknown routes", func(t *testing.T) {
		request := httptest.NewRequest("GET", "/asdf", nil)
		response := httptest.NewRecorder()
		result := NotFoundMiddleware(response, request)

		if result != true {
			t.Errorf("Expected true, got %v", result)
		}
	})

	t.Run("should return false for known routes", func(t *testing.T) {
		request := httptest.NewRequest("GET", "/users", nil)
		response := httptest.NewRecorder()
		result := NotFoundMiddleware(response, request)

		if result != false {
			t.Errorf("Expected false, got %v", result)
		}
	})
}
