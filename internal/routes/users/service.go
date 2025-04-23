package users

import "cribeapp.com/cribe-server/internal/utils"

type UserServiceInterface interface {
	PostUser(user UserDTO) (User, *utils.ErrorResponse)
}

type UserService struct {
	repo UserRepository
}

func NewUserService(repo UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (service *UserService) PostUser(user UserDTO) (User, *utils.ErrorResponse) {
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
