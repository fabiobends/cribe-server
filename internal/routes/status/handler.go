package status

import (
	"net/http"

	"cribeapp.com/cribe-server/internal/core/logger"
	"cribeapp.com/cribe-server/internal/utils"
)

type StatusHandler struct {
	service StatusServiceInterface
	logger  *logger.ContextualLogger
}

func NewStatusHandler(service StatusServiceInterface) *StatusHandler {
	return &StatusHandler{
		service: service,
		logger:  logger.NewHandlerLogger("StatusHandler"),
	}
}

func (handler *StatusHandler) HandleRequest(w http.ResponseWriter, r *http.Request) {
	handler.logger.Debug("Status request received", map[string]any{
		"method": r.Method,
		"path":   r.URL.Path,
	})

	switch r.Method {
	case http.MethodGet:
		handler.handleGet(w, r)
	default:
		handler.logger.Warn("Method not allowed for status endpoint", map[string]any{
			"method": r.Method,
		})
		utils.NotAllowed(w)
	}
}

func (handler *StatusHandler) handleGet(w http.ResponseWriter, r *http.Request) {
	handler.logger.Debug("Processing status GET request")

	response := handler.service.GetStatus()

	handler.logger.Info("Status response generated successfully", map[string]any{
		"updatedAt": response.UpdatedAt,
	})

	utils.EncodeResponse(w, http.StatusOK, response)
}
