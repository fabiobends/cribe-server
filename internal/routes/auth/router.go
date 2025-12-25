//go:build !test

package auth

import (
	"net/http"

	"cribeapp.com/cribe-server/internal/routes/users"
)

func HandleHTTPRequests() func(http.ResponseWriter, *http.Request) {
	repo := users.NewUserRepository()
	tokenService := NewTokenServiceReady()
	service := NewAuthService(repo, tokenService)
	handler := NewAuthHandler(service)

	return handler.HandleRequest
}
