package migrations

import (
	"strings"
	"time"

	"cribeapp.com/cribe-server/internal/core/logger"
	"github.com/golang-migrate/migrate/v4"
)

type MigrationServiceInterface interface {
	DoDryRunMigrations() []Migration
	DoLiveRunMigrations() []Migration
}

type MigrationExecutor interface {
	Up() error
}

type MigrationFile interface {
	Name() string
}

type MigrationService struct {
	repo              MigrationRepository
	filesReader       func() ([]MigrationFile, error)
	getCurrentTime    func() time.Time
	migrationsManager func() (MigrationExecutor, error)
	logger            *logger.ContextualLogger
}

func NewMigrationService(service MigrationService) *MigrationService {
	return &MigrationService{
		repo:              service.repo,
		filesReader:       service.filesReader,
		getCurrentTime:    service.getCurrentTime,
		migrationsManager: service.migrationsManager,
		logger:            logger.NewServiceLogger("MigrationService"),
	}
}

func (s *MigrationService) fetchMigrations() []Migration {
	s.logger.Debug("Fetching pending migrations")

	files, err := s.filesReader()
	if err != nil {
		s.logger.Error("Unable to read migrations directory", map[string]any{
			"error": err.Error(),
		})
		return nil
	}

	s.logger.Debug("Migration files loaded", map[string]any{
		"filesCount": len(files),
	})

	lastExecutedMigration, err := s.repo.GetLastMigration()
	if err != nil {
		s.logger.Warn("Could not get last executed migration", map[string]any{
			"error": err.Error(),
		})
	}

	migrations := []Migration{}
	lastUpIndex := len(files) / 2

	for i := len(files) - 1; i >= 0; i-- {
		file := files[i]
		if !strings.HasSuffix(file.Name(), ".up.sql") {
			continue
		}

		name := strings.TrimSuffix(file.Name(), ".up.sql")
		if name == lastExecutedMigration.Name {
			s.logger.Debug("Found last executed migration", map[string]any{
				"migrationName": name,
			})
			break
		}

		migrations = append(migrations, Migration{ID: lastUpIndex, Name: name, CreatedAt: s.getCurrentTime()})
		lastUpIndex--
	}

	s.logger.Info("Pending migrations identified", map[string]any{
		"pendingCount": len(migrations),
	})

	return migrations
}

func (s *MigrationService) execMigrationsUp() []Migration {
	s.logger.Debug("Executing migrations up")

	m, err := s.migrationsManager()
	if err != nil {
		s.logger.Error("Unable to create migration instance", map[string]any{
			"error": err.Error(),
		})
		return []Migration{}
	}

	// Force the migration to run
	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		s.logger.Error("Unable to migrate up", map[string]any{
			"error": err.Error(),
		})
		return []Migration{}
	}

	if err == migrate.ErrNoChange {
		s.logger.Info("No migrations to apply")
		return []Migration{}
	}

	// Get the list of migrations that should have been run
	migrations := s.fetchMigrations()
	if len(migrations) == 0 {
		s.logger.Info("No migrations were applied")
		return []Migration{}
	}

	// Save the top migration to track it
	topMigration := migrations[0]
	err = s.repo.SaveMigration(topMigration.Name)
	if err != nil {
		s.logger.Error("Couldn't update migrations table", map[string]any{
			"migrationName": topMigration.Name,
			"error":         err.Error(),
		})
	} else {
		s.logger.Info("Migration tracking updated", map[string]any{
			"migrationName": topMigration.Name,
		})
	}

	s.logger.Info("Migrations executed successfully", map[string]any{
		"appliedCount": len(migrations),
	})

	return migrations
}

func (s *MigrationService) DoDryRunMigrations() []Migration {
	s.logger.Debug("Performing migration dry run")
	return s.fetchMigrations()
}

func (s *MigrationService) DoLiveRunMigrations() []Migration {
	s.logger.Debug("Performing migration live run")
	return s.execMigrationsUp()
}
