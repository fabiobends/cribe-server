package users

import "log"

type UserServiceInterface interface {
	PostUser(user UserDTO) (UserWithoutPassword, error)
}

type UserService struct {
	repo UserRepository
}

func NewUserService(repo UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (service *UserService) PostUser(user UserDTO) (UserWithoutPassword, error) {
	result, err := service.repo.CreateUser(user)
	if err != nil {
		log.Println(err)
	}
	return UserWithoutPassword{
		ID:        result.ID,
		FirstName: result.FirstName,
		LastName:  result.LastName,
		Email:     result.Email,
		CreatedAt: result.CreatedAt,
		UpdatedAt: result.UpdatedAt,
	}, nil
}
