package utils

import (
	"context"
	"os"

	"github.com/jackc/pgx/v5"
)

func newConnection() *pgx.Conn {
	databaseUrl := os.Getenv("DATABASE_URL")

	if os.Getenv("APP_ENV") == "test" {
		log.Debug("Using database connection", map[string]interface{}{
			"database_url": databaseUrl,
		})
	}

	conn, err := pgx.Connect(context.Background(), databaseUrl)

	if err != nil {
		log.Error("Unable to connect to database", map[string]interface{}{
			"error": err.Error(),
		})
	}

	return conn
}

func QueryItem[T any](query string, params ...any) (T, error) {
	conn := newConnection()

	rows, err := conn.Query(context.Background(), query, params...)

	defer func() {
		if closeErr := conn.Close(context.Background()); closeErr != nil {
			log.Error("Failed to close database connection", map[string]interface{}{
				"error": closeErr.Error(),
			})
		}
	}()

	if err != nil {
		log.Error("Unable to query rows", map[string]interface{}{
			"error": err.Error(),
			"query": query,
		})
	}

	item, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByNameLax[T])

	if err != nil && err.Error() != "no rows in result set" {
		log.Error("Unable to collect row", map[string]interface{}{
			"error": err.Error(),
			"query": query,
		})
	}

	return item, err
}

func QueryList[T any](query string, params ...any) ([]T, error) {
	conn := newConnection()

	rows, err := conn.Query(context.Background(), query, params...)

	defer func() {
		if closeErr := conn.Close(context.Background()); closeErr != nil {
			log.Error("Failed to close database connection", map[string]interface{}{
				"error": closeErr.Error(),
			})
		}
	}()

	if err != nil {
		log.Error("Unable to query rows", map[string]interface{}{
			"error": err.Error(),
			"query": query,
		})
		return []T{}, err
	}

	items, err := pgx.CollectRows(rows, pgx.RowToStructByNameLax[T])

	if err != nil {
		log.Error("Unable to query rows", map[string]interface{}{
			"error": err.Error(),
			"query": query,
		})
		return []T{}, err
	}

	return items, err
}

func Exec(query string, params ...any) error {
	conn := newConnection()

	_, err := conn.Exec(context.Background(), query, params...)

	defer func() {
		if closeErr := conn.Close(context.Background()); closeErr != nil {
			log.Error("Failed to close database connection", map[string]interface{}{
				"error": closeErr.Error(),
			})
		}
	}()

	if err != nil {
		log.Error("Unable to execute query", map[string]interface{}{
			"error": err.Error(),
			"query": query,
		})
	}

	return err
}
