package utils

import (
	"errors"
	"testing"
)

func TestNewRepository(t *testing.T) {
	type TestStruct struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}

	// Test creating repository with default options
	repo := NewRepository[TestStruct]()

	if repo == nil {
		t.Error("NewRepository should not return nil")
		return
	}

	if repo.Executor.QueryItem == nil {
		t.Error("Repository should have QueryItem function")
	}

	if repo.Executor.QueryList == nil {
		t.Error("Repository should have QueryList function")
	}

	if repo.Executor.Exec == nil {
		t.Error("Repository should have Exec function")
	}
}

func TestNewRepository_WithCustomExecutor(t *testing.T) {
	type TestStruct struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}

	// Create a mock executor
	mockExecutor := QueryExecutor[TestStruct]{
		QueryItem: func(query string, args ...any) (TestStruct, error) {
			return TestStruct{ID: 1, Name: "test"}, nil
		},
		QueryList: func(query string, args ...any) ([]TestStruct, error) {
			return []TestStruct{{ID: 1, Name: "test"}}, nil
		},
		Exec: func(query string, args ...any) error {
			return nil
		},
	}

	// Test creating repository with custom executor
	repo := NewRepository[TestStruct](WithQueryExecutor(mockExecutor))

	if repo == nil {
		t.Error("NewRepository should not return nil")
		return
	}

	// Test that the mock executor was set
	result, err := repo.Executor.QueryItem("SELECT * FROM test")
	if err != nil {
		t.Errorf("Expected no error from mock executor, got: %v", err)
	}

	if result.ID != 1 || result.Name != "test" {
		t.Errorf("Expected ID=1, Name=test, got ID=%d, Name=%s", result.ID, result.Name)
	}
}

func TestWithQueryExecutor(t *testing.T) {
	type TestStruct struct {
		ID int `json:"id"`
	}

	// Create a mock executor that returns an error
	mockExecutor := QueryExecutor[TestStruct]{
		QueryItem: func(query string, args ...any) (TestStruct, error) {
			return TestStruct{}, errors.New("mock error")
		},
		QueryList: func(query string, args ...any) ([]TestStruct, error) {
			return nil, errors.New("mock list error")
		},
		Exec: func(query string, args ...any) error {
			return errors.New("mock exec error")
		},
	}

	// Test the option function
	option := WithQueryExecutor(mockExecutor)

	// Apply the option to a repository
	repo := &Repository[TestStruct]{}
	option(repo)

	// Test that the executor was set correctly
	_, err := repo.Executor.QueryItem("test")
	if err == nil {
		t.Error("Expected error from mock executor")
	}
	if err.Error() != "mock error" {
		t.Errorf("Expected 'mock error', got '%s'", err.Error())
	}

	_, err = repo.Executor.QueryList("test")
	if err == nil {
		t.Error("Expected error from mock executor")
	}
	if err.Error() != "mock list error" {
		t.Errorf("Expected 'mock list error', got '%s'", err.Error())
	}

	err = repo.Executor.Exec("test")
	if err == nil {
		t.Error("Expected error from mock executor")
	}
	if err.Error() != "mock exec error" {
		t.Errorf("Expected 'mock exec error', got '%s'", err.Error())
	}
}

func TestRepository_MultipleOptions(t *testing.T) {
	type TestStruct struct {
		ID int `json:"id"`
	}

	// Create first mock executor
	firstExecutor := QueryExecutor[TestStruct]{
		QueryItem: func(query string, args ...any) (TestStruct, error) {
			return TestStruct{ID: 1}, nil
		},
		QueryList: func(query string, args ...any) ([]TestStruct, error) {
			return []TestStruct{{ID: 1}}, nil
		},
		Exec: func(query string, args ...any) error {
			return errors.New("first executor")
		},
	}

	// Create second mock executor that should override the first
	secondExecutor := QueryExecutor[TestStruct]{
		QueryItem: func(query string, args ...any) (TestStruct, error) {
			return TestStruct{ID: 2}, nil
		},
		QueryList: func(query string, args ...any) ([]TestStruct, error) {
			return []TestStruct{{ID: 2}}, nil
		},
		Exec: func(query string, args ...any) error {
			return errors.New("second executor")
		},
	}

	// Test with multiple options (second should override first)
	repo := NewRepository[TestStruct](
		WithQueryExecutor(firstExecutor),
		WithQueryExecutor(secondExecutor),
	)

	// Test that the second executor was used
	result, err := repo.Executor.QueryItem("test")
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if result.ID != 2 {
		t.Errorf("Expected ID=2 (second executor), got ID=%d", result.ID)
	}

	err = repo.Executor.Exec("test")
	if err == nil {
		t.Error("Expected error from second executor")
	}
	if err.Error() != "second executor" {
		t.Errorf("Expected 'second executor', got '%s'", err.Error())
	}
}

func TestQueryExecutor_Struct(t *testing.T) {
	type TestStruct struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}

	// Test that QueryExecutor can be created with different function signatures
	executor := QueryExecutor[TestStruct]{
		QueryItem: func(query string, args ...any) (TestStruct, error) {
			return TestStruct{ID: 42, Name: "test"}, nil
		},
		QueryList: func(query string, args ...any) ([]TestStruct, error) {
			return []TestStruct{{ID: 1, Name: "first"}, {ID: 2, Name: "second"}}, nil
		},
		Exec: func(query string, args ...any) error {
			return nil
		},
	}

	// Test QueryItem
	item, err := executor.QueryItem("SELECT * FROM test WHERE id = ?", 42)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if item.ID != 42 || item.Name != "test" {
		t.Errorf("Expected ID=42, Name=test, got ID=%d, Name=%s", item.ID, item.Name)
	}

	// Test QueryList
	items, err := executor.QueryList("SELECT * FROM test")
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if len(items) != 2 {
		t.Errorf("Expected 2 items, got %d", len(items))
	}
	if items[0].ID != 1 || items[1].ID != 2 {
		t.Errorf("Expected IDs 1,2, got %d,%d", items[0].ID, items[1].ID)
	}

	// Test Exec
	err = executor.Exec("UPDATE test SET name = ?", "updated")
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
}
