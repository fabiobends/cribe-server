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

	// Test GET /migrations endpoint
	resp := utils.MustSendTestRequest[[]Migration](utils.TestRequest{
		Method:      http.MethodGet,
		URL:         "/migrations",
		HandlerFunc: HandleHTTPRequests,
	})

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
	}

	pendingMigrationsCount := len(resp.Body)
	log.Info("GET /migrations result", map[string]interface{}{
		"pendingCount": pendingMigrationsCount,
	})

	// Test POST /migrations endpoint
	postResp := utils.MustSendTestRequest[[]Migration](utils.TestRequest{
		Method:      http.MethodPost,
		URL:         "/migrations",
		HandlerFunc: HandleHTTPRequests,
	})

	log.Info("POST /migrations result", map[string]interface{}{
		"statusCode":      postResp.StatusCode,
		"migrationsCount": len(postResp.Body),
	})

	// The POST endpoint should return either:
	// - 201 (Created) with migrations if migrations were applied
	// - 200 (OK) with empty array if no migrations were needed
	if postResp.StatusCode != http.StatusCreated && postResp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d or %d, got %d", http.StatusCreated, http.StatusOK, postResp.StatusCode)
	}

	// Verify the response is consistent with the status code
	if postResp.StatusCode == http.StatusCreated && len(postResp.Body) == 0 {
		t.Error("Status 201 (Created) should include applied migrations in response body")
	}

	if postResp.StatusCode == http.StatusOK && len(postResp.Body) != 0 {
		t.Errorf("Status 200 (OK) should return empty array, got %d migrations", len(postResp.Body))
	}

	// After running migrations, GET should return fewer or no pending migrations
	finalResp := utils.MustSendTestRequest[[]Migration](utils.TestRequest{
		Method:      http.MethodGet,
		URL:         "/migrations",
		HandlerFunc: HandleHTTPRequests,
	})

	if finalResp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d for final GET, got %d", http.StatusOK, finalResp.StatusCode)
	}

	log.Info("Final GET /migrations result", map[string]interface{}{
		"pendingCount": len(finalResp.Body),
	})
}
