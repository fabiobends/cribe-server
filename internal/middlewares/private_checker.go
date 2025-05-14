package middlewares

import (
	"log"
	"net/http"
	"strings"
)

var privateRoutes = map[string]bool{
	"/users": true,
}

func isPrivateRoute(path string) bool {
	return privateRoutes[path]
}

// Check if the route should go to the next handler
func PrivateCheckerMiddleware(w http.ResponseWriter, r *http.Request) bool {
	log.Println("Public middleware")
	log.Println("Requested path", r.URL.Path)
	path := strings.TrimSuffix(r.URL.Path, "/")
	log.Println("Trimmed path", path)

	if path == "/migrations" {
		return r.Header.Get("x-migration-run") != "true"
	}

	return isPrivateRoute(path)
}
