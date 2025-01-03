package status

import "testing"

type MockRepository struct{}

func (m *MockRepository) GetStatusMessage() string {
	return "mocked message"
}

func TestStatusService_GetStatus(t *testing.T) {
	repo := &MockRepository{}
	service := NewStatusService(repo)

	expected := "mocked message"
	result := service.GetStatus()

	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}
