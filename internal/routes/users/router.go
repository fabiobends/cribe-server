package users

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	"cribeapp.com/cribe-server/internal/utils"
)

func HandleHTTPRequests(w http.ResponseWriter, r *http.Request) {
	repo := *NewUserRepository()
	service := NewUserService(repo)
	handler := NewUserHandler(service)

	if r.Method == http.MethodPost {
		userDTO, errResp := utils.DecodeBody[UserDTO](r)
		if errResp != nil {
			utils.EncodeResponse(w, http.StatusBadRequest, errResp)
			return
		}
		response, errResp := handler.PostUser(userDTO)
		if errResp != nil {
			utils.EncodeResponse(w, errResp.StatusCode, errResp)
			return
		}
		utils.EncodeResponse(w, http.StatusCreated, response)
		return
	}

	if r.Method == http.MethodGet {
		// Get route param in get request if it exists
		path := strings.TrimPrefix(r.URL.Path, "/users")
		path = strings.TrimPrefix(path, "/")
		if path == "" {
			response, errResp := handler.GetUsers()
			if errResp != nil {
				utils.EncodeResponse(w, errResp.StatusCode, errResp)
				return
			}
			utils.EncodeResponse(w, http.StatusOK, response)
			return
		}

		id, err := strconv.Atoi(path)
		log.Println("ID", id)
		if err != nil {
			log.Println("Error", err)
			statusCode := http.StatusBadRequest
			utils.EncodeResponse(w, statusCode, utils.NewErrorResponse(statusCode, "Invalid id parameter", err.Error()))
			return
		}

		response, errResp := handler.GetUserById(id)
		if errResp != nil {
			utils.EncodeResponse(w, errResp.StatusCode, errResp)
			return
		}
		utils.EncodeResponse(w, http.StatusOK, response)
		return
	}

	utils.NotAllowed(w)
}
