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
	log := logger.NewMiddlewareLogger("AuthMiddleware")

	log.Debug("Extracting authorization token from request")

	// Get the Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		log.Warn("Authorization header missing")
		return nil, &errors.ErrorResponse{
			Message: errors.Unauthorized,
			Details: "Authorization header is required",
		}
	}

	// Extract the token from the Bearer scheme
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		log.Warn("Invalid authorization header format", map[string]any{
			"header": authHeader,
		})
		return nil, &errors.ErrorResponse{
			Message: errors.Unauthorized,
			Details: "Authorization header must be in the format 'Bearer <token>'",
		}
	}

	token := parts[1]
	log.Debug("Token extracted from authorization header")

	// Validate the token
	userToken, err := tokenService.ValidateToken(token)
	if err != nil {
		log.Warn("Token validation failed", map[string]any{
			"error": err.Error(),
		})
		return nil, &errors.ErrorResponse{
			Message: errors.Unauthorized,
			Details: "Invalid or expired token",
		}
	}

	log.Info("Token validation successful", map[string]any{
		"userID": userToken.UserID,
	})

	return userToken, nil
}
