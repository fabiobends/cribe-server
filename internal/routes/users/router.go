package users

import "net/http"

func HandleHTTPRequests(w http.ResponseWriter, r *http.Request) {
	repo := *NewUserRepository()
	service := NewUserService(repo)
	handler := NewUserHandler(service)

	handler.HandleRequest(w, r)
}
