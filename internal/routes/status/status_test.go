package status

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"cribeapp.com/cribe-server/internal/utils"
)

func TestStatusIntegration(t *testing.T) {
	req, err := http.NewRequest("GET", "/status", nil)

	if err != nil {
		t.Fatalf("Could not create request: %v", err)
	}

	rec := httptest.NewRecorder()

	Handler(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, rec.Code)
	}

	result := utils.DecodeResponse[GetStatusResponse](rec.Body.String())

	if !strings.Contains(result.Dependencies.Database.Version, "PostgreSQL 17") {
		t.Errorf("Expected version to be 17, got %s", result.Dependencies.Database.Version)
	}

	if !(result.Dependencies.Database.MaxConnections == 100) {
		t.Errorf("Expected max connections to be 100, got %d", result.Dependencies.Database.MaxConnections)
	}

	if !(result.Dependencies.Database.OpenedConnections == 1) {
		t.Errorf("Expected opened connections to be 1, got %d", result.Dependencies.Database.OpenedConnections)
	}

}
