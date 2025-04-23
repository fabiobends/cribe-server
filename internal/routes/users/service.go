package users

import "cribeapp.com/cribe-server/internal/utils"

type UserServiceInterface interface {
	PostUser(user UserDTO) (UserWithoutPassword, *utils.ErrorResponse)
}

type UserService struct {
	repo UserRepository
}

func NewUserService(repo UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (service *UserService) PostUser(user UserDTO) (UserWithoutPassword, *utils.ErrorResponse) {
	// Validate required fields
	missingFields := utils.ValidateRequiredFields(
		utils.ValidateRequiredField("first_name", user.FirstName),
		utils.ValidateRequiredField("last_name", user.LastName),
		utils.ValidateRequiredField("email", user.Email),
		utils.ValidateRequiredField("password", user.Password),
	)

	if len(missingFields) > 0 {
		return UserWithoutPassword{}, utils.NewValidationError(missingFields...)
	}

	// Create user in repository
	result, err := service.repo.CreateUser(user)
	if err != nil {
		return UserWithoutPassword{}, utils.NewDatabaseError(err)
	}

	// Return user without password
	return UserWithoutPassword{
		ID:        result.ID,
		FirstName: result.FirstName,
		LastName:  result.LastName,
		Email:     result.Email,
		CreatedAt: result.CreatedAt,
		UpdatedAt: result.UpdatedAt,
	}, nil
}
