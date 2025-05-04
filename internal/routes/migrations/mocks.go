package migrations

import "cribeapp.com/cribe-server/internal/utils"

func MockExec(query string, args ...any) error {
	return nil
}

type SpyQueryExecutor struct {
	ExecCalledWith                []any
	HasExecBeenCalledSuccessfully bool
}

func (s *SpyQueryExecutor) Exec(query string, args ...any) error {
	s.ExecCalledWith = append([]any{query}, args...)
	err := MockExec(query, args...)

	if err == nil {
		s.HasExecBeenCalledSuccessfully = true
	}

	return err
}

func (s *SpyQueryExecutor) QueryItemWithPopulatedDatabase(query string, args ...any) (Migration, error) {
	if s.HasExecBeenCalledSuccessfully {
		return Migration{
			ID:        1,
			Name:      "000002_second",
			CreatedAt: utils.MockGetCurrentTime(),
		}, nil
	}
	return Migration{
		ID:        1,
		Name:      "000001_initial",
		CreatedAt: utils.MockGetCurrentTime(),
	}, nil
}

func (s *SpyQueryExecutor) QueryItemWithEmptyDatabase(query string, args ...any) (Migration, error) {
	if s.HasExecBeenCalledSuccessfully {
		return Migration{
			ID:        1,
			Name:      "000002_second",
			CreatedAt: utils.MockGetCurrentTime(),
		}, nil
	}
	return Migration{}, nil
}

func NewMockMigrationRepoWithEmptyDatabase() MigrationRepository {
	spy := SpyQueryExecutor{}
	return *NewMigrationRepository(WithQueryExecutor(QueryExecutor{QueryItem: spy.QueryItemWithEmptyDatabase, Exec: spy.Exec}))
}

func NewMockMigrationRepoWithPopulatedDatabase() MigrationRepository {
	spy := SpyQueryExecutor{}
	return *NewMigrationRepository(WithQueryExecutor(QueryExecutor{QueryItem: spy.QueryItemWithPopulatedDatabase, Exec: MockExec}))
}

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
func NewMockMigrationServiceReady() *MigrationService {
	repo := NewMockMigrationRepoWithEmptyDatabase()
	return NewMigrationService(MigrationService{repo: repo, filesReader: MockFilesReader, getCurrentTime: utils.MockGetCurrentTime, migrationsManager: MockMigrationManager})
}

func NewMockMigrationHandlerReady() *MigrationHandler {
	service := NewMockMigrationServiceReady()
	return NewMigrationHandler(service)
}
