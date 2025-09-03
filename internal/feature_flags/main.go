package feature_flags

import (
	"os"
	"strings"

	"cribeapp.com/cribe-server/internal/routes/users"
	"cribeapp.com/cribe-server/internal/utils"
)

// FeatureFlags holds all feature flag configurations
type FeatureFlags struct {
	DevAuthEnabled bool
}

// GetFeatureFlags returns the current feature flags configuration
func GetFeatureFlags() *FeatureFlags {
	return &FeatureFlags{
		DevAuthEnabled: IsDevAuthEnabled(),
	}
}

// IsDevAuthEnabled checks if development authentication is enabled
// It requires both APP_ENV=development and DEFAULT_EMAIL to be set
func IsDevAuthEnabled() bool {
	appEnv := strings.ToLower(os.Getenv("APP_ENV"))
	defaultEmail := strings.TrimSpace(os.Getenv("DEFAULT_EMAIL"))

	return appEnv == "development" && defaultEmail != ""
}

// GetDefaultEmail returns the DEFAULT_EMAIL environment variable if dev auth is enabled
func (ff *FeatureFlags) GetDefaultEmail() string {
	if !ff.DevAuthEnabled {
		return ""
	}
	return strings.TrimSpace(os.Getenv("DEFAULT_EMAIL"))
}

// GetDefaultEmail returns the DEFAULT_EMAIL environment variable
func GetDefaultEmail() string {
	return strings.TrimSpace(os.Getenv("DEFAULT_EMAIL"))
}

// TryDevAuth attempts to authenticate using the default email for development
func TryDevAuth(defaultEmail string) (int, *utils.ErrorResponse) {
	if defaultEmail == "" {
		return 0, &utils.ErrorResponse{
			Message: "dev_auth_failed",
			Details: "No default email provided",
		}
	}

	// Create a user repository to fetch the user
	userRepo := users.NewUserRepository()
	user, err := userRepo.GetUserByEmail(defaultEmail)
	if err != nil {
		return 0, &utils.ErrorResponse{
			Message: "dev_auth_failed",
			Details: "User not found with default email",
		}
	}

	return user.ID, nil
}
