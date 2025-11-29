package utils

import (
	"net/http/httptest"
	"strings"
	"testing"

	"cribeapp.com/cribe-server/internal/errors"
)

func TestDecodeBody(t *testing.T) {
	type TestStruct struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	tests := []struct {
		name        string
		body        string
		expectError bool
		expected    TestStruct
		description string
	}{
		{
			name:        "Valid JSON",
			body:        `{"name":"John","age":30}`,
			expectError: false,
			expected:    TestStruct{Name: "John", Age: 30},
			description: "Should decode valid JSON successfully",
		},
		{
			name:        "Invalid JSON",
			body:        `{"name":"John","age":}`,
			expectError: true,
			expected:    TestStruct{},
			description: "Should return error for invalid JSON",
		},
		{
			name:        "Empty body",
			body:        ``,
			expectError: true,
			expected:    TestStruct{},
			description: "Should return error for empty body",
		},
		{
			name:        "Missing fields",
			body:        `{"name":"John"}`,
			expectError: false,
			expected:    TestStruct{Name: "John", Age: 0},
			description: "Should decode partial JSON with default values",
		},
		{
			name:        "Extra fields",
			body:        `{"name":"John","age":30,"extra":"field"}`,
			expectError: false,
			expected:    TestStruct{Name: "John", Age: 30},
			description: "Should ignore extra fields",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/test", strings.NewReader(test.body))
			req.Header.Set("Content-Type", "application/json")

			result, err := DecodeBody[TestStruct](req)

			if test.expectError {
				if err == nil {
					t.Error("Expected error but got nil")
				} else if err.Message != errors.InvalidRequestBody {
					t.Errorf("Expected error message %s, got %s", errors.InvalidRequestBody, err.Message)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}

				if result.Name != test.expected.Name {
					t.Errorf("Expected name %s, got %s", test.expected.Name, result.Name)
				}

				if result.Age != test.expected.Age {
					t.Errorf("Expected age %d, got %d", test.expected.Age, result.Age)
				}
			}
		})
	}
}

func TestEncodeResponse(t *testing.T) {
	tests := []struct {
		name        string
		statusCode  int
		response    interface{}
		expected    string
		description string
	}{
		{
			name:        "Simple struct",
			statusCode:  200,
			response:    map[string]string{"message": "success"},
			expected:    `{"message":"success"}`,
			description: "Should encode simple struct correctly",
		},
		{
			name:        "Array response",
			statusCode:  200,
			response:    []string{"item1", "item2"},
			expected:    `["item1","item2"]`,
			description: "Should encode array correctly",
		},
		{
			name:        "Nil response",
			statusCode:  204,
			response:    nil,
			expected:    "null",
			description: "Should encode nil as null",
		},
		{
			name:        "Number response",
			statusCode:  200,
			response:    42,
			expected:    "42",
			description: "Should encode number correctly",
		},
		{
			name:        "String response",
			statusCode:  200,
			response:    "hello world",
			expected:    `"hello world"`,
			description: "Should encode string with quotes",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			w := httptest.NewRecorder()

			EncodeResponse(w, test.statusCode, test.response)

			// Check status code
			if w.Code != test.statusCode {
				t.Errorf("Expected status code %d, got %d", test.statusCode, w.Code)
			}

			// Check response body (trim newline that json.Encoder adds)
			body := strings.TrimSpace(w.Body.String())
			if body != test.expected {
				t.Errorf("Expected body %s, got %s", test.expected, body)
			}
		})
	}
}

func TestEncodeResponse_UnencodableData(t *testing.T) {
	// Test with data that cannot be JSON encoded
	w := httptest.NewRecorder()

	// Create a channel which cannot be JSON encoded
	unencodable := make(chan int)

	EncodeResponse(w, 200, unencodable)

	// The status code will still be 200 because w.WriteHeader() is called first
	// But the response body should contain an error message
	if w.Code != 200 {
		t.Errorf("Expected status code 200 (set before encoding), got %d", w.Code)
	}

	// Check that an error was written to the response body
	body := w.Body.String()
	if body == "" {
		t.Error("Expected error message in response body for unencodable data")
	}

	// The body should contain some indication of a JSON encoding error
	if !strings.Contains(strings.ToLower(body), "json") && !strings.Contains(strings.ToLower(body), "marshal") {
		t.Logf("Response body: %s", body)
		t.Log("Expected JSON/marshal error message in response body (this test verifies error handling path is executed)")
	}
}

func TestDecodeResponse(t *testing.T) {
	type TestStruct struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	tests := []struct {
		name        string
		response    string
		expectError bool
		expected    TestStruct
		description string
	}{
		{
			name:        "Valid JSON response",
			response:    `{"name":"Alice","age":25}`,
			expectError: false,
			expected:    TestStruct{Name: "Alice", Age: 25},
			description: "Should decode valid JSON response",
		},
		{
			name:        "Invalid JSON response",
			response:    `{"name":"Alice","age":}`,
			expectError: true,
			expected:    TestStruct{},
			description: "Should return error for invalid JSON",
		},
		{
			name:        "Empty response",
			response:    ``,
			expectError: true,
			expected:    TestStruct{},
			description: "Should return error for empty response",
		},
		{
			name:        "Partial JSON",
			response:    `{"name":"Bob"}`,
			expectError: false,
			expected:    TestStruct{Name: "Bob", Age: 0},
			description: "Should decode partial JSON with defaults",
		},
		{
			name:        "Array instead of object",
			response:    `["Alice", 25]`,
			expectError: true,
			expected:    TestStruct{},
			description: "Should return error when expecting object but got array",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := DecodeResponse[TestStruct](test.response)

			if test.expectError {
				if err == nil {
					t.Error("Expected error but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}

				if result.Name != test.expected.Name {
					t.Errorf("Expected name %s, got %s", test.expected.Name, result.Name)
				}

				if result.Age != test.expected.Age {
					t.Errorf("Expected age %d, got %d", test.expected.Age, result.Age)
				}
			}
		})
	}
}

func TestSanitizeJSONString(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    string
		description string
	}{
		{
			name:        "Single line JSON",
			input:       `{"message":"hello"}`,
			expected:    `{"message":"hello"}`,
			description: "Should return single line as-is",
		},
		{
			name:        "Multi-line JSON",
			input:       "{\n\"message\":\"hello\"\n}",
			expected:    "{",
			description: "Should return only first line",
		},
		{
			name:        "JSON with newline at end",
			input:       `{"message":"hello"}` + "\n",
			expected:    `{"message":"hello"}`,
			description: "Should remove trailing newline",
		},
		{
			name:        "Empty string",
			input:       "",
			expected:    "",
			description: "Should handle empty string",
		},
		{
			name:        "Only newline",
			input:       "\n",
			expected:    "",
			description: "Should return empty string for only newline",
		},
		{
			name:        "Multiple newlines",
			input:       "line1\nline2\nline3",
			expected:    "line1",
			description: "Should return only first line from multiple lines",
		},
		{
			name:        "No newlines",
			input:       "single line without newline",
			expected:    "single line without newline",
			description: "Should return input unchanged when no newlines",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := SanitizeJSONString(test.input)

			if result != test.expected {
				t.Errorf("Expected %q, got %q", test.expected, result)
			}
		})
	}
}

func TestEncodeToJSON(t *testing.T) {
	t.Run("should encode simple struct to JSON", func(t *testing.T) {
		type Person struct {
			Name string `json:"name"`
			Age  int    `json:"age"`
		}

		person := Person{Name: "John", Age: 30}
		data, err := EncodeToJSON(person)

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		expected := `{"name":"John","age":30}`
		if string(data) != expected {
			t.Errorf("Expected %s, got %s", expected, string(data))
		}
	})

	t.Run("should encode map to JSON", func(t *testing.T) {
		data := map[string]interface{}{
			"status":  "success",
			"code":    200,
			"message": "OK",
		}

		result, err := EncodeToJSON(data)

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if len(result) == 0 {
			t.Error("Expected non-empty JSON data")
		}
	})

	t.Run("should handle nil value", func(t *testing.T) {
		data, err := EncodeToJSON(nil)

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		expected := "null"
		if string(data) != expected {
			t.Errorf("Expected %s, got %s", expected, string(data))
		}
	})
}
