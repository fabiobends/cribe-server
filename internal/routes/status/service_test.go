package status

import (
	"testing"

	"cribeapp.com/cribe-server/internal/utils"
)

type MockRepository struct{}

func (m *MockRepository) GetDatabaseInfo() (DatabaseInfo, error) {
	return DatabaseInfo{
		Version:           "mocked version",
		MaxConnections:    1,
		OpenedConnections: 1,
	}, nil
}

func TestStatusService_GetStatus(t *testing.T) {
	mockExecutor := QueryExecutor{
		QueryItem: func(query string, args ...interface{}) (DatabaseInfo, error) {
			return DatabaseInfo{
				Version:           "mocked version",
				MaxConnections:    1,
				OpenedConnections: 1,
			}, nil
		},
	}
	repo := NewStatusRepository(WithQueryExecutor(mockExecutor))
	service := NewStatusService(*repo, utils.MockGetCurrentTime)

	expected := GetStatusResponse{
		UpdatedAt: utils.MockGetCurrentTimeISO(),
		Dependencies: Dependencies{
			Database: DatabaseInfo{
				Version:           "mocked version",
				MaxConnections:    1,
				OpenedConnections: 1,
			}}}

	result := service.GetStatus()

	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}
