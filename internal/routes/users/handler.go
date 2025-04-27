package users

import "cribeapp.com/cribe-server/internal/utils"

type UserHandler struct {
	service UserService
}

func NewUserHandler(service *UserService) *UserHandler {
	return &UserHandler{service: *service}
}

func (handler *UserHandler) PostUser(user UserDTO) (User, *utils.ErrorResponse) {
	return handler.service.CreateUser(user)
}

func (handler *UserHandler) GetUserById(id int) (User, *utils.ErrorResponse) {
	return handler.service.GetUserById(id)
}

func (handler *UserHandler) GetUsers() ([]User, *utils.ErrorResponse) {
	return handler.service.GetUsers()
}
