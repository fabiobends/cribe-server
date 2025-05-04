package middlewares

import (
	"net/http"
	"strings"
)

var knownRoutes = []string{
	"/status",
	"/auth",
	"/users",
	"/migrations",
}

func NotFoundMiddleware(w http.ResponseWriter, r *http.Request) bool {
	for _, route := range knownRoutes {
		if strings.HasPrefix(r.URL.Path, route) {
			return false
		}
	}
	return true
}
