package migrations

import (
	"testing"

	"cribeapp.com/cribe-server/internal/utils"
)

type SpyQueryExecutor struct {
	ExecCalledWith                []any
	HasExecBeenCalledSuccessfully bool
}

func MockExec(query string, args ...any) error {
	return nil
}

func (s *SpyQueryExecutor) Exec(query string, args ...any) error {
	s.ExecCalledWith = append([]any{query}, args...)
	err := MockExec(query, args...)

	if err == nil {
		s.HasExecBeenCalledSuccessfully = true
	}

	return err
}

func (s *SpyQueryExecutor) QueryItemWithPopulatedDatabase(query string, args ...any) Migration {
	if s.HasExecBeenCalledSuccessfully {
		return Migration{
			ID:        1,
			Name:      "000002_second",
			CreatedAt: utils.MockGetCurrentTime(),
		}
	}
	return Migration{
		ID:        1,
		Name:      "000001_initial",
		CreatedAt: utils.MockGetCurrentTime(),
	}
}

func (s *SpyQueryExecutor) QueryItemWithEmptyDatabase(query string, args ...any) Migration {
	if s.HasExecBeenCalledSuccessfully {
		return Migration{
			ID:        1,
			Name:      "000002_second",
			CreatedAt: utils.MockGetCurrentTime(),
		}
	}
	return Migration{}
}

func GetNewMockMigrationRepoWithEmptyDatabase() MigrationRepository {
	spy := SpyQueryExecutor{}
	return *NewMigrationRepository(WithQueryExecutor(QueryExecutor{QueryItem: spy.QueryItemWithEmptyDatabase, Exec: spy.Exec}))
}

func GetNewMockMigrationRepoWithPopulatedDatabase() MigrationRepository {
	spy := SpyQueryExecutor{}
	return *NewMigrationRepository(WithQueryExecutor(QueryExecutor{QueryItem: spy.QueryItemWithPopulatedDatabase, Exec: MockExec}))
}

func TestMigrationRepository_GetLastMigration(t *testing.T) {
	repo := GetNewMockMigrationRepoWithPopulatedDatabase()
	expected := Migration{
		ID:        1,
		Name:      "000001_initial",
		CreatedAt: utils.MockGetCurrentTime(),
	}

	result := repo.GetLastMigration()

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
