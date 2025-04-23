package users

import (
	"cribeapp.com/cribe-server/internal/utils"
)

type UserRepository struct {
	*utils.Repository[UserWithPassword]
}

func NewUserRepository(options ...utils.Option[UserWithPassword]) *UserRepository {
	repo := utils.NewRepository(options...)
	return &UserRepository{Repository: repo}
}

func (r *UserRepository) CreateUser(user UserDTO) (UserWithPassword, error) {
	query := `
		INSERT INTO users (first_name, last_name, email, password)
		VALUES ($1, $2, $3, $4)
		RETURNING *
	`

	return r.Repository.Executor.QueryItem(query, user.FirstName, user.LastName, user.Email, user.Password)
}
