package auth

import (
	"net/http"
	"slices"
	"strings"

	"cribeapp.com/cribe-server/internal/core/logger"
	"cribeapp.com/cribe-server/internal/errors"
	"cribeapp.com/cribe-server/internal/routes/users"
	"cribeapp.com/cribe-server/internal/utils"
)

type AuthHandler struct {
	service AuthService
	logger  *logger.ContextualLogger
}

func NewAuthHandler(service *AuthService) *AuthHandler {
	return &AuthHandler{
		service: *service,
		logger:  logger.NewHandlerLogger("AuthHandler"),
	}
}

func (handler *AuthHandler) HandleRequest(w http.ResponseWriter, r *http.Request) {
	handler.logger.Debug("Processing auth request", map[string]interface{}{
		"method": r.Method,
		"path":   r.URL.Path,
		"ip":     r.RemoteAddr,
	})

	paths := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	validPaths := []string{"refresh", "register", "login"}
	if len(paths) == 1 || len(paths) > 1 && !slices.Contains(validPaths, paths[1]) {
		handler.logger.Warn("Invalid auth path requested", map[string]interface{}{
			"path":       r.URL.Path,
			"validPaths": validPaths,
		})
		utils.NotFound(w, r)
		return
	}
	switch r.Method {
	case http.MethodPost:
		handler.handlePost(w, r)
	default:
		handler.logger.Warn("Method not allowed", map[string]interface{}{
			"method": r.Method,
			"path":   r.URL.Path,
		})
		utils.NotAllowed(w)
	}
}

func (handler *AuthHandler) handlePost(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/auth")
	path = strings.TrimPrefix(path, "/")

	handler.logger.Debug("Processing auth POST request", map[string]interface{}{
		"endpoint": path,
	})

	if path == "register" {
		handler.logger.Info("Processing user registration request")
		userDTO, err := utils.DecodeBody[users.UserDTO](r)
		if err != nil {
			handler.logger.Error("Failed to decode registration request body", map[string]interface{}{
				"error": err.Details,
			})
			utils.EncodeResponse(w, http.StatusBadRequest, err)
			return
		}
		err = userDTO.Validate()
		if err != nil {
			handler.logger.Warn("Registration validation failed", map[string]interface{}{
				"error": err.Details,
			})
			utils.EncodeResponse(w, http.StatusBadRequest, err)
			return
		}
		response, err := handler.service.Register(userDTO)
		if err != nil {
			handler.logger.Error("Registration failed", map[string]interface{}{
				"error": err.Details,
			})
			utils.EncodeResponse(w, http.StatusInternalServerError, err)
			return
		}
		handler.logger.Info("User registration successful", map[string]interface{}{
			"userID": response.ID,
		})
		utils.EncodeResponse(w, http.StatusCreated, response)
		return
	}

	if path == "login" {
		handler.logger.Info("Processing login request")
		loginRequest, err := utils.DecodeBody[LoginRequest](r)
		if err != nil {
			handler.logger.Error("Failed to decode login request body", map[string]interface{}{
				"error": err.Details,
			})
			utils.EncodeResponse(w, http.StatusBadRequest, err)
			return
		}
		err = loginRequest.Validate()
		if err != nil {
			handler.logger.Warn("Login validation failed", map[string]interface{}{
				"error": err.Details,
			})
			utils.EncodeResponse(w, http.StatusBadRequest, err)
			return
		}
		response, err := handler.service.Login(loginRequest)
		if err != nil {
			handler.logger.Error("Login failed", map[string]interface{}{
				"error": err.Details,
				"email": loginRequest.Email, // Will be automatically masked
			})
			utils.EncodeResponse(w, http.StatusBadRequest, err)
			return
		}
		handler.logger.Info("Login successful", map[string]interface{}{
			"email": loginRequest.Email, // Will be automatically masked
		})
		utils.EncodeResponse(w, http.StatusOK, response)
		return
	}

	if path == "refresh" {
		handler.logger.Info("Processing token refresh request")
		refreshRequest, err := utils.DecodeBody[RefreshTokenRequest](r)
		if err != nil {
			if err.Message == errors.InvalidRequestBody {
				handler.logger.Warn("Invalid refresh token request body", map[string]interface{}{
					"error": err.Details,
				})
				utils.EncodeResponse(w, http.StatusBadRequest, err)
				return
			}
			handler.logger.Error("Failed to decode refresh token request", map[string]interface{}{
				"error": err.Details,
			})
			utils.EncodeResponse(w, http.StatusInternalServerError, err)
			return
		}
		err = refreshRequest.Validate()
		if err != nil {
			handler.logger.Warn("Refresh token validation failed", map[string]interface{}{
				"error": err.Details,
			})
			utils.EncodeResponse(w, http.StatusBadRequest, err)
			return
		}
		response, err := handler.service.RefreshToken(refreshRequest)
		if err != nil {
			handler.logger.Error("Token refresh failed", map[string]interface{}{
				"error": err.Details,
			})
			utils.EncodeResponse(w, http.StatusUnauthorized, err)
			return
		}
		handler.logger.Info("Token refresh successful")
		utils.EncodeResponse(w, http.StatusOK, response)
		return
	}

	utils.EncodeResponse(w, http.StatusNotFound, &errors.ErrorResponse{
		Message: errors.RouteNotFound,
		Details: "Invalid path",
	})
}
