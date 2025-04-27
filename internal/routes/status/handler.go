package status

import (
	"net/http"

	"cribeapp.com/cribe-server/internal/utils"
)

type StatusHandler struct {
	service StatusServiceInterface
}

func NewStatusHandler(service StatusServiceInterface) *StatusHandler {
	return &StatusHandler{service: service}
}

func (handler *StatusHandler) HandleRequest(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		handler.handleGet(w, r)
	default:
		utils.NotAllowed(w)
	}
}

func (handler *StatusHandler) handleGet(w http.ResponseWriter, r *http.Request) {
	response := handler.service.GetStatus()
	utils.EncodeResponse(w, http.StatusOK, response)
}
