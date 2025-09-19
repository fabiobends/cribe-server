package status

import (
	"net/http"
	"strings"
	"testing"

	"cribeapp.com/cribe-server/internal/utils"
)

func TestStatusIntegration(t *testing.T) {
	resp := utils.MustSendTestRequest[StatusInfo](utils.TestRequest{
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

	if resp.Body.Dependencies.Database.MaxConnections != 100 {
		t.Errorf("Expected max connections to be 100, got %d", resp.Body.Dependencies.Database.MaxConnections)
	}

	if resp.Body.Dependencies.Database.OpenedConnections <= 0 {
		t.Errorf("Expected at least 1 opened connection, got %d", resp.Body.Dependencies.Database.OpenedConnections)
	}

}
