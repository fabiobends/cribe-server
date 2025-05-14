package middlewares

import (
	"context"
	"net/http"

	"cribeapp.com/cribe-server/internal/routes/auth"
	"cribeapp.com/cribe-server/internal/utils"
)

func MainMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set the content type to json
		w.Header().Set("Content-Type", "application/json")

		// Check if the route is public or unknown
		isPrivate := PrivateCheckerMiddleware(w, r)
		if isPrivate {
			tokenService := auth.NewTokenServiceReady()
			userToken, err := AuthMiddleware(w, r, tokenService)
			if err != nil {
				utils.EncodeResponse(w, http.StatusUnauthorized, err)
				return
			}

			// Add the user ID to the request context
			ctx := r.Context()
			ctx = context.WithValue(ctx, UserIDContextKey, userToken.UserID)
			r = r.WithContext(ctx)
		}

		// Go to next handler
		next.ServeHTTP(w, r)
	})
}
