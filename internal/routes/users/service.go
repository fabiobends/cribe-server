package users

import (
	"net/http"

	"cribeapp.com/cribe-server/internal/utils"
)

type UserService struct {
	repo UserRepository
}

func NewUserService(repo UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (service *UserService) CreateUser(user UserDTO) (User, *utils.ErrorResponse) {
	// Validate required fields
	missingFields := utils.ValidateRequiredFields(
		utils.ValidateRequiredField("first_name", user.FirstName),
		utils.ValidateRequiredField("last_name", user.LastName),
		utils.ValidateRequiredField("email", user.Email),
		utils.ValidateRequiredField("password", user.Password),
	)

	if len(missingFields) > 0 {
		return User{}, utils.NewValidationError(missingFields...)
	}

	// Create user in repository
	result, err := service.repo.CreateUser(user)
	if err != nil {
		return User{}, utils.NewDatabaseError(err)
	}

	// Return user without password
	return User{
		ID:        result.ID,
		FirstName: result.FirstName,
		LastName:  result.LastName,
		Email:     result.Email,
		CreatedAt: result.CreatedAt,
		UpdatedAt: result.UpdatedAt,
	}, nil
}

func (service *UserService) GetUserById(id int) (User, *utils.ErrorResponse) {
	result, err := service.repo.GetUserById(id)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return User{}, utils.NewErrorResponse(http.StatusNotFound, "User not found")
		}
		return User{}, utils.NewDatabaseError(err)
	}

	return User{
		ID:        result.ID,
		FirstName: result.FirstName,
		LastName:  result.LastName,
		Email:     result.Email,
		CreatedAt: result.CreatedAt,
		UpdatedAt: result.UpdatedAt,
	}, nil
}

func (service *UserService) GetUsers() ([]User, *utils.ErrorResponse) {
	result, err := service.repo.GetUsers()
	if err != nil {
		return nil, utils.NewDatabaseError(err)
	}

	users := make([]User, len(result))
	for i, user := range result {
		users[i] = User{
			ID:        user.ID,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Email:     user.Email,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		}
	}

	return users, nil
}
