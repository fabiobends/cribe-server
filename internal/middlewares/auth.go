package middlewares

import (
	"net/http"
	"strings"

	"cribeapp.com/cribe-server/internal/core/logger"
	"cribeapp.com/cribe-server/internal/errors"
	"cribeapp.com/cribe-server/internal/routes/auth"
)

// Define the context key at package level
type contextUserIDKey string

const UserIDContextKey = contextUserIDKey("user_id")

// AuthMiddleware extracts and validates the JWT token from the Authorization header
func AuthMiddleware(w http.ResponseWriter, r *http.Request, tokenService auth.TokenService) (*auth.JWTObject, *errors.ErrorResponse) {
	authLogger := logger.NewMiddlewareLogger("AuthMiddleware")

	authLogger.Debug("Extracting authorization token from request")

	// Get the Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		authLogger.Warn("Authorization header missing")
		return nil, &errors.ErrorResponse{
			Message: errors.Unauthorized,
			Details: "Authorization header is required",
		}
	}

	// Extract the token from the Bearer scheme
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		authLogger.Warn("Invalid authorization header format", map[string]interface{}{
			"header": authHeader,
		})
		return nil, &errors.ErrorResponse{
			Message: errors.Unauthorized,
			Details: "Authorization header must be in the format 'Bearer <token>'",
		}
	}

	token := parts[1]
	authLogger.Debug("Token extracted from authorization header")

	// Validate the token
	userToken, err := tokenService.ValidateToken(token)
	if err != nil {
		authLogger.Warn("Token validation failed", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, &errors.ErrorResponse{
			Message: errors.Unauthorized,
			Details: "Invalid or expired token",
		}
	}

	authLogger.Info("Token validation successful", map[string]interface{}{
		"userID": userToken.UserID,
	})

	return userToken, nil
}
