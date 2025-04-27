package migrations

import (
	"net/http"

	"cribeapp.com/cribe-server/internal/utils"
)

type MigrationHandler struct {
	service MigrationServiceInterface
}

func NewMigrationHandler(service MigrationServiceInterface) *MigrationHandler {
	return &MigrationHandler{service: service}
}

func (handler *MigrationHandler) HandleRequest(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		handler.handleGet(w, r)
	case http.MethodPost:
		handler.handlePost(w, r)
	default:
		utils.NotAllowed(w)
	}
}

func (handler *MigrationHandler) handleGet(w http.ResponseWriter, _ *http.Request) {
	migrations := handler.service.DoDryRunMigrations()
	utils.EncodeResponse(w, http.StatusOK, migrations)
}

func (handler *MigrationHandler) handlePost(w http.ResponseWriter, _ *http.Request) {
	migrations := handler.service.DoLiveRunMigrations()
	if len(migrations) == 0 {
		utils.EncodeResponse(w, http.StatusOK, migrations)
	} else {
		utils.EncodeResponse(w, http.StatusCreated, migrations)
	}
}
