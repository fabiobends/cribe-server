package auth

import (
	"cribeapp.com/cribe-server/internal/routes/users"
	"cribeapp.com/cribe-server/internal/utils"
)

type AuthService struct {
	userRepo     *users.UserRepository
	tokenService TokenService
}

func NewAuthService(userRepo *users.UserRepository, tokenService TokenService) *AuthService {
	return &AuthService{
		userRepo:     userRepo,
		tokenService: tokenService,
	}
}

func (s *AuthService) Register(data users.UserDTO) (*RegisterResponse, *utils.ErrorResponse) {
	hashedPassword, err := s.tokenService.GenerateHash(data.Password)
	if err != nil {
		return nil, &utils.ErrorResponse{
			Message: utils.InternalError,
			Details: err.Error(),
		}
	}
	user, err := s.userRepo.CreateUser(users.UserDTO{
		FirstName: data.FirstName,
		LastName:  data.LastName,
		Email:     data.Email,
		Password:  string(hashedPassword),
	})
	if err != nil {
		return nil, &utils.ErrorResponse{
			Message: utils.DatabaseError,
			Details: err.Error(),
		}
	}
	return &RegisterResponse{ID: user.ID}, nil
}

func (s *AuthService) Login(data LoginRequest) (*LoginResponse, *utils.ErrorResponse) {
	user, err := s.userRepo.GetUserByEmail(data.Email)
	if err != nil {
		return nil, &utils.ErrorResponse{
			Message: utils.InvalidCredentials,
			Details: err.Error(),
		}
	}

	err = s.tokenService.CompareHashAndPassword(user.Password, data.Password)
	if err != nil {
		return nil, &utils.ErrorResponse{
			Message: utils.InvalidCredentials,
			Details: err.Error(),
		}
	}

	accessToken, err := s.tokenService.GetAccessToken(user.ID)
	if err != nil {
		return nil, &utils.ErrorResponse{
			Message: utils.InternalError,
			Details: err.Error(),
		}
	}

	refreshToken, err := s.tokenService.GetRefreshToken(user.ID)
	if err != nil {
		return nil, &utils.ErrorResponse{
			Message: utils.InternalError,
			Details: err.Error(),
		}
	}

	return &LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *AuthService) RefreshToken(data RefreshTokenRequest) (*RefreshTokenResponse, *utils.ErrorResponse) {
	user, err := s.tokenService.ValidateToken(data.RefreshToken)
	if err != nil {
		return nil, &utils.ErrorResponse{
			Message: utils.InvalidRequestBody,
			Details: err.Error(),
		}
	}

	_, err = s.userRepo.GetUserById(user.UserID)
	if err != nil {
		return nil, &utils.ErrorResponse{
			Message: utils.DatabaseError,
			Details: err.Error(),
		}
	}

	accessToken, err := s.tokenService.GetAccessToken(user.UserID)
	if err != nil {
		return nil, &utils.ErrorResponse{
			Message: utils.InternalError,
			Details: err.Error(),
		}
	}

	return &RefreshTokenResponse{
		AccessToken: accessToken,
	}, nil
}
