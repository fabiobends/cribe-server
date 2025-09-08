package auth

import (
	"log"
	"net/http"
	"slices"
	"strings"

	"cribeapp.com/cribe-server/internal/errors"
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
	paths := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	validPaths := []string{"refresh", "register", "login"}
	if len(paths) == 1 || len(paths) > 1 && !slices.Contains(validPaths, paths[1]) {
		log.Println("Invalid path", r.URL.Path)
		utils.NotFound(w, r)
		return
	}
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
			utils.EncodeResponse(w, http.StatusBadRequest, err)
			return
		}
		err = userDTO.Validate()
		if err != nil {
			utils.EncodeResponse(w, http.StatusBadRequest, err)
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
			utils.EncodeResponse(w, http.StatusInternalServerError, err)
			return
		}
		response, err := handler.service.Login(loginRequest)
		if err != nil {
			utils.EncodeResponse(w, http.StatusBadRequest, err)
			return
		}
		utils.EncodeResponse(w, http.StatusOK, response)
		return
	}

	if path == "refresh" {
		refreshRequest, err := utils.DecodeBody[RefreshTokenRequest](r)
		if err != nil {
			if err.Message == errors.InvalidRequestBody {
				utils.EncodeResponse(w, http.StatusBadRequest, err)
				return
			}
			utils.EncodeResponse(w, http.StatusInternalServerError, err)
			return
		}
		response, err := handler.service.RefreshToken(refreshRequest)
		if err != nil {
			utils.EncodeResponse(w, http.StatusUnauthorized, err)
			return
		}
		utils.EncodeResponse(w, http.StatusOK, response)
		return
	}

	utils.EncodeResponse(w, http.StatusNotFound, &errors.ErrorResponse{
		Message: errors.RouteNotFound,
		Details: "Invalid path",
	})
}
