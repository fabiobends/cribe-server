package migrations

import (
	"fmt"
	"testing"

	"cribeapp.com/cribe-server/internal/utils"
)

func TestMigrationsHandler_GetMigrations(t *testing.T) {
	service := GetNewMockMigrationService()
	handler := NewMigrationHandler(service)

	expected := []Migration{
		{
			ID:        2,
			Name:      "000002_second",
			CreatedAt: utils.MockGetCurrentTime(),
		},
		{
			ID:        1,
			Name:      "000001_initial",
			CreatedAt: utils.MockGetCurrentTime(),
		},
	}

	result := handler.GetMigrations()

	if fmt.Sprint(result) != fmt.Sprint(expected) {
		t.Errorf("expected %q, got %q", expected, result)
	}

}

func TestMigrationsHandler_PostMigrations(t *testing.T) {
	service := GetNewMockMigrationService()
	handler := NewMigrationHandler(service)

	// First run
	firstResult := handler.PostMigrations()
	firstExpected := []Migration{
		{
			ID:        2,
			Name:      "000002_second",
			CreatedAt: utils.MockGetCurrentTime(),
		},
		{
			ID:        1,
			Name:      "000001_initial",
			CreatedAt: utils.MockGetCurrentTime(),
		},
	}

	if fmt.Sprint(firstResult) != fmt.Sprint(firstExpected) {
		t.Errorf("expected %q, got %q", firstExpected, firstResult)
	}

	// Second run
	secondResult := handler.PostMigrations()
	secondExpected := []Migration{}

	if fmt.Sprint(secondResult) != fmt.Sprint(secondExpected) {
		t.Errorf("expected %q, got %q", secondExpected, secondResult)
	}
}
