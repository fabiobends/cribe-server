package users

import (
	"cribeapp.com/cribe-server/internal/utils"
)

type UserService struct {
	repo UserRepository
}

func NewUserService(repo UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (service *UserService) CreateUser(user UserDTO) (User, *utils.ErrorResponse) {
	// Validate using domain validation
	if err := user.Validate(); err != nil {
		return User{}, err
	}

	// Create user in repository
	result, err := service.repo.CreateUser(user)
	if err != nil {
		return User{}, &utils.ErrorResponse{
			Message: utils.DatabaseError,
			Details: err.Error(),
		}
	}

	// Return sanitized user without sensitive data
	return service.sanitizeUser(result), nil
}

func (service *UserService) GetUserById(id int) (User, *utils.ErrorResponse) {
	result, err := service.repo.GetUserById(id)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return User{}, &utils.ErrorResponse{
				Message: utils.UserNotFound,
				Details: "The requested user was not found",
			}
		}
		return User{}, &utils.ErrorResponse{
			Message: utils.DatabaseError,
			Details: err.Error(),
		}
	}

	return service.sanitizeUser(result), nil
}

func (service *UserService) GetUsers() ([]User, *utils.ErrorResponse) {
	result, err := service.repo.GetUsers()
	if err != nil {
		return nil, &utils.ErrorResponse{
			Message: utils.DatabaseError,
			Details: err.Error(),
		}
	}

	users := make([]User, len(result))
	for i, user := range result {
		users[i] = service.sanitizeUser(user)
	}

	return users, nil
}

// sanitizeUser removes sensitive data from the user object
func (service *UserService) sanitizeUser(user UserWithPassword) User {
	return User{
		ID:        user.ID,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}
