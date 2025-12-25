package users

import "net/http"

func HandleHTTPRequests() func(http.ResponseWriter, *http.Request) {
	repo := *NewUserRepository()
	service := NewUserService(repo)
	handler := NewUserHandler(service)

	return handler.HandleRequest
}
