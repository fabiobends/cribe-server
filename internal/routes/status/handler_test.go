package status

import (
	"testing"

	"cribeapp.com/cribe-server/internal/utils"
)

type MockService struct{}

func (m *MockService) GetStatus() GetStatusResponse {
	return GetStatusResponse{
		UpdatedAt: utils.MockGetCurrentTimeISO(),
		Dependencies: Dependencies{
			Database: DatabaseInfo{
				Version:           "mocked version",
				MaxConnections:    1,
				OpenedConnections: 1,
			},
		},
	}
}

func TestStatusHandler_GetStatus(t *testing.T) {
	service := &MockService{}
	handler := NewStatusHandler(service)

	expected := GetStatusResponse{
		UpdatedAt: utils.MockGetCurrentTimeISO(),
		Dependencies: Dependencies{
			Database: DatabaseInfo{
				Version:           "mocked version",
				MaxConnections:    1,
				OpenedConnections: 1,
			},
		},
	}

	result := handler.GetStatus()

	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}
