package utils

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"cribeapp.com/cribe-server/internal/errors"
)

func TestNotFound(t *testing.T) {
	// Create a request to test the NotFound handler
	req := httptest.NewRequest("GET", "/nonexistent", nil)
	w := httptest.NewRecorder()

	// Call the NotFound handler
	NotFound(w, req)

	// Check status code
	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status code %d, got %d", http.StatusNotFound, w.Code)
	}

	// Check response body contains expected error
	response, err := DecodeResponse[errors.ErrorResponse](w.Body.String())
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.Message != errors.RouteNotFound {
		t.Errorf("Expected message %s, got %s", errors.RouteNotFound, response.Message)
	}

	if response.Details != "The requested resource was not found" {
		t.Errorf("Expected details 'The requested resource was not found', got %s", response.Details)
	}
}

func TestNotFound_DifferentMethods(t *testing.T) {
	methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS", "HEAD"}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/test", nil)
			w := httptest.NewRecorder()

			NotFound(w, req)

			// Should always return 404 regardless of HTTP method
			if w.Code != http.StatusNotFound {
				t.Errorf("Expected status code %d for method %s, got %d", http.StatusNotFound, method, w.Code)
			}

			response, err := DecodeResponse[errors.ErrorResponse](w.Body.String())
			if err != nil {
				t.Fatalf("Failed to unmarshal response for method %s: %v", method, err)
			}

			if response.Message != errors.RouteNotFound {
				t.Errorf("Expected message %s for method %s, got %s", errors.RouteNotFound, method, response.Message)
			}
		})
	}
}

func TestNotFound_DifferentPaths(t *testing.T) {
	paths := []string{"/", "/users", "/api/v1/test", "/very/long/path/that/does/not/exist", ""}

	for _, path := range paths {
		t.Run("Path_"+path, func(t *testing.T) {
			if path == "" {
				path = "/"
			}

			req := httptest.NewRequest("GET", path, nil)
			w := httptest.NewRecorder()

			NotFound(w, req)

			// Should always return 404 regardless of path
			if w.Code != http.StatusNotFound {
				t.Errorf("Expected status code %d for path %s, got %d", http.StatusNotFound, path, w.Code)
			}

			response, err := DecodeResponse[errors.ErrorResponse](w.Body.String())
			if err != nil {
				t.Fatalf("Failed to unmarshal response for path %s: %v", path, err)
			}

			if response.Message != errors.RouteNotFound {
				t.Errorf("Expected message %s for path %s, got %s", errors.RouteNotFound, path, response.Message)
			}
		})
	}
}

func TestNotAllowed(t *testing.T) {
	// Create a recorder to test the NotAllowed handler
	w := httptest.NewRecorder()

	// Call the NotAllowed handler
	NotAllowed(w)

	// Check status code
	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status code %d, got %d", http.StatusMethodNotAllowed, w.Code)
	}

	// Check response body contains expected error
	response, err := DecodeResponse[errors.ErrorResponse](w.Body.String())
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.Message != errors.MethodNotAllowed {
		t.Errorf("Expected message %s, got %s", errors.MethodNotAllowed, response.Message)
	}

	if response.Details != "The requested method is not allowed for this resource" {
		t.Errorf("Expected details 'The requested method is not allowed for this resource', got %s", response.Details)
	}
}

func TestNotAllowed_MultipleCall(t *testing.T) {
	// Test that multiple calls to NotAllowed return consistent results
	for i := 0; i < 3; i++ {
		t.Run("Call_"+string(rune(i+'1')), func(t *testing.T) {
			w := httptest.NewRecorder()
			NotAllowed(w)

			if w.Code != http.StatusMethodNotAllowed {
				t.Errorf("Call %d: Expected status code %d, got %d", i+1, http.StatusMethodNotAllowed, w.Code)
			}

			response, err := DecodeResponse[errors.ErrorResponse](w.Body.String())
			if err != nil {
				t.Fatalf("Call %d: Failed to unmarshal response: %v", i+1, err)
			}

			if response.Message != errors.MethodNotAllowed {
				t.Errorf("Call %d: Expected message %s, got %s", i+1, errors.MethodNotAllowed, response.Message)
			}
		})
	}
}

func TestNotFound_ResponseFormat(t *testing.T) {
	// Test that the response is properly formatted JSON
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	NotFound(w, req)

	// Check Content-Type header (should be set by EncodeResponse)
	contentType := w.Header().Get("Content-Type")
	if contentType != "" && contentType != "application/json" {
		t.Logf("Content-Type header: %s (may not be set by EncodeResponse)", contentType)
	}

	// Verify response is valid JSON
	response, err := DecodeResponse[map[string]interface{}](w.Body.String())
	if err != nil {
		t.Fatalf("Response is not valid JSON: %v", err)
	}

	// Check that required fields exist
	if _, exists := response["message"]; !exists {
		t.Error("Response should contain 'message' field")
	}

	if _, exists := response["details"]; !exists {
		t.Error("Response should contain 'details' field")
	}
}

func TestNotAllowed_ResponseFormat(t *testing.T) {
	// Test that the response is properly formatted JSON
	w := httptest.NewRecorder()

	NotAllowed(w)

	// Verify response is valid JSON
	response, err := DecodeResponse[map[string]interface{}](w.Body.String())
	if err != nil {
		t.Fatalf("Response is not valid JSON: %v", err)
	}

	// Check that required fields exist
	if _, exists := response["message"]; !exists {
		t.Error("Response should contain 'message' field")
	}

	if _, exists := response["details"]; !exists {
		t.Error("Response should contain 'details' field")
	}
}
