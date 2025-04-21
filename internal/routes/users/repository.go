package users

import (
	"time"

	"cribeapp.com/cribe-server/internal/utils"
)

type User struct {
	ID        int       `json:"id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Email     string    `json:"email"`
	Password  string    `json:"password"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type UserDTO struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Password  string `json:"password"`
}

type UserWithoutPassword struct {
	ID        int       `json:"id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type UserRepository struct {
	*utils.Repository[User]
}

func NewUserRepository(options ...utils.Option[User]) *UserRepository {
	repo := utils.NewRepository(options...)
	return &UserRepository{Repository: repo}
}

func (r *UserRepository) CreateUser(user UserDTO) (User, error) {
	query := `
		INSERT INTO users (first_name, last_name, email, password)
		VALUES ($1, $2, $3, $4)
		RETURNING *
	`

	return r.Repository.Executor.QueryItem(query, user.FirstName, user.LastName, user.Email, user.Password)
}
