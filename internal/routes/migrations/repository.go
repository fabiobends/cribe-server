package migrations

import (
	"cribeapp.com/cribe-server/internal/utils"
)

type MigrationRepository struct {
	executor QueryExecutor
}

type QueryExecutor struct {
	QueryItem func(query string, args ...any) (Migration, error)
	Exec      func(query string, args ...any) error
}

type Option func(*MigrationRepository)

func WithQueryExecutor(executor QueryExecutor) Option {
	return func(r *MigrationRepository) {
		r.executor = executor
	}
}

var defaultRepo = utils.NewRepository[Migration]()

func defaultExecutor() QueryExecutor {
	return QueryExecutor{
		QueryItem: defaultRepo.Executor.QueryItem,
		Exec:      defaultRepo.Executor.Exec,
	}
}

func NewMigrationRepository(options ...Option) *MigrationRepository {
	repo := &MigrationRepository{
		executor: defaultExecutor(),
	}

	for _, opt := range options {
		opt(repo)
	}

	return repo
}

func (r *MigrationRepository) GetLastMigration() (Migration, error) {
	return r.executor.QueryItem("SELECT * FROM migrations ORDER BY id DESC LIMIT 1")
}

func (r *MigrationRepository) SaveMigration(name string) error {
	return r.executor.Exec("INSERT INTO migrations (name) VALUES ($1)", name)
}
