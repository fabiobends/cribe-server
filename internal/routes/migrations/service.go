package migrations

import (
	"log"
	"strings"
	"time"
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
}

func NewMigrationService(service MigrationService) *MigrationService {
	return &MigrationService{repo: service.repo, filesReader: service.filesReader, getCurrentTime: service.getCurrentTime, migrationsManager: service.migrationsManager}
}

func (s *MigrationService) fetchMigrations() []Migration {
	files, err := s.filesReader()

	if err != nil {
		log.Printf("Unable to read migrations directory or it is empty: %v", err)
		return nil
	}

	lastExecutedMigration, _ := s.repo.GetLastMigration()

	migrations := []Migration{}
	lastUpIndex := len(files) / 2
	for i := len(files) - 1; i >= 0; i-- {
		file := files[i]
		if !strings.HasSuffix(file.Name(), ".up.sql") {
			continue
		}

		name := strings.TrimSuffix(file.Name(), ".up.sql")
		if name == lastExecutedMigration.Name {
			log.Printf("Last executed migration: %s", name)
			break
		}

		migrations = append(migrations, Migration{ID: lastUpIndex, Name: name, CreatedAt: s.getCurrentTime()})
		lastUpIndex--
	}

	return migrations
}

func (s *MigrationService) execMigrationsUp() []Migration {
	m, err := s.migrationsManager()
	if err != nil {
		log.Fatalf("Unable to create migration instance: %v", err)
	}

	err = m.Up()
	if err != nil {
		log.Printf("Unable to migrate up: %v", err)
	} else {
		log.Printf("Migration up successful")
	}

	migrations := s.fetchMigrations()
	if len(migrations) == 0 {
		return []Migration{}
	}

	lastMigration := migrations[len(migrations)-1]

	err = s.repo.SaveMigration(lastMigration.Name)

	if err != nil {
		log.Printf("Couldn't update migrations table: %v", err)
	}

	return migrations
}

func (s *MigrationService) DoDryRunMigrations() []Migration {
	return s.fetchMigrations()
}

func (s *MigrationService) DoLiveRunMigrations() []Migration {
	return s.execMigrationsUp()
}
