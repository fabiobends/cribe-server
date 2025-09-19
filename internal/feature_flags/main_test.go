package feature_flags

import (
	"os"
	"testing"
)

func TestFeatureFlags_DevAuthEnabled(t *testing.T) {
	tests := []struct {
		name         string
		appEnv       string
		defaultEmail string
		expected     bool
	}{
		{
			name:         "should enable dev auth when APP_ENV=development and DEFAULT_EMAIL is set",
			appEnv:       "development",
			defaultEmail: "test@example.com",
			expected:     true,
		},
		{
			name:         "should not enable dev auth when APP_ENV is not development",
			appEnv:       "production",
			defaultEmail: "test@example.com",
			expected:     false,
		},
		{
			name:         "should not enable dev auth when DEFAULT_EMAIL is empty",
			appEnv:       "development",
			defaultEmail: "",
			expected:     false,
		},
		{
			name:         "should not enable dev auth when both conditions are not met",
			appEnv:       "production",
			defaultEmail: "",
			expected:     false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Set environment variables
			if err := os.Setenv("APP_ENV", test.appEnv); err != nil {
				t.Fatalf("Failed to set APP_ENV: %v", err)
			}
			if err := os.Setenv("DEFAULT_EMAIL", test.defaultEmail); err != nil {
				t.Fatalf("Failed to set DEFAULT_EMAIL: %v", err)
			}

			// Test the feature flag
			featureFlags := GetFeatureFlags()
			if featureFlags.DevAuthEnabled != test.expected {
				t.Errorf("DevAuthEnabled = %v, expected %v", featureFlags.DevAuthEnabled, test.expected)
			}

			// Test GetDefaultEmail method
			if test.expected {
				email := featureFlags.GetDefaultEmail()
				if email != test.defaultEmail {
					t.Errorf("GetDefaultEmail() = %v, expected %v", email, test.defaultEmail)
				}
			} else {
				email := featureFlags.GetDefaultEmail()
				if email != "" {
					t.Errorf("GetDefaultEmail() = %v, expected empty string when dev auth is disabled", email)
				}
			}

			// Test standalone functions
			if IsDevAuthEnabled() != test.expected {
				t.Errorf("IsDevAuthEnabled() = %v, expected %v", IsDevAuthEnabled(), test.expected)
			}

			if test.expected {
				email := GetDefaultEmail()
				if email != test.defaultEmail {
					t.Errorf("GetDefaultEmail() = %v, expected %v", email, test.defaultEmail)
				}
			}

			// Clean up
			if err := os.Unsetenv("APP_ENV"); err != nil {
				t.Logf("Warning: Failed to unset APP_ENV: %v", err)
			}
			if err := os.Unsetenv("DEFAULT_EMAIL"); err != nil {
				t.Logf("Warning: Failed to unset DEFAULT_EMAIL: %v", err)
			}
		})
	}
}
