package status

import (
	"os"

	"cribeapp.com/cribe-server/internal/utils"
)

type QueryExecutor struct {
	QueryItem func(query string, args ...any) (DatabaseInfo, error)
}

type StatusRepository struct {
	executor QueryExecutor
}

type Option func(*StatusRepository)

func WithQueryExecutor(executor QueryExecutor) Option {
	return func(r *StatusRepository) {
		r.executor = executor
	}
}

func defaultExecutor() QueryExecutor {
	return QueryExecutor{
		QueryItem: utils.QueryItem[DatabaseInfo],
	}
}

func NewStatusRepository(options ...Option) *StatusRepository {
	repo := &StatusRepository{
		executor: defaultExecutor(),
	}

	for _, opt := range options {
		opt(repo)
	}

	return repo
}

func (r *StatusRepository) GetDatabaseInfo() (DatabaseInfo, error) {
	query := `
		SELECT
			version() AS version,
			current_setting('max_connections')::int AS max_connections,
			(SELECT COUNT(*)::int FROM pg_stat_activity WHERE datname = $1) AS opened_connections;
	`
	return r.executor.QueryItem(query, os.Getenv("POSTGRES_DB"))
}
