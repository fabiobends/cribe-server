package migrations

import (
	"time"

	"cribeapp.com/cribe-server/internal/utils"
)

type Migration struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type MigrationRepository struct {
	executor QueryExecutor
}

type QueryExecutor struct {
	QueryItem func(query string, args ...interface{}) Migration
	Exec      func(query string, args ...interface{}) error
}

type Option func(*MigrationRepository)

func WithQueryExecutor(executor QueryExecutor) Option {
	return func(r *MigrationRepository) {
		r.executor = executor
	}
}

func defaultExecutor() QueryExecutor {
	return QueryExecutor{
		QueryItem: utils.QueryItem[Migration],
		Exec:      utils.Exec,
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

func (r *MigrationRepository) GetLastMigration() Migration {
	return r.executor.QueryItem("SELECT * FROM migrations ORDER BY id DESC LIMIT 1")
}

func (r *MigrationRepository) SaveMigration(name string) error {
	return r.executor.Exec("INSERT INTO migrations (name) VALUES ($1)", name)
}
