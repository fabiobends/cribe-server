package core

import (
	"net/http"

	"cribeapp.com/cribe-server/internal/routes/status"
	"cribeapp.com/cribe-server/internal/utils"
)

type Response struct {
	Message string `json:"message"`
}

func middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

func Handler(port string) error {
	mux := http.NewServeMux()

	mux.HandleFunc("/status", status.Handler)
	mux.HandleFunc("/", utils.NotFound)

	muxWithMiddleware := middleware(mux)

	return http.ListenAndServe(":"+port, muxWithMiddleware)
}
