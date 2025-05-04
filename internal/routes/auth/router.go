package auth

import (
	"net/http"

	"cribeapp.com/cribe-server/internal/routes/users"
)

func HandleHTTPRequests(w http.ResponseWriter, r *http.Request) {
	repo := *users.NewUserRepository()
	service := users.NewUserService(repo)
	handler := users.NewUserHandler(service)

	handler.HandleRequest(w, r)
}
