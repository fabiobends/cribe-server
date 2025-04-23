package users

import (
	"reflect"
	"testing"

	"cribeapp.com/cribe-server/internal/utils"
)

func GetNewMockUserService() *UserService {
	repository := GetNewMockRepository()
	return NewUserService(*repository)
}

func TestUserService_CreateUser(t *testing.T) {
	t.Run("should create a user with valid input and return the user", func(t *testing.T) {
		user := UserWithPassword{
			ID:        1,
			FirstName: "John",
			LastName:  "Doe",
			Email:     "john.doe@example.com",
			Password:  "password123",
			CreatedAt: utils.MockGetCurrentTime(),
			UpdatedAt: utils.MockGetCurrentTime(),
		}

		service := GetNewMockUserService()

		userDTO := UserDTO{
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Email:     user.Email,
			Password:  user.Password,
		}

		result, err := service.PostUser(userDTO)
		if err != nil {
			t.Errorf("Error creating user: %v", err)
		}

		// They should not be equal because the password is not returned
		if reflect.DeepEqual(result, user) {
			t.Errorf("Expected user to be %v, got %v", user, result)
		}

		if result.ID != user.ID {
			t.Errorf("Expected user ID to be %v, got %v", user.ID, result.ID)
		}

		if result.FirstName != user.FirstName {
			t.Errorf("Expected user first name to be %v, got %v", user.FirstName, result.FirstName)
		}

		if result.LastName != user.LastName {
			t.Errorf("Expected user last name to be %v, got %v", user.LastName, result.LastName)
		}

		if result.Email != user.Email {
			t.Errorf("Expected user email to be %v, got %v", user.Email, result.Email)
		}

		if result.CreatedAt != user.CreatedAt {
			t.Errorf("Expected user created at to be %v, got %v", user.CreatedAt, result.CreatedAt)
		}

		if result.UpdatedAt != user.UpdatedAt {
			t.Errorf("Expected user updated at to be %v, got %v", user.UpdatedAt, result.UpdatedAt)
		}
	})

	t.Run("should not create a user with invalid input", func(t *testing.T) {
		service := GetNewMockUserService()
		userDTO := UserDTO{
			FirstName: "John",
			LastName:  "Doe",
		}

		_, err := service.PostUser(userDTO)
		if err == nil {
			t.Errorf("Expected error, got nil")
		}
	})
}
