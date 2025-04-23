package utils

import (
	"context"
	"log"
	"os"

	"github.com/jackc/pgx/v5"
)

func newConnection() *pgx.Conn {
	databaseUrl := os.Getenv("DATABASE_URL")

	if os.Getenv("APP_ENV") == "test" {
		log.Printf("Using database: %s", databaseUrl)
	}

	conn, err := pgx.Connect(context.Background(), databaseUrl)

	if err != nil {
		log.Printf("Unable to connect to database: %v", err)
	}

	return conn
}

func QueryItem[T any](query string, params ...any) (T, error) {
	conn := newConnection()

	rows, err := conn.Query(context.Background(), query, params...)

	defer conn.Close(context.Background())

	if err != nil {
		log.Printf("Unable to query rows: %v", err)
	}

	item, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByNameLax[T])

	if err != nil {
		log.Printf("Unable to collect row: %v", err)
	}

	return item, err
}

func QueryList[T any](query string, params ...any) ([]T, error) {
	conn := newConnection()

	rows, err := conn.Query(context.Background(), query, params...)

	defer conn.Close(context.Background())

	if err != nil {
		log.Printf("Unable to query rows: %v", err)
		return []T{}, err
	}

	items, err := pgx.CollectRows(rows, pgx.RowToStructByNameLax[T])

	if err != nil {
		log.Printf("Unable to query rows: %v", err)
		return []T{}, err
	}

	return items, err
}

func Exec(query string, params ...any) error {
	conn := newConnection()

	_, err := conn.Exec(context.Background(), query, params...)

	defer conn.Close(context.Background())

	if err != nil {
		log.Printf("Unable to execute query: %v", err)
	}

	return err
}
