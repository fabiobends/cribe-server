package transcripts

import (
	"net/http"
	"testing"

	"cribeapp.com/cribe-server/internal/core/logger"
	"cribeapp.com/cribe-server/internal/routes/migrations"
	"cribeapp.com/cribe-server/internal/utils"
)

var log = logger.NewCoreLogger("TranscriptsRouterTest")

func TestTranscripts_IntegrationTests(t *testing.T) {
	log.Info("Setting up test environment", nil)
	if err := utils.CleanDatabaseAndRunMigrations(migrations.HandleHTTPRequests); err != nil {
		t.Fatalf("Failed to setup test environment: %v", err)
	}

	t.Run("should return bad request for missing episode_id", func(t *testing.T) {
		resp := utils.MustSendTestRequest[map[string]string](utils.TestRequest{
			Method:      http.MethodGet,
			URL:         "/transcripts/stream/sse",
			HandlerFunc: HandleHTTPRequests,
		})

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, resp.StatusCode)
		}
	})

	t.Run("should return bad request for invalid episode_id", func(t *testing.T) {
		resp := utils.MustSendTestRequest[map[string]string](utils.TestRequest{
			Method:      http.MethodGet,
			URL:         "/transcripts/stream/sse?episode_id=invalid",
			HandlerFunc: HandleHTTPRequests,
		})

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, resp.StatusCode)
		}
	})

	t.Run("should return not found for unknown path", func(t *testing.T) {
		resp := utils.MustSendTestRequest[map[string]string](utils.TestRequest{
			Method:      http.MethodGet,
			URL:         "/transcripts/unknown",
			HandlerFunc: HandleHTTPRequests,
		})

		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("Expected status code %d, got %d", http.StatusNotFound, resp.StatusCode)
		}
	})
}
