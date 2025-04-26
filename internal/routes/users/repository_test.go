package users

import (
	"fmt"
	"reflect"
	"testing"

	"cribeapp.com/cribe-server/internal/utils"
)

// MockQueryExecutor is a mock implementation of QueryExecutor
type MockQueryExecutor struct {
	QueryItemFunc func(query string, args ...any) (UserWithPassword, error)
	QueryListFunc func(query string, args ...any) ([]UserWithPassword, error)
}

func (m *MockQueryExecutor) QueryItem(query string, args ...any) (UserWithPassword, error) {
	return m.QueryItemFunc(query, args...)
}

func (m *MockQueryExecutor) QueryList(query string, args ...any) ([]UserWithPassword, error) {
	return m.QueryListFunc(query, args...)
}

var users = []UserWithPassword{
	{
		ID:        1,
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john.doe@example.com",
		Password:  "password123",
		CreatedAt: utils.MockGetCurrentTime(),
		UpdatedAt: utils.MockGetCurrentTime(),
	},
	{
		ID:        2,
		FirstName: "Jane",
		LastName:  "Doe",
		Email:     "jane.doe@example.com",
		Password:  "password123",
		CreatedAt: utils.MockGetCurrentTime(),
		UpdatedAt: utils.MockGetCurrentTime(),
	},
}

func GetNewMockRepository() *UserRepository {
	mockExecutor := &MockQueryExecutor{
		QueryItemFunc: func(query string, args ...any) (UserWithPassword, error) {
			if query == "SELECT * FROM users WHERE id = $1" {
				id := args[0].(int)
				for _, user := range users {
					if user.ID == id {
						return user, nil
					}
				}
				return UserWithPassword{}, fmt.Errorf("user not found")
			}

			neededArgsLength := 4
			if len(args) < neededArgsLength {
				return UserWithPassword{}, fmt.Errorf("expected %d arguments, got %d", neededArgsLength, len(args))
			}

			// Check if any field is empty
			for _, arg := range args {
				if arg == "" {
					return UserWithPassword{}, fmt.Errorf("empty field")
				}
			}

			return users[0], nil
		},
		QueryListFunc: func(query string, args ...any) ([]UserWithPassword, error) {
			return users, nil
		},
	}

	return NewUserRepository(utils.WithQueryExecutor(utils.QueryExecutor[UserWithPassword]{
		QueryItem: mockExecutor.QueryItem,
		QueryList: mockExecutor.QueryList,
	}))
}

func TestUserRepository_CreateUser(t *testing.T) {
	t.Run("should create a user with valid input", func(t *testing.T) {
		repo := GetNewMockRepository()

		firstUser := users[0]
		userDTO := UserDTO{
			FirstName: firstUser.FirstName,
			LastName:  firstUser.LastName,
			Email:     firstUser.Email,
			Password:  firstUser.Password,
		}

		result, err := repo.CreateUser(userDTO)

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if !reflect.DeepEqual(firstUser, result) {
			t.Errorf("Expected %+v, got %+v", firstUser, result)
		}
	})

	t.Run("should not create a user with invalid input", func(t *testing.T) {
		repo := GetNewMockRepository()
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
		repo := GetNewMockRepository()
		firstUser := users[0]
		user, err := repo.GetUserById(firstUser.ID)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if !reflect.DeepEqual(firstUser, user) {
			t.Errorf("Expected %+v, got %+v", firstUser, user)
		}
	})

	t.Run("should return an error if the user is not found", func(t *testing.T) {
		repo := GetNewMockRepository()
		user, err := repo.GetUserById(0)
		if err == nil {
			t.Errorf("Expected error, got nil")
		}

		if user != (UserWithPassword{}) {
			t.Errorf("Expected empty user, got %+v", user)
		}
	})
}

func TestUserRepository_GetUsers(t *testing.T) {
	t.Run("should get all users", func(t *testing.T) {
		repo := GetNewMockRepository()
		result, err := repo.GetUsers()
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if len(users) != len(result) {
			t.Errorf("Expected %d users, got %d", len(users), len(result))
		}

		if !reflect.DeepEqual(users[0], result[0]) {
			t.Errorf("Expected %+v, got %+v", users[0], result[0])
		}

		if !reflect.DeepEqual(users[1], result[1]) {
			t.Errorf("Expected %+v, got %+v", users[1], result[1])
		}
	})
}
