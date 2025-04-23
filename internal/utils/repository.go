package utils

// QueryExecutor is a generic interface for executing database queries
type QueryExecutor[T any] struct {
	QueryItem func(query string, args ...any) (T, error)
	QueryList func(query string, args ...any) ([]T, error)
	Exec      func(query string, args ...any) error
}

// Repository is a generic interface for repository operations
type Repository[T any] struct {
	Executor QueryExecutor[T]
}

// Option is a function type for configuring repositories
type Option[T any] func(*Repository[T])

// WithQueryExecutor creates an option to set a custom query executor
func WithQueryExecutor[T any](executor QueryExecutor[T]) Option[T] {
	return func(r *Repository[T]) {
		r.Executor = executor
	}
}

// NewRepository creates a new repository with the given options
func NewRepository[T any](options ...Option[T]) *Repository[T] {
	repo := &Repository[T]{
		// The default executor is the database
		Executor: QueryExecutor[T]{
			QueryItem: QueryItem[T],
			QueryList: QueryList[T],
			Exec:      Exec,
		},
	}
	for _, option := range options {
		option(repo)
	}
	return repo
}
