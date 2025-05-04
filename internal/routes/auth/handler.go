package auth

import (
	"net/http"
	"strings"

	"cribeapp.com/cribe-server/internal/routes/users"
	"cribeapp.com/cribe-server/internal/utils"
)

type AuthHandler struct {
	service AuthService
}

func NewAuthHandler(service *AuthService) *AuthHandler {
	return &AuthHandler{service: *service}
}

func (handler *AuthHandler) HandleRequest(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		handler.handlePost(w, r)
	default:
		utils.NotAllowed(w)
	}
}

func (handler *AuthHandler) handlePost(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/auth")
	path = strings.TrimPrefix(path, "/")

	if path == "register" {
		userDTO, err := utils.DecodeBody[users.UserDTO](r)
		if err != nil {
			utils.EncodeResponse(w, err.StatusCode, err)
			return
		}
		err = userDTO.Validate()
		if err != nil {
			utils.EncodeResponse(w, err.StatusCode, err)
			return
		}
		response, err := handler.service.Register(userDTO)
		if err != nil {
			utils.EncodeResponse(w, http.StatusInternalServerError, err)
			return
		}
		utils.EncodeResponse(w, http.StatusCreated, response)
		return
	}

	if path == "login" {
		loginRequest, err := utils.DecodeBody[LoginRequest](r)
		if err != nil {
			utils.EncodeResponse(w, err.StatusCode, err)
			return
		}
		response, err := handler.service.Login(loginRequest)
		if err != nil {
			utils.EncodeResponse(w, err.StatusCode, err)
			return
		}
		utils.EncodeResponse(w, http.StatusOK, response)
		return
	}

	if path == "refresh" {
		refreshRequest, err := utils.DecodeBody[RefreshTokenRequest](r)
		if err != nil {
			utils.EncodeResponse(w, err.StatusCode, err)
			return
		}
		response, err := handler.service.RefreshToken(refreshRequest)
		if err != nil {
			utils.EncodeResponse(w, err.StatusCode, err)
			return
		}
		utils.EncodeResponse(w, http.StatusOK, response)
		return
	}

	utils.EncodeResponse(w, http.StatusNotFound, utils.NewErrorResponse(http.StatusNotFound, "Route not found", "Invalid path"))
}
