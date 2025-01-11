package status

type StatusHandler struct {
	service StatusServiceInterface
}

func NewStatusHandler(service StatusServiceInterface) *StatusHandler {
	return &StatusHandler{service: service}
}

func (handler *StatusHandler) GetStatus() GetStatusResponse {
	return handler.service.GetStatus()
}
