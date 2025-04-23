package status

import (
	"net/http"
	"strings"
	"testing"

	"cribeapp.com/cribe-server/internal/utils"
)

func TestStatusIntegration(t *testing.T) {
	resp := utils.MustSendTestRequest[GetStatusResponse](utils.TestRequest{
		Method:      http.MethodGet,
		URL:         "/status",
		HandlerFunc: HandleHTTPRequests,
	})

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
	}

	if !strings.Contains(resp.Body.Dependencies.Database.Version, "PostgreSQL 17") {
		t.Errorf("Expected version to be 17, got %s", resp.Body.Dependencies.Database.Version)
	}

	if !(resp.Body.Dependencies.Database.MaxConnections == 100) {
		t.Errorf("Expected max connections to be 100, got %d", resp.Body.Dependencies.Database.MaxConnections)
	}

	if !(resp.Body.Dependencies.Database.OpenedConnections == 1) {
		t.Errorf("Expected opened connections to be 1, got %d", resp.Body.Dependencies.Database.OpenedConnections)
	}

}
