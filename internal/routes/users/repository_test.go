package users

import (
	"testing"
)

const userTestEmail = "john.doe.user.repository@example.com"

var repo = NewMockUserRepositoryReady()

func TestUserRepository_CreateUser(t *testing.T) {
	t.Run("should create a user with valid input", func(t *testing.T) {
		userDTO := UserDTO{
			FirstName: "John",
			LastName:  "Doe",
			Email:     userTestEmail,
			Password:  "password123",
		}

		result, err := repo.CreateUser(userDTO)

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if result.ID == 0 {
			t.Errorf("Expected user ID, got 0")
		}

		if result.FirstName != userDTO.FirstName {
			t.Errorf("Expected user first name to be %v, got %v", userDTO.FirstName, result.FirstName)
		}

		if result.LastName != userDTO.LastName {
			t.Errorf("Expected user last name to be %v, got %v", userDTO.LastName, result.LastName)
		}

		if result.Email != userDTO.Email {
			t.Errorf("Expected user email to be %v, got %v", userDTO.Email, result.Email)
		}
	})

	t.Run("should not create a user with invalid input", func(t *testing.T) {
		repo := NewMockUserRepositoryReady()
		userDTO := UserDTO{
			FirstName: "John",
			LastName:  "Doe",
			Email:     "john.doe@example.com",
		}

		_, err := repo.CreateUser(userDTO)
		if err == nil {
			t.Errorf("Expected error, got nil")
		}
	})
}

func TestUserRepository_GetUserById(t *testing.T) {
	t.Run("should get a user by id", func(t *testing.T) {
		user, err := repo.GetUserById(1)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if user.ID != 1 {
			t.Errorf("Expected user ID to be 1, got %v", user.ID)
		}

		if user.FirstName != "John" {
			t.Errorf("Expected user first name to be John, got %v", user.FirstName)
		}

		if user.LastName != "Doe" {
			t.Errorf("Expected user last name to be Doe, got %v", user.LastName)
		}

		if user.Email != "john.doe.user.repository@example.com" {
			t.Errorf("Expected user email to be john.doe.user.repository@example.com, got %v", user.Email)
		}
	})

	t.Run("should return an error if the user is not found", func(t *testing.T) {
		user, err := repo.GetUserById(0)
		if err == nil {
			t.Errorf("Expected error, got nil")
		}

		if user != (UserWithPassword{}) {
			t.Errorf("Expected empty user, got %+v", user)
		}
	})
}

func TestUserRepository_GetUserByEmail(t *testing.T) {
	t.Run("should get a user by email", func(t *testing.T) {
		user, err := repo.GetUserByEmail(userTestEmail)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if user.ID != 1 {
			t.Errorf("Expected user ID to be 1, got %v", user.ID)
		}

		if user.FirstName != "John" {
			t.Errorf("Expected user first name to be John, got %v", user.FirstName)
		}

		if user.LastName != "Doe" {
			t.Errorf("Expected user last name to be Doe, got %v", user.LastName)
		}

		if user.Email != userTestEmail {
			t.Errorf("Expected user email to be %v, got %v", userTestEmail, user.Email)
		}
	})
}

func TestUserRepository_GetUsers(t *testing.T) {
	t.Run("should get all users", func(t *testing.T) {
		presetUsers := []UserWithPassword{
			{
				ID:        1,
				FirstName: "John",
				LastName:  "Doe",
				Email:     userTestEmail,
			},
			{
				ID:        2,
				FirstName: "Jane",
				LastName:  "Doe",
				Email:     "jane.doe@example.com",
			},
		}
		repo := NewMockUserRepositoryReady(presetUsers...)
		result, err := repo.GetUsers()
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if len(result) == 0 {
			t.Errorf("Expected at least 1 user, got %d", len(result))
		}

		if result[0].ID != 1 {
			t.Errorf("Expected user ID to be 1, got %v", result[0].ID)
		}

		if result[1].ID != 2 {
			t.Errorf("Expected user ID to be 2, got %v", result[1].ID)
		}
	})
}
