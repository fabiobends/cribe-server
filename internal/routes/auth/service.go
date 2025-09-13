package auth

import (
	"strings"

	"cribeapp.com/cribe-server/internal/core/logger"
	"cribeapp.com/cribe-server/internal/errors"
	"cribeapp.com/cribe-server/internal/routes/users"
)

type AuthService struct {
	userRepo     *users.UserRepository
	tokenService TokenService
	logger       *logger.ContextualLogger
}

func NewAuthService(userRepo *users.UserRepository, tokenService TokenService) *AuthService {
	return &AuthService{
		userRepo:     userRepo,
		tokenService: tokenService,
		logger:       logger.NewServiceLogger("AuthService"),
	}
}

func (s *AuthService) Register(data users.UserDTO) (*RegisterResponse, *errors.ErrorResponse) {
	s.logger.Debug("Starting user registration process", map[string]interface{}{
		"email": data.Email, // Will be automatically masked
	})

	hashedPassword, err := s.tokenService.GenerateHash(data.Password)
	if err != nil {
		s.logger.Error("Failed to hash password", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, &errors.ErrorResponse{
			Message: errors.InternalServerError,
			Details: err.Error(),
		}
	}

	s.logger.Debug("Creating user in database")
	user, err := s.userRepo.CreateUser(users.UserDTO{
		FirstName: data.FirstName,
		LastName:  data.LastName,
		Email:     data.Email,
		Password:  string(hashedPassword),
	})
	if err != nil {
		s.logger.Error("Failed to create user in database", map[string]interface{}{
			"error": err.Error(),
			"email": data.Email, // Will be automatically masked
		})

		// Check if it's a duplicate email error
		if strings.Contains(err.Error(), "unique constraint") || strings.Contains(err.Error(), "duplicate key") {
			return nil, &errors.ErrorResponse{
				Message: errors.DatabaseError,
				Details: "Email address is already registered",
			}
		}

		return nil, &errors.ErrorResponse{
			Message: errors.DatabaseError,
			Details: "Failed to create user account",
		}
	}

	s.logger.Info("User registration completed successfully", map[string]interface{}{
		"userID": user.ID,
		"email":  data.Email, // Will be automatically masked
	})
	return &RegisterResponse{ID: user.ID}, nil
}

func (s *AuthService) Login(data LoginRequest) (*LoginResponse, *errors.ErrorResponse) {
	s.logger.Debug("Starting login process", map[string]interface{}{
		"email": data.Email, // Will be automatically masked
	})

	user, err := s.userRepo.GetUserByEmail(data.Email)
	if err != nil {
		s.logger.Warn("Login failed: user not found", map[string]interface{}{
			"email": data.Email, // Will be automatically masked
			"error": err.Error(),
		})
		return nil, &errors.ErrorResponse{
			Message: errors.InvalidCredentials,
			Details: "Invalid email or password",
		}
	}

	s.logger.Debug("Verifying password for user", map[string]interface{}{
		"userID": user.ID,
	})

	err = s.tokenService.CompareHashAndPassword(user.Password, data.Password)
	if err != nil {
		s.logger.Warn("Login failed: invalid password", map[string]interface{}{
			"userID": user.ID,
			"email":  data.Email, // Will be automatically masked
		})
		return nil, &errors.ErrorResponse{
			Message: errors.InvalidCredentials,
			Details: "Invalid email or password",
		}
	}

	s.logger.Debug("Generating access token", map[string]interface{}{
		"userID": user.ID,
	})

	accessToken, err := s.tokenService.GetAccessToken(user.ID)
	if err != nil {
		s.logger.Error("Failed to generate access token", map[string]interface{}{
			"userID": user.ID,
			"error":  err.Error(),
		})
		return nil, &errors.ErrorResponse{
			Message: errors.InternalServerError,
			Details: err.Error(),
		}
	}

	s.logger.Debug("Generating refresh token", map[string]interface{}{
		"userID": user.ID,
	})

	refreshToken, err := s.tokenService.GetRefreshToken(user.ID)
	if err != nil {
		s.logger.Error("Failed to generate refresh token", map[string]interface{}{
			"userID": user.ID,
			"error":  err.Error(),
		})
		return nil, &errors.ErrorResponse{
			Message: errors.InternalServerError,
			Details: err.Error(),
		}
	}

	s.logger.Info("Login completed successfully", map[string]interface{}{
		"userID": user.ID,
		"email":  data.Email, // Will be automatically masked
	})

	return &LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *AuthService) RefreshToken(data RefreshTokenRequest) (*RefreshTokenResponse, *errors.ErrorResponse) {
	s.logger.Debug("Starting token refresh process")

	user, err := s.tokenService.ValidateToken(data.RefreshToken)
	if err != nil {
		s.logger.Warn("Token refresh failed: invalid refresh token", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, &errors.ErrorResponse{
			Message: errors.InvalidRequestBody,
			Details: err.Error(),
		}
	}

	s.logger.Debug("Validating user exists", map[string]interface{}{
		"userID": user.UserID,
	})

	_, err = s.userRepo.GetUserById(user.UserID)
	if err != nil {
		s.logger.Error("Token refresh failed: user not found", map[string]interface{}{
			"userID": user.UserID,
			"error":  err.Error(),
		})
		return nil, &errors.ErrorResponse{
			Message: errors.DatabaseError,
			Details: err.Error(),
		}
	}

	s.logger.Debug("Generating new access token", map[string]interface{}{
		"userID": user.UserID,
	})

	accessToken, err := s.tokenService.GetAccessToken(user.UserID)
	if err != nil {
		s.logger.Error("Failed to generate new access token", map[string]interface{}{
			"userID": user.UserID,
			"error":  err.Error(),
		})
		return nil, &errors.ErrorResponse{
			Message: errors.InternalServerError,
			Details: err.Error(),
		}
	}

	s.logger.Info("Token refresh completed successfully", map[string]interface{}{
		"userID": user.UserID,
	})

	return &RefreshTokenResponse{
		AccessToken: accessToken,
	}, nil
}
