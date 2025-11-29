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
	"cribeapp.com/cribe-server/internal/routes/status"
	"cribeapp.com/cribe-server/internal/routes/transcripts"
	"cribeapp.com/cribe-server/internal/routes/users"
	"cribeapp.com/cribe-server/internal/utils"
)

type Response struct {
	Message string `json:"message"`
}

func Handler(port string) error {
	log := logger.NewCoreLogger("Router")

	log.Info("Initializing HTTP router and routes")

	mux := http.NewServeMux()

	// Register routes
	mux.HandleFunc("/auth/", auth.HandleHTTPRequests)
	mux.HandleFunc("/migrations", migrations.HandleHTTPRequests)
	mux.HandleFunc("/podcasts/", podcasts.HandleHTTPRequests)
	mux.HandleFunc("/status/", status.HandleHTTPRequests)
	mux.HandleFunc("/transcripts/", transcripts.HandleHTTPRequests)
	mux.HandleFunc("/users/", users.HandleHTTPRequests)
	mux.HandleFunc("/", utils.NotFound)

	log.Debug("Registered routes", map[string]any{
		"routes": []string{"/auth/", "/migrations", "/podcasts/", "/status/", "/transcripts/", "/users/", "/"},
	})

	muxWithMiddleware := middlewares.MainMiddleware(mux)

	server := &http.Server{
		Addr:           ":" + port,
		Handler:        muxWithMiddleware,
		ReadTimeout:    15 * time.Second,
		WriteTimeout:   15 * time.Second,
		IdleTimeout:    60 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1 MB
	}

	log.Info("Starting HTTP server", map[string]any{
		"port":           port,
		"readTimeout":    "15s",
		"writeTimeout":   "15s",
		"idleTimeout":    "60s",
		"maxHeaderBytes": "1MB",
	})

	return server.ListenAndServe()
}
