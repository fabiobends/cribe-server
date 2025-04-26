package users

import "cribeapp.com/cribe-server/internal/utils"

type UserHandler struct {
	service UserServiceInterface
}

func NewUserHandler(service UserServiceInterface) *UserHandler {
	return &UserHandler{service: service}
}

func (handler *UserHandler) PostUser(user UserDTO) (User, *utils.ErrorResponse) {
	return handler.service.CreateUser(user)
}
