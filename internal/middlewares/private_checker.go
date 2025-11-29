package middlewares

import (
	"net/http"
	"strings"

	"cribeapp.com/cribe-server/internal/core/logger"
)

var privateRoutes = map[string]bool{
	"/users":       true,
	"/podcasts":    true,
	"/transcripts": true,
}

func isPrivateRoute(path string) bool {
	return privateRoutes[path]
}

// Check if the route should go to the next handler
func PrivateCheckerMiddleware(w http.ResponseWriter, r *http.Request) bool {
	log := logger.NewMiddlewareLogger("PrivateCheckerMiddleware")

	log.Debug("Checking route privacy", map[string]any{
		"originalPath": r.URL.Path,
	})

	path := strings.TrimSuffix(r.URL.Path, "/")
	log.Debug("Path after trimming", map[string]any{
		"trimmedPath": path,
	})

	if path == "/migrations" {
		migrationHeader := r.Header.Get("x-migration-run")
		isPrivate := migrationHeader != "true"
		log.Debug("Migration route check", map[string]any{
			"header":    migrationHeader,
			"isPrivate": isPrivate,
		})
		return isPrivate
	}

	isPrivate := isPrivateRoute(path)
	log.Debug("Route privacy determined", map[string]any{
		"path":      path,
		"isPrivate": isPrivate,
	})

	return isPrivate
}
