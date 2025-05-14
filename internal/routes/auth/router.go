package auth

import (
	"net/http"

	"cribeapp.com/cribe-server/internal/routes/users"
)

func HandleHTTPRequests(w http.ResponseWriter, r *http.Request) {
	repo := users.NewUserRepository()
	tokenService := NewTokenServiceReady()
	service := NewAuthService(repo, tokenService)
	handler := NewAuthHandler(service)

	handler.HandleRequest(w, r)
}
