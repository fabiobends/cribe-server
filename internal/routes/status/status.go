package status

import (
	"net/http"

	"cribeapp.com/cribe-server/internal/utils"
)

func Handler(w http.ResponseWriter, r *http.Request) {
	repo := &StatusRepository{}
	service := NewStatusService(repo)
	handler := NewStatusHandler(service)

	if r.Method == http.MethodGet {
		response := handler.GetStatus()
		utils.EncodeResponse(w, http.StatusOK, response)
		return
	}
	utils.NotAllowed(w)
}
