package middlewares

import (
	"net/http"
	"strings"
)

var allowedRoutes = []string{
	"/status",
	"/auth",
}

func PublicMiddleware(w http.ResponseWriter, r *http.Request) bool {
	currentPath := r.URL.Path
	for _, route := range allowedRoutes {
		if strings.HasPrefix(currentPath, route) {
			return true
		}
	}
	return false
}
