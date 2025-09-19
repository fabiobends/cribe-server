package migrations

import (
	"net/http"
	"testing"

	"cribeapp.com/cribe-server/internal/core/logger"
	"cribeapp.com/cribe-server/internal/utils"
)

var log = logger.NewCoreLogger("MigrationsRouterTest")

func TestMigrations_IntegrationTests(t *testing.T) {
	log.Info("Setting up test environment", nil)
	_ = utils.CleanDatabase()

	resp := utils.MustSendTestRequest[[]Migration](utils.TestRequest{
		Method:      http.MethodGet,
		URL:         "/migrations",
		HandlerFunc: HandleHTTPRequests,
	})

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
	}

	if len(resp.Body) == 0 {
		t.Errorf("Expected at least one migration, got %d", len(resp.Body))
	}

	resp = utils.MustSendTestRequest[[]Migration](utils.TestRequest{
		Method:      http.MethodPost,
		URL:         "/migrations",
		HandlerFunc: HandleHTTPRequests,
	})

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("Expected status code %d, got %d", http.StatusCreated, resp.StatusCode)
	}

	if len(resp.Body) == 0 {
		t.Error("Expected at least one migration, got 0")
	}

}
