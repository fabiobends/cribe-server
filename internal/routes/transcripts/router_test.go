package transcripts

import (
	"context"
	"net/http"
	"testing"

	"cribeapp.com/cribe-server/internal/core/logger"
	"cribeapp.com/cribe-server/internal/middlewares"
	"cribeapp.com/cribe-server/internal/routes/migrations"
	"cribeapp.com/cribe-server/internal/utils"
)

var log = logger.NewCoreLogger("TranscriptsRouterTest")

// handlerWithAuth wraps the handler with a test context that includes the test user ID
func handlerWithAuth(w http.ResponseWriter, r *http.Request) {
	ctx := context.WithValue(r.Context(), middlewares.UserIDContextKey, utils.TestUserID)
	r = r.WithContext(ctx)
	HandleHTTPRequests()(w, r)
}

func TestTranscripts_IntegrationTests(t *testing.T) {
	log.Info("Setting up test environment", nil)
	if err := utils.CleanDatabaseAndRunMigrations(migrations.HandleHTTPRequests); err != nil {
		t.Fatalf("Failed to setup test environment: %v", err)
	}

	// Create test data
	_ = utils.CreateTestUser(utils.TestUserID)
	_, _ = utils.CreateTestPodcast()
	_, _ = utils.CreateTestEpisode(utils.TestPodcastID)
	_, _ = utils.CreateTestTranscript(utils.TestUserID)

	log.Info("Test data created", map[string]any{
		"userID":       utils.TestUserID,
		"podcastID":    utils.TestPodcastID,
		"episodeID":    utils.TestEpisodeID,
		"transcriptID": utils.TestTranscriptID,
	})

	t.Run("GET /transcripts/stream/sse returns bad request for missing episode_id", func(t *testing.T) {
		resp := utils.MustSendTestRequest[map[string]string](utils.TestRequest{
			Method:      http.MethodGet,
			URL:         "/transcripts/stream/sse",
			HandlerFunc: handlerWithAuth,
		})

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, resp.StatusCode)
		}
	})

	t.Run("GET /transcripts/stream/sse returns bad request for invalid episode_id", func(t *testing.T) {
		resp := utils.MustSendTestRequest[map[string]string](utils.TestRequest{
			Method:      http.MethodGet,
			URL:         "/transcripts/stream/sse?episode_id=invalid",
			HandlerFunc: handlerWithAuth,
		})

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, resp.StatusCode)
		}
	})

	t.Run("GET /transcripts/unknown returns 404", func(t *testing.T) {
		resp := utils.MustSendTestRequest[map[string]string](utils.TestRequest{
			Method:      http.MethodGet,
			URL:         "/transcripts/unknown",
			HandlerFunc: handlerWithAuth,
		})

		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("Expected status code %d, got %d", http.StatusNotFound, resp.StatusCode)
		}
	})

	t.Run("PUT /transcripts/stream/sse returns 404", func(t *testing.T) {
		resp := utils.MustSendTestRequest[map[string]string](utils.TestRequest{
			Method:      http.MethodPut,
			URL:         "/transcripts/stream/sse",
			HandlerFunc: handlerWithAuth,
		})

		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("Expected status code %d, got %d", http.StatusNotFound, resp.StatusCode)
		}
	})
}
