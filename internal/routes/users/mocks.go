package users

import (
	"fmt"
	"strings"

	"cribeapp.com/cribe-server/internal/utils"
)

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

func NewMockUserRepositoryReady(presetUsers ...UserWithPassword) *UserRepository {
	// Create a local copy of the users to avoid modifying the original slice
	users := make([]UserWithPassword, len(presetUsers))
	copy(users, presetUsers)

	mockExecutor := &MockQueryExecutor{
		QueryItemFunc: func(query string, args ...any) (UserWithPassword, error) {
			// Check if any field is empty
			for _, arg := range args {
				if arg == "" {
					return UserWithPassword{}, fmt.Errorf("Empty field")
				}
			}
			// Get user by id
			if query == "SELECT * FROM users WHERE id = $1" {
				id := args[0].(int)
				for _, user := range users {
					if user.ID == id {
						return user, nil
					}
				}
				return UserWithPassword{}, fmt.Errorf("User not found")
			}

			// Get user by email
			if query == "SELECT * FROM users WHERE email = $1" {
				for _, user := range users {
					if user.Email == args[0].(string) {
						return user, nil
					}
				}
				return UserWithPassword{}, fmt.Errorf("User not found")
			}

			// Insert user
			if strings.Contains(query, "INSERT INTO users") {
				neededArgsLength := 4
				if len(args) < neededArgsLength {
					return UserWithPassword{}, fmt.Errorf("Expected %d arguments, got %d", neededArgsLength, len(args))
				}

				// Check if email is already taken, don't insert the user
				for _, user := range users {
					if user.Email == args[2].(string) {
						return UserWithPassword{}, fmt.Errorf("Email already taken")
					}
				}

				user := UserWithPassword{
					ID:        len(users) + 1,
					FirstName: args[0].(string),
					LastName:  args[1].(string),
					Email:     args[2].(string),
					Password:  args[3].(string),
					CreatedAt: utils.MockGetCurrentTime(),
					UpdatedAt: utils.MockGetCurrentTime(),
				}

				users = append(users, user)
				return user, nil
			}

			return UserWithPassword{}, fmt.Errorf("User not found")
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

func NewMockUserServiceReady() *UserService {
	repository := NewMockUserRepositoryReady()
	return NewUserService(*repository)
}

func NewMockUserHandlerReady() *UserHandler {
	service := NewMockUserServiceReady()
	return NewUserHandler(service)
}
