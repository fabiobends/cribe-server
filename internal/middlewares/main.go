package middlewares

import (
	"log"
	"net/http"

	"cribeapp.com/cribe-server/internal/routes/auth"
	"cribeapp.com/cribe-server/internal/utils"
)

func MainMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set the content type to json
		w.Header().Set("Content-Type", "application/json")

		// Apply the middlewares
		ok := PublicMiddleware(w, r)
		if ok {
			log.Println("Public route")
			next.ServeHTTP(w, r)
			return
		}

		ok = NotFoundMiddleware(w, r)
		if ok {
			log.Println("Not found")
			utils.NotFound(w, r)
			return
		}

		tokenService := auth.NewTokenServiceReady()
		r, ok = AuthMiddleware(w, r, tokenService)
		if !ok {
			log.Println("Unauthorized")
			utils.EncodeResponse(w, http.StatusUnauthorized, utils.ErrorResponse{
				StatusCode: http.StatusUnauthorized,
				Message:    "Unauthorized",
				Details:    []string{"You are not authorized to access this resource"},
			})
			return
		}

		next.ServeHTTP(w, r)
	})
}
