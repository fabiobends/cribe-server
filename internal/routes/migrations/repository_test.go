package migrations

import (
	"testing"

	"cribeapp.com/cribe-server/internal/utils"
)

func TestMigrationRepository_GetLastMigration(t *testing.T) {
	repo := NewMockMigrationRepoWithPopulatedDatabase()
	expected := Migration{
		ID:        1,
		Name:      "000001_initial",
		CreatedAt: utils.MockGetCurrentTime(),
	}

	result, _ := repo.GetLastMigration()

	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestMigrationRepository_SaveMigration(t *testing.T) {
	spy := SpyQueryExecutor{}
	repo := NewMigrationRepository(WithQueryExecutor(QueryExecutor{Exec: spy.Exec}))
	name := "000001_initial"

	err := repo.SaveMigration(name)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if len(spy.ExecCalledWith) == 0 {
		t.Fatal("Expected `Exec` to be called with arguments, but got none")
	}

	secondArgument := spy.ExecCalledWith[1]

	if secondArgument != name {
		t.Errorf("expected second argument to be %q, got %q", name, secondArgument)
	}
}
