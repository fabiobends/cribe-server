package migrations

type MigrationHandler struct {
	service MigrationServiceInterface
}

func NewMigrationHandler(service MigrationServiceInterface) *MigrationHandler {
	return &MigrationHandler{service: service}
}

func (handler *MigrationHandler) GetMigrations() []Migration {
	return handler.service.DoDryRunMigrations()
}

func (handler *MigrationHandler) PostMigrations() []Migration {
	return handler.service.DoLiveRunMigrations()
}
