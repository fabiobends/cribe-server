package utils

import (
	"os"
	"testing"
)

func TestGetPort(t *testing.T) {
	tests := []struct {
		name        string
		portEnv     string
		expected    string
		description string
	}{
		{
			name:        "Default port when PORT not set",
			portEnv:     "",
			expected:    "8080",
			description: "Should return default port 8080 when PORT environment variable is empty",
		},
		{
			name:        "Custom port when PORT is set",
			portEnv:     "3000",
			expected:    "3000",
			description: "Should return the value from PORT environment variable",
		},
		{
			name:        "Port with leading/trailing spaces",
			portEnv:     " 9000 ",
			expected:    " 9000 ",
			description: "Should return PORT value as-is, including whitespace",
		},
		{
			name:        "Non-numeric port",
			portEnv:     "abc",
			expected:    "abc",
			description: "Should return PORT value even if not numeric (validation happens elsewhere)",
		},
		{
			name:        "Zero port",
			portEnv:     "0",
			expected:    "0",
			description: "Should return zero port if explicitly set",
		},
		{
			name:        "High port number",
			portEnv:     "65535",
			expected:    "65535",
			description: "Should return high port numbers",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Store original PORT value
			originalPort := os.Getenv("PORT")

			// Set test PORT value
			if test.portEnv == "" {
				_ = os.Unsetenv("PORT")
			} else {
				_ = os.Setenv("PORT", test.portEnv)
			}

			// Test the function
			result := GetPort()

			// Restore original PORT value
			if originalPort != "" {
				_ = os.Setenv("PORT", originalPort)
			} else {
				_ = os.Unsetenv("PORT")
			}

			// Verify result
			if result != test.expected {
				t.Errorf("Expected %s, got %s", test.expected, result)
			}
		})
	}
}

func TestGetPort_MultipleCallsConsistent(t *testing.T) {
	// Test that multiple calls return consistent results
	originalPort := os.Getenv("PORT")

	// Test with custom port
	_ = os.Setenv("PORT", "5000")

	result1 := GetPort()
	result2 := GetPort()
	result3 := GetPort()

	// Restore original PORT value
	if originalPort != "" {
		_ = os.Setenv("PORT", originalPort)
	} else {
		_ = os.Unsetenv("PORT")
	}

	if result1 != result2 || result2 != result3 {
		t.Errorf("Inconsistent results: %s, %s, %s", result1, result2, result3)
	}

	if result1 != "5000" {
		t.Errorf("Expected 5000, got %s", result1)
	}
}

func TestGetPort_EnvironmentIsolation(t *testing.T) {
	// Test that changes to environment during test don't affect each other
	originalPort := os.Getenv("PORT")

	// First test - set to 4000
	_ = os.Setenv("PORT", "4000")
	result1 := GetPort()

	// Second test - change to 7000
	_ = os.Setenv("PORT", "7000")
	result2 := GetPort()

	// Third test - unset PORT
	_ = os.Unsetenv("PORT")
	result3 := GetPort()

	// Restore original PORT value
	if originalPort != "" {
		_ = os.Setenv("PORT", originalPort)
	} else {
		_ = os.Unsetenv("PORT")
	}

	// Verify each result
	if result1 != "4000" {
		t.Errorf("First call: expected 4000, got %s", result1)
	}

	if result2 != "7000" {
		t.Errorf("Second call: expected 7000, got %s", result2)
	}

	if result3 != "8080" {
		t.Errorf("Third call: expected 8080, got %s", result3)
	}
}
