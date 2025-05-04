package auth

import (
	"errors"
	"strings"
	"time"

	"cribeapp.com/cribe-server/internal/routes/users"
	"cribeapp.com/cribe-server/internal/utils"
)

type MockTokenService struct {
	secretKey              []byte
	accessTokenExpiration  time.Duration
	refreshTokenExpiration time.Duration
	currentTime            func() time.Time
}

func NewMockAuthServiceReady(presetUsers ...users.UserWithPassword) *AuthService {
	repo := users.NewMockUserRepositoryReady(presetUsers...)
	tokenService := NewMockTokenService([]byte("test"), time.Hour, time.Hour*24*30, utils.MockGetCurrentTime)
	service := NewAuthService(repo, tokenService)
	return service
}

func NewMockAuthHandlerReady() *AuthHandler {
	service := NewMockAuthServiceReady()
	return NewAuthHandler(service)
}

func NewMockTokenService(secretKey []byte, accessTokenExpiration, refreshTokenExpiration time.Duration, currentTime func() time.Time) *MockTokenService {
	return &MockTokenService{
		secretKey:              secretKey,
		accessTokenExpiration:  accessTokenExpiration,
		refreshTokenExpiration: refreshTokenExpiration,
		currentTime:            currentTime,
	}
}

func (s *MockTokenService) GetAccessToken(userID int) (string, error) {
	return "access_token" + "_" + string(s.secretKey), nil
}

func (s *MockTokenService) GetRefreshToken(userID int) (string, error) {
	return "refresh_token" + "_" + string(s.secretKey), nil
}

func (s *MockTokenService) GenerateHash(text string) (string, error) {
	return "hashed_password" + "_" + string(s.secretKey), nil
}

func (s *MockTokenService) CompareHashAndPassword(hashedPassword, password string) error {
	if strings.Contains(hashedPassword, "hashed_password") {
		return nil
	}
	return errors.New("Invalid password")
}

func (s *MockTokenService) ValidateToken(token string) (*JWTObject, error) {
	if token != "access_token"+"_"+string(s.secretKey) && token != "refresh_token"+"_"+string(s.secretKey) {
		return nil, errors.New("Invalid token")
	}
	if s.currentTime().Add(s.accessTokenExpiration).Before(s.currentTime()) {
		return nil, errors.New("Access token expired")
	}
	if s.currentTime().Add(s.refreshTokenExpiration).Before(s.currentTime()) {
		return nil, errors.New("Refresh token expired")
	}
	return &JWTObject{
		UserID: 1,
		Exp:    s.currentTime().Add(s.accessTokenExpiration).Unix(),
		Typ:    "access",
	}, nil
}
