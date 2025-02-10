package migrations

import (
	"testing"

	"cribeapp.com/cribe-server/internal/utils"
)

type MockMigrationExecutor struct {
}

func (m *MockMigrationExecutor) Up() error {
	return nil
}

type MigrationFileMock struct {
	Title string
}

func (f *MigrationFileMock) Name() string {
	return f.Title
}

func MockFilesReader() ([]MigrationFile, error) {
	return []MigrationFile{
		&MigrationFileMock{Title: "000001_initial.up.sql"},
		&MigrationFileMock{Title: "000001_initial.down.sql"},
		&MigrationFileMock{Title: "000002_second.up.sql"},
		&MigrationFileMock{Title: "000002_second.down.sql"},
	}, nil
}

func MockMigrationManager() (MigrationExecutor, error) {
	return &MockMigrationExecutor{}, nil
}

func GetNewMockMigrationService() *MigrationService {
	return NewMigrationService(MigrationService{repo: GetNewMockMigrationRepoWithEmptyDatabase(), filesReader: MockFilesReader, getCurrentTime: utils.MockGetCurrentTime, migrationsManager: MockMigrationManager})
}

func TestMigrationsService_DoDryRunMigrations(t *testing.T) {
	service := GetNewMockMigrationService()

	availableMigrations := service.DoDryRunMigrations()

	if len(availableMigrations) != 2 {
		t.Errorf("expected 2 migrations, got %d", len(availableMigrations))
	}

	if availableMigrations[0].Name != "000002_second" {
		t.Errorf("expected migration name to be '000002_second', got %s", availableMigrations[0].Name)
	}
}

func TestMigrationsService_DoLiveRunMigrations(t *testing.T) {
	service := GetNewMockMigrationService()
	availableMigrations := service.DoLiveRunMigrations()

	if len(availableMigrations) != 2 {
		t.Errorf("expected 2 migrations, got %d", len(availableMigrations))
	}
}
