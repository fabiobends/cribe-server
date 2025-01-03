package status

import (
	"cribeapp.com/cribe-server/internal/utils"
)

type StatusHandler struct {
	service StatusServiceInterface
}

func NewStatusHandler(service StatusServiceInterface) *StatusHandler {
	return &StatusHandler{service: service}
}

func (handler *StatusHandler) GetStatus() utils.StandardResponse {
	return utils.StandardResponse{Message: handler.service.GetStatus()}
}
