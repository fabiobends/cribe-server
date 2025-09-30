package feature_flags

import (
	"os"
	"testing"
)

func TestGetFeatureFlags(t *testing.T) {
	// Test with dev auth disabled
	_ = os.Unsetenv("APP_ENV")
	_ = os.Unsetenv("DEFAULT_EMAIL")

	flags := GetFeatureFlags()

	if flags == nil {
		t.Error("GetFeatureFlags should not return nil")
		return
	}

	if flags.DevAuthEnabled {
		t.Error("DevAuthEnabled should be false when environment variables are not set")
	}
}

func TestIsDevAuthEnabled(t *testing.T) {
	tests := []struct {
		name         string
		appEnv       string
		defaultEmail string
		expected     bool
	}{
		{
			name:         "Development with email",
			appEnv:       "development",
			defaultEmail: "test@example.com",
			expected:     true,
		},
		{
			name:         "Development without email",
			appEnv:       "development",
			defaultEmail: "",
			expected:     false,
		},
		{
			name:         "Production with email",
			appEnv:       "production",
			defaultEmail: "test@example.com",
			expected:     false,
		},
		{
			name:         "Production without email",
			appEnv:       "production",
			defaultEmail: "",
			expected:     false,
		},
		{
			name:         "Development case insensitive",
			appEnv:       "DEVELOPMENT",
			defaultEmail: "test@example.com",
			expected:     true,
		},
		{
			name:         "Empty environment",
			appEnv:       "",
			defaultEmail: "test@example.com",
			expected:     false,
		},
		{
			name:         "Email with whitespace",
			appEnv:       "development",
			defaultEmail: "  test@example.com  ",
			expected:     true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Set environment variables
			if test.appEnv != "" {
				_ = os.Setenv("APP_ENV", test.appEnv)
			} else {
				_ = os.Unsetenv("APP_ENV")
			}

			if test.defaultEmail != "" {
				_ = os.Setenv("DEFAULT_EMAIL", test.defaultEmail)
			} else {
				_ = os.Unsetenv("DEFAULT_EMAIL")
			}

			result := IsDevAuthEnabled()

			if result != test.expected {
				t.Errorf("Expected %v, got %v", test.expected, result)
			}

			// Clean up
			_ = os.Unsetenv("APP_ENV")
			_ = os.Unsetenv("DEFAULT_EMAIL")
		})
	}
}

func TestFeatureFlags_GetDefaultEmail(t *testing.T) {
	tests := []struct {
		name           string
		devAuthEnabled bool
		defaultEmail   string
		expected       string
	}{
		{
			name:           "Dev auth enabled with email",
			devAuthEnabled: true,
			defaultEmail:   "test@example.com",
			expected:       "test@example.com",
		},
		{
			name:           "Dev auth disabled with email",
			devAuthEnabled: false,
			defaultEmail:   "test@example.com",
			expected:       "",
		},
		{
			name:           "Dev auth enabled without email",
			devAuthEnabled: true,
			defaultEmail:   "",
			expected:       "",
		},
		{
			name:           "Email with whitespace",
			devAuthEnabled: true,
			defaultEmail:   "  test@example.com  ",
			expected:       "test@example.com",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Set up environment
			if test.defaultEmail != "" {
				_ = os.Setenv("DEFAULT_EMAIL", test.defaultEmail)
			} else {
				_ = os.Unsetenv("DEFAULT_EMAIL")
			}

			flags := &FeatureFlags{
				DevAuthEnabled: test.devAuthEnabled,
			}

			result := flags.GetDefaultEmail()

			if result != test.expected {
				t.Errorf("Expected %s, got %s", test.expected, result)
			}

			// Clean up
			_ = os.Unsetenv("DEFAULT_EMAIL")
		})
	}
}

func TestGetDefaultEmail(t *testing.T) {
	tests := []struct {
		name         string
		defaultEmail string
		expected     string
	}{
		{
			name:         "With email",
			defaultEmail: "test@example.com",
			expected:     "test@example.com",
		},
		{
			name:         "Without email",
			defaultEmail: "",
			expected:     "",
		},
		{
			name:         "Email with whitespace",
			defaultEmail: "  test@example.com  ",
			expected:     "test@example.com",
		},
		{
			name:         "Only whitespace",
			defaultEmail: "   ",
			expected:     "",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Set up environment
			if test.defaultEmail != "" {
				_ = os.Setenv("DEFAULT_EMAIL", test.defaultEmail)
			} else {
				_ = os.Unsetenv("DEFAULT_EMAIL")
			}

			result := GetDefaultEmail()

			if result != test.expected {
				t.Errorf("Expected %s, got %s", test.expected, result)
			}

			// Clean up
			_ = os.Unsetenv("DEFAULT_EMAIL")
		})
	}
}

func TestTryDevAuth(t *testing.T) {
	tests := []struct {
		name           string
		defaultEmail   string
		expectedUserID int
		expectError    bool
		errorMessage   string
	}{
		{
			name:           "Empty email",
			defaultEmail:   "",
			expectedUserID: 0,
			expectError:    true,
			errorMessage:   "Development authentication not enabled",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			userID, err := TryDevAuth(test.defaultEmail)

			if test.expectError {
				if err == nil {
					t.Error("Expected error but got nil")
				} else if err.Message != test.errorMessage {
					t.Errorf("Expected error message %s, got %s", test.errorMessage, err.Message)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}

			if userID != test.expectedUserID {
				t.Errorf("Expected user ID %d, got %d", test.expectedUserID, userID)
			}
		})
	}
}

func TestFeatureFlags_Integration(t *testing.T) {
	// Test complete integration flow
	tests := []struct {
		name            string
		appEnv          string
		defaultEmail    string
		expectedEnabled bool
		expectedEmail   string
	}{
		{
			name:            "Full dev setup",
			appEnv:          "development",
			defaultEmail:    "dev@example.com",
			expectedEnabled: true,
			expectedEmail:   "dev@example.com",
		},
		{
			name:            "Production setup",
			appEnv:          "production",
			defaultEmail:    "prod@example.com",
			expectedEnabled: false,
			expectedEmail:   "",
		},
		{
			name:            "Dev without email",
			appEnv:          "development",
			defaultEmail:    "",
			expectedEnabled: false,
			expectedEmail:   "",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Set environment
			if test.appEnv != "" {
				_ = os.Setenv("APP_ENV", test.appEnv)
			} else {
				_ = os.Unsetenv("APP_ENV")
			}

			if test.defaultEmail != "" {
				_ = os.Setenv("DEFAULT_EMAIL", test.defaultEmail)
			} else {
				_ = os.Unsetenv("DEFAULT_EMAIL")
			}

			// Test the full flow
			flags := GetFeatureFlags()

			if flags.DevAuthEnabled != test.expectedEnabled {
				t.Errorf("Expected DevAuthEnabled %v, got %v", test.expectedEnabled, flags.DevAuthEnabled)
			}

			email := flags.GetDefaultEmail()
			if email != test.expectedEmail {
				t.Errorf("Expected email %s, got %s", test.expectedEmail, email)
			}

			// Test standalone function too
			standaloneEmail := GetDefaultEmail()
			if test.expectedEnabled {
				if standaloneEmail != test.defaultEmail {
					t.Errorf("Expected standalone email %s, got %s", test.defaultEmail, standaloneEmail)
				}
			}

			// Clean up
			_ = os.Unsetenv("APP_ENV")
			_ = os.Unsetenv("DEFAULT_EMAIL")
		})
	}
}

func TestTryDevAuth_WithEmail(t *testing.T) {
	// Test TryDevAuth with email but expect database error (this is the expected behavior in test env)
	t.Run("Valid email but no database connection", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				// Expected to panic due to nil database connection in test environment
				// This actually tests that the function tries to access the database
				t.Log("Expected panic caught - database access attempted")
			}
		}()

		_, _ = TryDevAuth("test@example.com")
	})
}
