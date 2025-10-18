package utils

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/pashagolub/pgxmock/v4"
)

type TestItem struct {
	ID   int    `db:"id"`
	Name string `db:"name"`
}

func TestQueryItem_SyntaxError(t *testing.T) {
	conn, _ := pgxmock.NewConn()
	defer func() { _ = conn.Close(context.Background()) }()
	db := NewDatabase[TestItem](conn)

	// Expect the query to fail with syntax error
	conn.ExpectQuery("SELEC").WillReturnError(fmt.Errorf("syntax error"))

	// Force SQL error → triggers `log.Error` branch
	_, err := db.QueryItem("SELEC * FROM test_items")
	if err == nil {
		t.Error("Expected syntax error, got nil")
	}

	if err := conn.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet expectations: %v", err)
	}
}

func TestQueryItemAndQueryList(t *testing.T) {
	conn, _ := pgxmock.NewConn()
	defer func() { _ = conn.Close(context.Background()) }()
	db := NewDatabase[TestItem](conn)

	// Set up expectations for Insert
	conn.ExpectExec("INSERT INTO test_items").WithArgs("Alice", "Bob").WillReturnResult(pgxmock.NewResult("INSERT", 2))

	// Insert test data
	err := db.Exec("INSERT INTO test_items (name) VALUES ($1), ($2)", "Alice", "Bob")
	if err != nil {
		t.Fatalf("Insert failed: %v", err)
	}

	if err := conn.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet expectations: %v", err)
	}

	// Test QueryItem happy path - need new connection for each test
	conn2, _ := pgxmock.NewConn()
	defer func() { _ = conn2.Close(context.Background()) }()
	db2 := NewDatabase[TestItem](conn2)

	rows := pgxmock.NewRows([]string{"id", "name"}).AddRow(1, "Alice")
	conn2.ExpectQuery("SELECT id, name FROM test_items WHERE name").WithArgs("Alice").WillReturnRows(rows)

	item, err := db2.QueryItem("SELECT id, name FROM test_items WHERE name = $1", "Alice")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if item.Name != "Alice" {
		t.Errorf("Expected Alice, got %s", item.Name)
	}

	if err := conn2.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet expectations: %v", err)
	}

	// Test QueryItem no rows - need another new connection
	conn3, _ := pgxmock.NewConn()
	defer func() { _ = conn3.Close(context.Background()) }()
	db3 := NewDatabase[TestItem](conn3)

	conn3.ExpectQuery("SELECT id, name FROM test_items WHERE name").WithArgs("Charlie").WillReturnError(fmt.Errorf("no rows in result set"))

	_, err = db3.QueryItem("SELECT id, name FROM test_items WHERE name = $1", "Charlie")
	if err == nil {
		t.Error("Expected error for no rows, got nil")
	}

	if err := conn3.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet expectations: %v", err)
	}

	// Test QueryList happy path - need another new connection
	conn4, _ := pgxmock.NewConn()
	defer func() { _ = conn4.Close(context.Background()) }()
	db4 := NewDatabase[TestItem](conn4)

	rows2 := pgxmock.NewRows([]string{"id", "name"}).AddRow(1, "Alice").AddRow(2, "Bob")
	conn4.ExpectQuery("SELECT id, name FROM test_items ORDER BY id").WillReturnRows(rows2)

	items, err := db4.QueryList("SELECT id, name FROM test_items ORDER BY id")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if len(items) != 2 {
		t.Errorf("Expected 2 rows, got %d", len(items))
	}

	if err := conn4.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet expectations: %v", err)
	}
}

func TestQueryList_SyntaxError(t *testing.T) {
	conn, _ := pgxmock.NewConn()
	defer func() { _ = conn.Close(context.Background()) }()
	db := NewDatabase[TestItem](conn)

	// Expect the query to fail with syntax error
	conn.ExpectQuery("SELEC").WillReturnError(fmt.Errorf("syntax error"))

	// Force SQL error → triggers `log.Error` branch
	result, err := db.QueryList("SELEC * FROM test_items")
	if err == nil {
		t.Error("Expected syntax error, got nil")
	}

	if err := conn.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet expectations: %v", err)
	}

	if len(result) != 0 {
		t.Errorf("Expected zero results on error, got %d", len(result))
	}
}

func TestQueryList_RowMappingError(t *testing.T) {
	conn, _ := pgxmock.NewConn()
	defer func() { _ = conn.Close(context.Background()) }()
	db := NewDatabase[TestItem](conn)

	// Set up expectations for Insert
	conn.ExpectExec("INSERT INTO test_items").WithArgs("Alice").WillReturnResult(pgxmock.NewResult("INSERT", 1))

	// Insert one row
	err := db.Exec("INSERT INTO test_items (name) VALUES ($1)", "Alice")
	if err != nil {
		t.Fatalf("Insert failed: %v", err)
	}

	if err := conn.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet expectations: %v", err)
	}

	// Test mapping error with new connection
	conn2, _ := pgxmock.NewConn()
	defer func() { _ = conn2.Close(context.Background()) }()
	db2 := NewDatabase[TestItem](conn2)

	// Return rows with wrong column name that won't map to struct
	rows := pgxmock.NewRows([]string{"wrong_column"}).AddRow(1)
	conn2.ExpectQuery("SELECT id AS wrong_column FROM test_items").WillReturnRows(rows)

	// Query column that doesn't match struct → CollectRows should fail
	_, err = db2.QueryList("SELECT id AS wrong_column FROM test_items")
	if err == nil {
		t.Error("Expected mapping error from CollectRows, got nil")
	}

	if err := conn2.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet expectations: %v", err)
	}
}

func TestQueryItem_RowMappingError(t *testing.T) {
	conn, _ := pgxmock.NewConn()
	defer func() { _ = conn.Close(context.Background()) }()
	db := NewDatabase[TestItem](conn)

	// Set up expectations for Insert
	conn.ExpectExec("INSERT INTO test_items").WithArgs("Alice").WillReturnResult(pgxmock.NewResult("INSERT", 1))

	err := db.Exec("INSERT INTO test_items (name) VALUES ($1)", "Alice")
	if err != nil {
		t.Fatalf("Insert failed: %v", err)
	}

	if err := conn.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet expectations: %v", err)
	}

	// Test mapping error with new connection
	conn2, _ := pgxmock.NewConn()
	defer func() { _ = conn2.Close(context.Background()) }()
	db2 := NewDatabase[TestItem](conn2)

	// Return rows with wrong column name that won't map to struct
	rows := pgxmock.NewRows([]string{"wrong_column"}).AddRow(1)
	conn2.ExpectQuery("SELECT id AS wrong_column FROM test_items WHERE name").WithArgs("Alice").WillReturnRows(rows)

	// Query column that doesn't match struct → CollectExactlyOneRow should fail
	_, err = db2.QueryItem("SELECT id AS wrong_column FROM test_items WHERE name = $1", "Alice")
	if err == nil {
		t.Error("Expected mapping error from CollectExactlyOneRow, got nil")
	}

	if err := conn2.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet expectations: %v", err)
	}
}

func TestQueryItem_TableNotFoundError(t *testing.T) {
	conn, _ := pgxmock.NewConn()
	defer func() { _ = conn.Close(context.Background()) }()
	db := NewDatabase[TestItem](conn)

	// Expect query to fail - table doesn't exist
	conn.ExpectQuery("SELECT \\* FROM table_that_does_not_exist").WillReturnError(fmt.Errorf("relation \"table_that_does_not_exist\" does not exist"))

	// Table doesn't exist → query itself will fail
	_, err := db.QueryItem("SELECT * FROM table_that_does_not_exist")
	if err == nil {
		t.Error("Expected query error for missing table, got nil")
	}

	if err := conn.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet expectations: %v", err)
	}
}

func TestExec_SuccessAndError(t *testing.T) {
	// Test happy path
	conn, _ := pgxmock.NewConn()
	defer func() { _ = conn.Close(context.Background()) }()
	db := NewDatabase[TestItem](conn)

	// Set up expectations for successful Insert
	conn.ExpectExec("INSERT INTO test_items").WithArgs("Charlie").WillReturnResult(pgxmock.NewResult("INSERT", 1))

	err := db.Exec("INSERT INTO test_items (name) VALUES ($1)", "Charlie")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if err := conn.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet expectations: %v", err)
	}

	// Test error path with new connection
	conn2, _ := pgxmock.NewConn()
	defer func() { _ = conn2.Close(context.Background()) }()
	db2 := NewDatabase[TestItem](conn2)

	// Set up expectations for failed Insert (syntax error)
	conn2.ExpectExec("INSER INTO test_items").WithArgs("BadSQL").WillReturnError(fmt.Errorf("syntax error at or near \"INSER\""))

	// Force SQL error → triggers `log.Error` branch
	err = db2.Exec("INSER INTO test_items (name) VALUES ($1)", "BadSQL") // typo "INSERT"
	if err == nil {
		t.Error("Expected syntax error, got nil")
	}

	if err := conn2.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet expectations: %v", err)
	}
}

func TestNewConnection_InvalidURL(t *testing.T) {
	oldURL := os.Getenv("DATABASE_URL")
	_ = os.Setenv("DATABASE_URL", "invalid-url")
	defer func() { _ = os.Setenv("DATABASE_URL", oldURL) }()

	conn := NewConnection()
	if conn != nil {
		_, err := conn.Exec(context.Background(), "SELECT 1")
		if err == nil {
			t.Error("Expected error for invalid URL, query succeeded")
		}
	}
}
