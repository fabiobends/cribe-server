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
		log.Printf("Unable to connect to database: %v", err)
	}

	return conn
}

func QueryItem[T any](query string, params ...any) T {
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

	return item
}

func QueryList[T any](query string, params ...any) []T {
	conn := newConnection()

	rows, err := conn.Query(context.Background(), query, params...)

	defer conn.Close(context.Background())

	if err != nil {
		log.Printf("Unable to query rows: %v", err)
		return []T{}
	}

	items, err := pgx.CollectRows(rows, pgx.RowToStructByNameLax[T])

	if err != nil {
		log.Printf("Unable to query rows: %v", err)
		return []T{}
	}

	return items
}

func Exec(query string, params ...any) error {
	conn := newConnection()

	_, err := conn.Exec(context.Background(), query, params...)

	defer conn.Close(context.Background())

	if err != nil {
		log.Printf("Unable to execute query: %v", err)
		return err
	}

	return nil
}
