package utils

import (
	"context"
	"log"
	"os"

	"github.com/jackc/pgx/v5"
)

func newConnection() *pgx.Conn {
	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))

	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}

	return conn
}

func QueryItem[T any](query string, params ...any) T {
	conn := newConnection()

	rows, err := conn.Query(context.Background(), query, params...)

	if err != nil {
		log.Fatalf("Unable to query rows: %v", err)
	}

	item, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByNameLax[T])

	if err != nil {
		log.Fatalf("Unable to collect row: %v", err)
	}

	conn.Close(context.Background())

	return item
}

func QueryList[T any](query string, params ...any) []T {
	conn := newConnection()

	rows, err := conn.Query(context.Background(), query, params...)

	if err != nil {
		log.Fatalf("Unable to query rows: %v", err)
	}

	items, err := pgx.CollectRows(rows, pgx.RowToStructByNameLax[T])

	if err != nil {
		log.Fatalf("Unable to query rows: %v", err)
	}

	conn.Close(context.Background())

	return items
}
