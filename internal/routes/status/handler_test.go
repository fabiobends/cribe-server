package status

import (
	"testing"

	"cribeapp.com/cribe-server/internal/utils"
)

type MockService struct{}

func (m *MockService) GetStatus() string {
	return "mocked service message"
}

func TestStatusHandler_GetStatus(t *testing.T) {
	service := &MockService{}
	handler := NewStatusHandler(service)

	expected := utils.StandardResponse{Message: "mocked service message"}
	result := handler.GetStatus()

	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}
