package status

import (
	"testing"
)

type MockQueryExecutor struct{}

func QueryItem(query string, param ...any) DatabaseInfo {
	return DatabaseInfo{
		Version:           "mocked version",
		MaxConnections:    1,
		OpenedConnections: 1,
	}
}

func TestStatusRepository_GetDatabaseInfo(t *testing.T) {
	repo := NewStatusRepository(WithQueryExecutor(QueryExecutor{QueryItem: QueryItem}))
	expected := DatabaseInfo{
		Version:           "mocked version",
		MaxConnections:    1,
		OpenedConnections: 1,
	}

	result := repo.GetDatabaseInfo()

	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}
