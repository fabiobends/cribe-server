package core

import (
	"net/http"
	"time"

	"cribeapp.com/cribe-server/internal/middlewares"
	"cribeapp.com/cribe-server/internal/routes/auth"
	"cribeapp.com/cribe-server/internal/routes/migrations"
	"cribeapp.com/cribe-server/internal/routes/status"
	"cribeapp.com/cribe-server/internal/routes/users"
	"cribeapp.com/cribe-server/internal/utils"
)

type Response struct {
	Message string `json:"message"`
}

func Handler(port string) error {
	mux := http.NewServeMux()

	mux.HandleFunc("/auth/", auth.HandleHTTPRequests)
	mux.HandleFunc("/migrations", migrations.HandleHTTPRequests)
	mux.HandleFunc("/status/", status.HandleHTTPRequests)
	mux.HandleFunc("/users/", users.HandleHTTPRequests)
	mux.HandleFunc("/", utils.NotFound)

	muxWithMiddleware := middlewares.MainMiddleware(mux)

	server := &http.Server{
		Addr:           ":" + port,
		Handler:        muxWithMiddleware,
		ReadTimeout:    15 * time.Second,
		WriteTimeout:   15 * time.Second,
		IdleTimeout:    60 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1 MB
	}

	return server.ListenAndServe()
}
