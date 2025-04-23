package users

import (
	"reflect"
	"testing"

	"cribeapp.com/cribe-server/internal/utils"
)

func TestUserHandler_PostUser(t *testing.T) {
	t.Run("should create a user with valid input and return the user", func(t *testing.T) {
		user := User{
			ID:        1,
			FirstName: "John",
			LastName:  "Doe",
			Email:     "john.doe@example.com",
			Password:  "password123",
			CreatedAt: utils.MockGetCurrentTime(),
			UpdatedAt: utils.MockGetCurrentTime(),
		}

		service := GetNewMockUserService()
		handler := NewUserHandler(service)

		userDTO := UserDTO{
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Email:     user.Email,
			Password:  user.Password,
		}

		result, err := handler.PostUser(userDTO)
		if err != nil {
			t.Errorf("Error creating user: %v", err)
		}

		// They should not be equal because the password is not returned
		if reflect.DeepEqual(result, user) {
			t.Errorf("Expected user to be %v, got %v", user, result)
		}
	})
	t.Run("should not create a user with invalid input and return an error", func(t *testing.T) {
		service := GetNewMockUserService()
		handler := NewUserHandler(service)

		userDTO := UserDTO{
			FirstName: "John",
			LastName:  "Doe",
			Email:     "john.doe@example.com",
		}

		_, err := handler.PostUser(userDTO)
		if err != nil && err.Message == "" {
			t.Errorf("Expected error message, got %v", err)
		}

		if len(err.Details) == 0 {
			t.Errorf("Expected error details, got %v", err)
		}
	})
}
