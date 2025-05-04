package middlewares

import (
	"context"
	"net/http"
	"strings"

	"cribeapp.com/cribe-server/internal/routes/auth"
	"cribeapp.com/cribe-server/internal/utils"
)

// Define the context key at package level
type contextUserIDKey string

const UserIDContextKey = contextUserIDKey("user_id")

func AuthMiddleware(w http.ResponseWriter, r *http.Request, tokenService auth.TokenService) (*http.Request, bool) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		utils.NewErrorResponse(http.StatusUnauthorized, "Missing authorization header")
		return nil, false
	}

	// Check if the header is in the format "Bearer <token>"
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		utils.NewErrorResponse(http.StatusUnauthorized, "Invalid authorization header format")
		return nil, false
	}

	tokenString := parts[1]

	// Parse and validate the token
	token, err := tokenService.ValidateToken(tokenString)

	if token == nil || err != nil || token.Typ != "access" {
		utils.NewErrorResponse(http.StatusUnauthorized, "Invalid token")
		return nil, false
	}

	// Add the user ID to the request context
	ctx := r.Context()
	ctx = context.WithValue(ctx, UserIDContextKey, token.UserID)
	r = r.WithContext(ctx)

	return r, true
}
