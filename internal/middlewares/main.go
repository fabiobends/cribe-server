package middlewares

import (
	"context"
	"log"
	"net/http"

	"cribeapp.com/cribe-server/internal/feature_flags"
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
			var userID int
			var authError *utils.ErrorResponse

			// Check if dev auth is enabled (using feature flags)
			if feature_flags.IsDevAuthEnabled() {
				defaultEmail := feature_flags.GetDefaultEmail()
				userID, authError = feature_flags.TryDevAuth(defaultEmail)
				if authError != nil {
					log.Printf("Dev auth failed for email %s: %v", defaultEmail, authError)
				}
			}

			// If dev auth didn't work or isn't enabled, use normal token auth
			if userID == 0 {
				tokenService := auth.NewTokenServiceReady()
				userToken, err := AuthMiddleware(w, r, tokenService)
				if err != nil {
					utils.EncodeResponse(w, http.StatusUnauthorized, err)
					return
				}
				userID = userToken.UserID
			}

			// Add the user ID to the request context
			ctx := r.Context()
			ctx = context.WithValue(ctx, UserIDContextKey, userID)
			r = r.WithContext(ctx)
		}

		// Go to next handler
		next.ServeHTTP(w, r)
	})
}
