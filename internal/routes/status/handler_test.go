package status

import (
	"net/http"
	"net/http/httptest"
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

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/status", nil)

	handler.HandleRequest(w, r)

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

	result := utils.DecodeResponse[GetStatusResponse](w.Body.String())

	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}
