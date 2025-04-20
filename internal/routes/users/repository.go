package users

import "time"

type QueryExecutor struct {
	QueryItem func(query string, args ...interface{}) (User, error)
}

type UserRepository struct {
	executor QueryExecutor
}

type User struct {
	ID        int       `json:"id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Email     string    `json:"email"`
	Password  string    `json:"password"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Option func(*UserRepository)

func WithQueryExecutor(executor QueryExecutor) Option {
	return func(r *UserRepository) {
		r.executor = executor
	}
}

func NewUserRepository(options ...Option) *UserRepository {
	repo := &UserRepository{}
	for _, option := range options {
		option(repo)
	}
	return repo
}

func (r *UserRepository) CreateUser(user User) (User, error) {
	query := `
		INSERT INTO users (first_name, last_name, email, password)
		VALUES ($1, $2, $3, $4)
		RETURNING *
	`

	return r.executor.QueryItem(query, user.FirstName, user.LastName, user.Email, user.Password)
}
