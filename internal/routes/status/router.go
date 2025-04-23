package status

import (
	"net/http"
	"time"

	"cribeapp.com/cribe-server/internal/utils"
)

func HandleHTTPRequests(w http.ResponseWriter, r *http.Request) {
	repo := *NewStatusRepository()
	service := NewStatusService(repo, time.Now)
	handler := NewStatusHandler(service)

	if r.Method == http.MethodGet {
		response := handler.GetStatus()
		utils.EncodeResponse(w, http.StatusOK, response)
		return
	}
	utils.NotAllowed(w)
}
