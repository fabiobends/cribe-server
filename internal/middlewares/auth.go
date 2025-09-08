package middlewares

import (
	"log"
	"net/http"
	"strings"

	"cribeapp.com/cribe-server/internal/errors"
	"cribeapp.com/cribe-server/internal/routes/auth"
)

// Define the context key at package level
type contextUserIDKey string

const UserIDContextKey = contextUserIDKey("user_id")

func AuthMiddleware(w http.ResponseWriter, r *http.Request, tokenService auth.TokenService) (*auth.JWTObject, *errors.ErrorResponse) {
	// Check if the path is a private route
	log.Println("Requested path", r.URL.Path)

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return nil, &errors.ErrorResponse{
			Message: errors.InvalidAuthorizationHeader,
			Details: "You are not authorized to access this resource",
		}
	}

	// Check if the header is in the format "Bearer <token>"
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return nil, &errors.ErrorResponse{
			Message: errors.InvalidAuthorizationHeader,
			Details: "The authorization header is not in the correct format",
		}
	}

	tokenString := parts[1]

	// Parse and validate the token
	token, err := tokenService.ValidateToken(tokenString)

	if token == nil || err != nil || token.Typ != "access" {
		return nil, &errors.ErrorResponse{
			Message: errors.InvalidAuthorizationHeader,
			Details: "The token is invalid",
		}
	}

	return token, nil
}
