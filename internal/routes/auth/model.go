package auth

import (
	"cribeapp.com/cribe-server/internal/errors"
	"cribeapp.com/cribe-server/internal/utils"
)

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=1"`
}

// Validate performs validation on the login request
func (req LoginRequest) Validate() *errors.ErrorResponse {
	return utils.ValidateStruct(req)
}

type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type RegisterResponse struct {
	ID int `json:"id"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required,min=1"`
}

// Validate performs validation on the refresh token request
func (req RefreshTokenRequest) Validate() *errors.ErrorResponse {
	return utils.ValidateStruct(req)
}

type RefreshTokenResponse struct {
	AccessToken string `json:"access_token"`
}
