package users

import (
	"reflect"
	"testing"

	"cribeapp.com/cribe-server/internal/utils"
)

// MockQueryExecutor is a mock implementation of QueryExecutor
type MockQueryExecutor struct {
	QueryItemFunc func(query string, args ...any) (User, error)
}

func (m *MockQueryExecutor) QueryItem(query string, args ...any) (User, error) {
	return m.QueryItemFunc(query, args...)
}

func GetNewMockRepository() *UserRepository {
	mockExecutor := &MockQueryExecutor{
		QueryItemFunc: func(query string, args ...any) (User, error) {
			return User{
				ID:        1,
				FirstName: "John",
				LastName:  "Doe",
				Email:     "john.doe@example.com",
				Password:  "password123",
				CreatedAt: utils.MockGetCurrentTime(),
				UpdatedAt: utils.MockGetCurrentTime(),
			}, nil
		},
	}

	return NewUserRepository(utils.WithQueryExecutor(utils.QueryExecutor[User]{
		QueryItem: mockExecutor.QueryItem,
	}))
}

func TestUserRepository_CreateUser(t *testing.T) {
	expected := User{
		ID:        1,
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john.doe@example.com",
		Password:  "password123",
		CreatedAt: utils.MockGetCurrentTime(),
		UpdatedAt: utils.MockGetCurrentTime(),
	}

	repo := GetNewMockRepository()

	userDTO := UserDTO{
		FirstName: expected.FirstName,
		LastName:  expected.LastName,
		Email:     expected.Email,
		Password:  expected.Password,
	}

	result, err := repo.CreateUser(userDTO)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if !reflect.DeepEqual(expected, result) {
		t.Errorf("Expected %+v, got %+v", expected, result)
	}
}
