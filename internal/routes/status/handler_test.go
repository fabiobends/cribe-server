package status

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"cribeapp.com/cribe-server/internal/utils"
)

type MockService struct{}

func (m *MockService) GetStatus() StatusInfo {
	return StatusInfo{
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

	expected := StatusInfo{
		UpdatedAt: utils.MockGetCurrentTimeISO(),
		Dependencies: Dependencies{
			Database: DatabaseInfo{
				Version:           "mocked version",
				MaxConnections:    1,
				OpenedConnections: 1,
			},
		},
	}

	result, err := utils.DecodeResponse[StatusInfo](w.Body.String())
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestStatusHandler_DeleteStatus(t *testing.T) {
	service := &MockService{}
	handler := NewStatusHandler(service)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodDelete, "/status", nil)

	handler.HandleRequest(w, r)

	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("Expected status code %d, got %d", http.StatusMethodNotAllowed, w.Code)
	}

}
