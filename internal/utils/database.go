package utils

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// PgxConn interface to allow mocking
type PgxConn interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Close(ctx context.Context) error
}

// Database holds the connection
type Database[T any] struct {
	conn PgxConn
}

// DatabaseInterface allows dependency injection for tests
type DatabaseInterface interface {
	QueryItem(query string, params ...any) (any, error)
	QueryList(query string, params ...any) ([]any, error)
	Exec(query string, params ...any) error
}

// NewDatabase returns a Database with an optional connection
// If conn is nil, it will create a new connection internally
func NewDatabase[T any](conn PgxConn) *Database[T] {
	if conn == nil {
		conn = NewConnection()
	}
	return &Database[T]{conn: conn}
}

// NewConnection creates a new pgx connection
func NewConnection() PgxConn {
	databaseUrl := os.Getenv("DATABASE_URL")

	if os.Getenv("APP_ENV") == "test" {
		log.Debug("Using database connection", map[string]any{
			"database_url": databaseUrl,
		})
	}

	conn, err := pgx.Connect(context.Background(), databaseUrl)
	if err != nil {
		log.Error("Unable to connect to database", map[string]any{
			"error": err.Error(),
		})
		return nil
	}

	return conn
}

// QueryItem fetches exactly one row and maps it to struct
func (db *Database[T]) QueryItem(query string, params ...any) (T, error) {
	var zero T
	conn := db.conn
	if conn == nil {
		return zero, fmt.Errorf("failed to connect to database")
	}

	rows, err := conn.Query(context.Background(), query, params...)
	if err != nil {
		log.Error("Unable to query rows", map[string]any{
			"error": err.Error(),
			"query": query,
		})
		return zero, err
	}
	defer rows.Close()

	item, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByNameLax[T])
	if err != nil && err.Error() != "no rows in result set" {
		log.Error("Unable to collect row", map[string]any{
			"error": err.Error(),
			"query": query,
		})
	}

	return item, err
}

// QueryList fetches multiple rows and maps them to a slice of structs
func (db *Database[T]) QueryList(query string, params ...any) ([]T, error) {
	conn := db.conn
	if conn == nil {
		return []T{}, fmt.Errorf("failed to connect to database")
	}

	rows, err := conn.Query(context.Background(), query, params...)
	if err != nil {
		log.Error("Unable to query rows", map[string]any{
			"error": err.Error(),
			"query": query,
		})
		return []T{}, err
	}
	defer rows.Close()

	items, err := pgx.CollectRows(rows, pgx.RowToStructByNameLax[T])
	if err != nil {
		log.Error("Unable to collect rows", map[string]any{
			"error": err.Error(),
			"query": query,
		})
		return []T{}, err
	}

	return items, nil
}

// Exec executes a query without returning rows
func (db *Database[T]) Exec(query string, params ...any) error {
	conn := db.conn
	if conn == nil {
		return fmt.Errorf("failed to connect to database")
	}

	_, err := conn.Exec(context.Background(), query, params...)
	if err != nil {
		log.Error("Unable to execute query", map[string]any{
			"error": err.Error(),
			"query": query,
		})
		return err
	}

	return nil
}
