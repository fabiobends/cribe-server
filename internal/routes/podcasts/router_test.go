package podcasts

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"cribeapp.com/cribe-server/internal/core/logger"
	"cribeapp.com/cribe-server/internal/middlewares"
	"cribeapp.com/cribe-server/internal/routes/migrations"
	"cribeapp.com/cribe-server/internal/utils"
)

var log = logger.NewCoreLogger("PodcastsRouterTest")

// handlerWithAuth injects userID into context for authenticated routes
func handlerWithAuth(w http.ResponseWriter, r *http.Request) {
	ctx := context.WithValue(r.Context(), middlewares.UserIDContextKey, utils.TestUserID)
	HandleHTTPRequests(w, r.WithContext(ctx))
}

func TestPodcasts_IntegrationTests(t *testing.T) {
	log.Info("Setting up test environment", nil)
	if err := utils.CleanDatabaseAndRunMigrations(migrations.HandleHTTPRequests); err != nil {
		t.Fatalf("Failed to setup test environment: %v", err)
	}

	var podcastID int
	var episodeID int

	// Setup: Create fake data
	t.Run("setup: create fake data in database", func(t *testing.T) {
		// Create test data
		_ = utils.CreateTestUser(utils.TestUserID)
		_, _ = utils.CreateTestPodcast()
		_, _ = utils.CreateTestEpisode(utils.TestPodcastID)

		log.Info("Test data created", map[string]any{
			"userID":    utils.TestUserID,
			"podcastID": utils.TestPodcastID,
			"episodeID": utils.TestEpisodeID,
		})

		podcastID = utils.TestPodcastID
		episodeID = utils.TestEpisodeID

		log.Info("Fake data created successfully", map[string]any{
			"podcastID": podcastID,
			"episodeID": episodeID,
		})
	})

	// Test GET /podcasts endpoint
	t.Run("GET /podcasts returns podcasts", func(t *testing.T) {
		resp := utils.MustSendTestRequest[[]Podcast](utils.TestRequest{
			Method:      http.MethodGet,
			URL:         "/podcasts",
			HandlerFunc: HandleHTTPRequests,
		})

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
		}

		if len(resp.Body) == 0 {
			t.Error("Expected at least one podcast")
		}

		log.Info("GET /podcasts result", map[string]any{
			"statusCode":    resp.StatusCode,
			"podcastsCount": len(resp.Body),
		})
	})

	// Test GET /podcasts/:id endpoint
	t.Run("GET /podcasts/:id returns specific podcast", func(t *testing.T) {
		resp := utils.MustSendTestRequest[Podcast](utils.TestRequest{
			Method:      http.MethodGet,
			URL:         fmt.Sprintf("/podcasts/%d", podcastID),
			HandlerFunc: HandleHTTPRequests,
		})

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
		}

		if resp.Body.ID != podcastID {
			t.Errorf("Expected podcast ID %d, got %d", podcastID, resp.Body.ID)
		}

		log.Info("GET /podcasts/:id result", map[string]any{
			"statusCode": resp.StatusCode,
			"podcastID":  resp.Body.ID,
			"name":       resp.Body.Name,
		})
	})

	// Test invalid POST path
	t.Run("POST /podcasts/invalid returns 404", func(t *testing.T) {
		resp := utils.MustSendTestRequest[map[string]any](utils.TestRequest{
			Method:      http.MethodPost,
			URL:         "/podcasts/invalid",
			HandlerFunc: HandleHTTPRequests,
		})

		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("Expected status code %d, got %d", http.StatusNotFound, resp.StatusCode)
		}
	})

	// Test invalid HTTP method
	t.Run("PUT /podcasts returns 405", func(t *testing.T) {
		resp := utils.MustSendTestRequest[map[string]any](utils.TestRequest{
			Method:      http.MethodPut,
			URL:         "/podcasts",
			HandlerFunc: HandleHTTPRequests,
		})

		if resp.StatusCode != http.StatusMethodNotAllowed {
			t.Errorf("Expected status code %d, got %d", http.StatusMethodNotAllowed, resp.StatusCode)
		}
	})

	// Test GET /podcasts/:id returns podcast with episodes
	t.Run("GET /podcasts/:id returns podcast with episodes", func(t *testing.T) {
		resp := utils.MustSendTestRequest[Podcast](utils.TestRequest{
			Method:      http.MethodGet,
			URL:         fmt.Sprintf("/podcasts/%d", podcastID),
			HandlerFunc: handlerWithAuth,
		})

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
		}

		if resp.Body.ID != podcastID {
			t.Errorf("Expected podcast ID %d, got %d", podcastID, resp.Body.ID)
		}

		if len(resp.Body.Episodes) == 0 {
			t.Error("Expected podcast to have episodes, got empty array")
		}

		log.Info("GET /podcasts/:id result", map[string]any{
			"statusCode":    resp.StatusCode,
			"podcastID":     resp.Body.ID,
			"episodesCount": len(resp.Body.Episodes),
		})
	})
}
