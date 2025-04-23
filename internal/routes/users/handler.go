package users

import "cribeapp.com/cribe-server/internal/utils"

type UserHandler struct {
	service UserServiceInterface
}

func NewUserHandler(service UserServiceInterface) *UserHandler {
	return &UserHandler{service: service}
}

func (handler *UserHandler) PostUser(user UserDTO) (UserWithoutPassword, *utils.ErrorResponse) {
	return handler.service.PostUser(user)
}
