package migrations

import (
	"net/http"
	"strings"

	"cribeapp.com/cribe-server/internal/core/logger"
	"cribeapp.com/cribe-server/internal/utils"
)

type MigrationHandler struct {
	service MigrationServiceInterface
	logger  *logger.ContextualLogger
}

func NewMigrationHandler(service MigrationServiceInterface) *MigrationHandler {
	return &MigrationHandler{
		service: service,
		logger:  logger.NewHandlerLogger("MigrationHandler"),
	}
}

func (handler *MigrationHandler) HandleRequest(w http.ResponseWriter, r *http.Request) {
	paths := strings.Split(strings.Trim(r.URL.Path, "/"), "/")

	handler.logger.Debug("Migration request received", map[string]any{
		"method":     r.Method,
		"path":       r.URL.Path,
		"pathLength": len(paths),
	})

	if len(paths) > 1 {
		handler.logger.Warn("Invalid migration route - extra path segments", map[string]any{
			"paths": paths,
		})
		utils.NotFound(w, r)
		return
	}

	switch r.Method {
	case http.MethodGet:
		handler.handleGet(w, r)
	case http.MethodPost:
		handler.handlePost(w, r)
	default:
		handler.logger.Warn("Method not allowed for migrations endpoint", map[string]any{
			"method": r.Method,
		})
		utils.NotAllowed(w)
	}
}

func (handler *MigrationHandler) handleGet(w http.ResponseWriter, _ *http.Request) {
	handler.logger.Debug("Processing migration dry run request")

	migrations := handler.service.DoDryRunMigrations()

	handler.logger.Info("Migration dry run completed", map[string]any{
		"migrationsCount": len(migrations),
	})

	utils.EncodeResponse(w, http.StatusOK, migrations)
}

func (handler *MigrationHandler) handlePost(w http.ResponseWriter, _ *http.Request) {
	handler.logger.Debug("Processing migration live run request")

	migrations := handler.service.DoLiveRunMigrations()

	if len(migrations) == 0 {
		handler.logger.Info("Migration live run completed - no migrations applied", map[string]any{
			"migrationsCount": 0,
		})
		utils.EncodeResponse(w, http.StatusOK, migrations)
	} else {
		handler.logger.Info("Migration live run completed - migrations applied", map[string]any{
			"migrationsCount": len(migrations),
		})
		utils.EncodeResponse(w, http.StatusCreated, migrations)
	}
}
