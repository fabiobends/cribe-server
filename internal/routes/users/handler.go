package users

import (
	"net/http"
	"strconv"
	"strings"

	"cribeapp.com/cribe-server/internal/errors"
	"cribeapp.com/cribe-server/internal/utils"
)

type UserHandler struct {
	service UserService
}

func NewUserHandler(service *UserService) *UserHandler {
	return &UserHandler{service: *service}
}

func (handler *UserHandler) HandleRequest(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		handler.handlePost(w, r)
	case http.MethodGet:
		handler.handleGet(w, r)
	default:
		utils.NotAllowed(w)
	}
}

func (handler *UserHandler) handlePost(w http.ResponseWriter, r *http.Request) {
	userDTO, errResp := utils.DecodeBody[UserDTO](r)
	if errResp != nil {
		utils.EncodeResponse(w, http.StatusBadRequest, errResp)
		return
	}

	response, errResp := handler.service.CreateUser(userDTO)
	if errResp != nil {
		if errResp.Message == errors.ValidationError {
			utils.EncodeResponse(w, http.StatusBadRequest, errResp)
		} else {
			utils.EncodeResponse(w, http.StatusInternalServerError, errResp)
		}
		return
	}

	utils.EncodeResponse(w, http.StatusCreated, response)
}

func (handler *UserHandler) handleGet(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/users")
	path = strings.TrimPrefix(path, "/")

	if path == "" {
		response, errResp := handler.service.GetUsers()
		if errResp != nil {
			utils.EncodeResponse(w, http.StatusNotFound, errResp)
			return
		}
		utils.EncodeResponse(w, http.StatusOK, response)
		return
	}

	id, err := strconv.Atoi(path)
	if err != nil {
		statusCode := http.StatusBadRequest
		utils.EncodeResponse(w, statusCode, &errors.ErrorResponse{
			Message: errors.InvalidIdParameter,
			Details: err.Error(),
		})
		return
	}

	response, errResp := handler.service.GetUserById(id)
	if errResp != nil {
		utils.EncodeResponse(w, http.StatusNotFound, errResp)
		return
	}
	utils.EncodeResponse(w, http.StatusOK, response)
}
