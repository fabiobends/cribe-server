package users

import (
	"reflect"
	"testing"

	"cribeapp.com/cribe-server/internal/utils"
)

func TestUserHandler_PostUser(t *testing.T) {
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
}
