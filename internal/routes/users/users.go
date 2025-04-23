package users

import (
	"net/http"

	"cribeapp.com/cribe-server/internal/utils"
)

func Handler(w http.ResponseWriter, r *http.Request) {
	repo := *NewUserRepository()
	service := NewUserService(repo)
	handler := NewUserHandler(service)

	if r.Method == http.MethodPost {
		userDTO, _ := utils.DecodeBody[UserDTO](r)
		response, err := handler.PostUser(userDTO)
		if err != nil {
			utils.EncodeResponse(w, http.StatusBadRequest, err)
			return
		}
		utils.EncodeResponse(w, http.StatusCreated, response)
		return
	}
	utils.NotAllowed(w)
}
