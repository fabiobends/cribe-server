package users

import (
	"testing"

	"cribeapp.com/cribe-server/internal/utils"
)

type MockQueryExecutor struct{}

func QueryItem(query string, args ...any) (User, error) {
	return User{
		ID:        1,
		FirstName: args[0].(string),
		LastName:  args[1].(string),
		Email:     args[2].(string),
		Password:  args[3].(string),
		CreatedAt: utils.MockGetCurrentTime(),
		UpdatedAt: utils.MockGetCurrentTime(),
	}, nil
}

func TestUserRepository_CreateUser(t *testing.T) {
	repo := NewUserRepository(WithQueryExecutor(QueryExecutor{QueryItem: QueryItem}))

	user := User{
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john.doe@example.com",
		Password:  "password",
	}

	result, err := repo.CreateUser(user)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	expected := User{
		ID:        1,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Email:     user.Email,
		Password:  user.Password,
		CreatedAt: utils.MockGetCurrentTime(),
		UpdatedAt: utils.MockGetCurrentTime(),
	}

	if result != expected {
		t.Errorf("expected %+v, got %+v", expected, result)
	}
}
