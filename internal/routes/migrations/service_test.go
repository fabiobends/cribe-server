package migrations

import (
	"testing"
)

var service = NewMockMigrationServiceReady()

func TestMigrationsService_DoDryRunMigrations(t *testing.T) {
	availableMigrations := service.DoDryRunMigrations()

	if len(availableMigrations) != 2 {
		t.Errorf("Expected 2 migrations, got %d", len(availableMigrations))
	}

	if availableMigrations[0].Name != "000002_second" {
		t.Errorf("Expected migration name to be '000002_second', got %s", availableMigrations[0].Name)
	}
}

func TestMigrationsService_DoLiveRunMigrations(t *testing.T) {
	availableMigrations := service.DoLiveRunMigrations()

	if len(availableMigrations) != 2 {
		t.Errorf("Expected 2 migrations, got %d", len(availableMigrations))
	}
}
