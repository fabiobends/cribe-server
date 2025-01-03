package status

import (
	"testing"
)

func TestStatusRepository_GetStatusMessage(t *testing.T) {
	repo := &StatusRepository{}
	expected := "ok"

	result := repo.GetStatusMessage()

	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}
