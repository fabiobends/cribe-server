package middlewares

import (
	"context"
	"net/http"
	"time"

	"cribeapp.com/cribe-server/internal/core/logger"
	"cribeapp.com/cribe-server/internal/errors"
	"cribeapp.com/cribe-server/internal/feature_flags"
	"cribeapp.com/cribe-server/internal/routes/auth"
	"cribeapp.com/cribe-server/internal/utils"
)

func MainMiddleware(next http.Handler) http.Handler {
	log := logger.NewMiddlewareLogger("MainMiddleware")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		log.Debug("Processing request", map[string]interface{}{
			"method": r.Method,
			"path":   r.URL.Path,
			"ip":     r.RemoteAddr,
		})

		// Set the content type to json
		w.Header().Set("Content-Type", "application/json")

		// Check if the route is public or unknown
		isPrivate := PrivateCheckerMiddleware(w, r)
		if isPrivate {
			log.Debug("Route requires authentication", map[string]interface{}{
				"path": r.URL.Path,
			})

			var userID int
			var authError *errors.ErrorResponse

			// Check if dev auth is enabled (using feature flags)
			if feature_flags.IsDevAuthEnabled() {
				log.Debug("Dev auth is enabled, attempting dev authentication")
				defaultEmail := feature_flags.GetDefaultEmail()
				userID, authError = feature_flags.TryDevAuth(defaultEmail)
				if authError != nil {
					log.Warn("Dev auth failed", map[string]interface{}{
						"email": defaultEmail, // Will be automatically masked
						"error": authError.Details,
					})
				} else {
					log.Info("Dev auth successful", map[string]interface{}{
						"email":  defaultEmail, // Will be automatically masked
						"userID": userID,
					})
				}
			}

			// If dev auth didn't work or isn't enabled, use normal token auth
			if userID == 0 {
				log.Debug("Using token-based authentication")
				tokenService := auth.NewTokenServiceReady()
				if tokenService == nil {
					log.Error("Token service not configured")
					utils.EncodeResponse(w, http.StatusInternalServerError, &errors.ErrorResponse{
						Message: errors.InternalServerError,
						Details: "Token service not configured",
					})
					return
				}
				userToken, err := AuthMiddleware(w, r, tokenService)
				if err != nil {
					log.Warn("Token authentication failed", map[string]interface{}{
						"error": err.Details,
					})
					utils.EncodeResponse(w, http.StatusUnauthorized, err)
					return
				}
				userID = userToken.UserID
				log.Debug("Token authentication successful", map[string]interface{}{
					"userID": userID,
				})
			}

			// Add the user ID to the request context
			ctx := r.Context()
			ctx = context.WithValue(ctx, UserIDContextKey, userID)
			r = r.WithContext(ctx)
		} else {
			log.Debug("Route is public, no authentication required", map[string]interface{}{
				"path": r.URL.Path,
			})
		}

		// Go to next handler
		next.ServeHTTP(w, r)

		duration := time.Since(start)
		log.Debug("Request completed", map[string]interface{}{
			"path":     r.URL.Path,
			"method":   r.Method,
			"duration": duration.String(),
		})
	})
}
