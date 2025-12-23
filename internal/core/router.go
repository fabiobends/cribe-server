//go:build !test

package core

import (
	"net/http"
	"time"

	"cribeapp.com/cribe-server/internal/core/logger"
	"cribeapp.com/cribe-server/internal/middlewares"
	"cribeapp.com/cribe-server/internal/routes/auth"
	"cribeapp.com/cribe-server/internal/routes/migrations"
	"cribeapp.com/cribe-server/internal/routes/podcasts"
	"cribeapp.com/cribe-server/internal/routes/quizzes"
	"cribeapp.com/cribe-server/internal/routes/status"
	"cribeapp.com/cribe-server/internal/routes/transcripts"
	"cribeapp.com/cribe-server/internal/routes/users"
	"cribeapp.com/cribe-server/internal/utils"
)

type Response struct {
	Message string `json:"message"`
}

// registerRoute registers both with and without trailing slash
func registerRoute(mux *http.ServeMux, path string, handler func(http.ResponseWriter, *http.Request)) {
	mux.HandleFunc(path, handler)
	if path[len(path)-1] != '/' {
		mux.HandleFunc(path+"/", handler)
	}
}

func Handler(port string) error {
	log := logger.NewCoreLogger("Router")

	log.Info("Initializing HTTP router and routes")

	mux := http.NewServeMux()

	// Register routes
	registerRoute(mux, "/auth", auth.HandleHTTPRequests)
	registerRoute(mux, "/migrations", migrations.HandleHTTPRequests)
	registerRoute(mux, "/podcasts", podcasts.HandleHTTPRequests)
	registerRoute(mux, "/quizzes", quizzes.HandleHTTPRequests)
	registerRoute(mux, "/status", status.HandleHTTPRequests)
	registerRoute(mux, "/transcripts", transcripts.HandleHTTPRequests)
	registerRoute(mux, "/users", users.HandleHTTPRequests)
	mux.HandleFunc("/", utils.NotFound)

	log.Debug("Registered routes", map[string]any{
		"routes": []string{"/auth/", "/migrations", "/podcasts/", "/quizzes/", "/status/", "/transcripts/", "/users/", "/"},
	})

	muxWithMiddleware := middlewares.MainMiddleware(mux)

	server := &http.Server{
		Addr:           ":" + port,
		Handler:        muxWithMiddleware,
		ReadTimeout:    15 * time.Second,
		WriteTimeout:   4 * time.Hour, // Long timeout for SSE streaming of transcripts
		IdleTimeout:    60 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1 MB
	}

	log.Info("Starting HTTP server", map[string]any{
		"port": port,
	})

	return server.ListenAndServe()
}
