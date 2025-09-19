package users

import (
	"cribeapp.com/cribe-server/internal/core/logger"
	"cribeapp.com/cribe-server/internal/utils"
)

type UserRepository struct {
	*utils.Repository[UserWithPassword]
	logger *logger.ContextualLogger
}

func NewUserRepository(options ...utils.Option[UserWithPassword]) *UserRepository {
	repo := utils.NewRepository(options...)
	return &UserRepository{
		Repository: repo,
		logger:     logger.NewRepositoryLogger("UserRepository"),
	}
}

func (r *UserRepository) CreateUser(user UserDTO) (UserWithPassword, error) {
	r.logger.Debug("Creating user in database", map[string]interface{}{
		"email": user.Email, // Will be automatically masked
	})

	query := `
		INSERT INTO users (first_name, last_name, email, password)
		VALUES ($1, $2, $3, $4)
		RETURNING *
	`

	result, err := r.Executor.QueryItem(query, user.FirstName, user.LastName, user.Email, user.Password)
	if err != nil {
		r.logger.Error("Failed to create user", map[string]interface{}{
			"email": user.Email, // Will be automatically masked
			"error": err.Error(),
		})
		return result, err
	}

	r.logger.Info("User created successfully", map[string]interface{}{
		"userID": result.ID,
		"email":  user.Email, // Will be automatically masked
	})

	return result, nil
}

func (r *UserRepository) GetUserById(id int) (UserWithPassword, error) {
	r.logger.Debug("Fetching user by ID", map[string]interface{}{
		"userID": id,
	})

	query := "SELECT * FROM users WHERE id = $1"

	result, err := r.Executor.QueryItem(query, id)
	if err != nil {
		r.logger.Error("Failed to fetch user by ID", map[string]interface{}{
			"userID": id,
			"error":  err.Error(),
		})
		return result, err
	}

	r.logger.Debug("User found by ID", map[string]interface{}{
		"userID": id,
	})

	return result, nil
}

func (r *UserRepository) GetUserByEmail(email string) (UserWithPassword, error) {
	r.logger.Debug("Fetching user by email", map[string]interface{}{
		"email": email, // Will be automatically masked
	})

	query := "SELECT * FROM users WHERE email = $1"

	result, err := r.Executor.QueryItem(query, email)
	if err != nil {
		r.logger.Error("Failed to fetch user by email", map[string]interface{}{
			"email": email, // Will be automatically masked
			"error": err.Error(),
		})
		return result, err
	}

	r.logger.Debug("User found by email", map[string]interface{}{
		"userID": result.ID,
		"email":  email, // Will be automatically masked
	})

	return result, nil
}

func (r *UserRepository) GetUsers() ([]UserWithPassword, error) {
	query := "SELECT * FROM users"

	return r.Executor.QueryList(query)
}
