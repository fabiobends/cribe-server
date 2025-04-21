package users

type UserHandler struct {
	service UserServiceInterface
}

func NewUserHandler(service UserServiceInterface) *UserHandler {
	return &UserHandler{service: service}
}

func (handler *UserHandler) PostUser(user UserDTO) (UserWithoutPassword, error) {
	return handler.service.PostUser(user)
}
