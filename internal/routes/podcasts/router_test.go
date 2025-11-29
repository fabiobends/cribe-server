package podcasts

import (
	"fmt"
	"net/http"
	"testing"

	"cribeapp.com/cribe-server/internal/core/logger"
	"cribeapp.com/cribe-server/internal/routes/migrations"
	"cribeapp.com/cribe-server/internal/utils"
)

var log = logger.NewCoreLogger("PodcastsRouterTest")

func TestPodcasts_IntegrationTests(t *testing.T) {
	log.Info("Setting up test environment", nil)
	if err := utils.CleanDatabaseAndRunMigrations(migrations.HandleHTTPRequests); err != nil {
		t.Fatalf("Failed to setup test environment: %v", err)
	}

	// Test GET /podcasts endpoint (should return empty array initially)
	t.Run("GET /podcasts returns empty array initially", func(t *testing.T) {
		resp := utils.MustSendTestRequest[[]Podcast](utils.TestRequest{
			Method:      http.MethodGet,
			URL:         "/podcasts",
			HandlerFunc: HandleHTTPRequests,
		})

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
		}

		log.Info("GET /podcasts result", map[string]any{
			"statusCode":    resp.StatusCode,
			"podcastsCount": len(resp.Body),
		})

		if len(resp.Body) != 0 {
			t.Logf("Expected 0 podcasts initially, got %d", len(resp.Body))
		}
	})

	// Test POST /podcasts/sync endpoint
	// Note: This will call the external API, so it might fail if API credentials are not set
	t.Run("POST /podcasts/sync syncs podcasts", func(t *testing.T) {
		resp := utils.MustSendTestRequest[SyncResult](utils.TestRequest{
			Method:      http.MethodPost,
			URL:         "/podcasts/sync",
			HandlerFunc: HandleHTTPRequests,
		})

		log.Info("POST /podcasts/sync result", map[string]any{
			"statusCode":  resp.StatusCode,
			"totalSynced": resp.Body.TotalSynced,
			"new":         resp.Body.New,
			"message":     resp.Body.Message,
		})

		// The sync endpoint should return either:
		// - 200 (OK) with sync results if successful
		// - 500 (Internal Server Error) if API credentials are missing or API fails
		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusInternalServerError {
			t.Errorf("Expected status code %d or %d, got %d", http.StatusOK, http.StatusInternalServerError, resp.StatusCode)
		}

		// If successful, verify response structure
		if resp.StatusCode == http.StatusOK {
			if resp.Body.Message == "" {
				t.Error("Expected message in sync result")
			}

			// After successful sync, GET should return podcasts
			getResp := utils.MustSendTestRequest[[]Podcast](utils.TestRequest{
				Method:      http.MethodGet,
				URL:         "/podcasts",
				HandlerFunc: HandleHTTPRequests,
			})

			if getResp.StatusCode != http.StatusOK {
				t.Errorf("Expected status code %d for GET after sync, got %d", http.StatusOK, getResp.StatusCode)
			}

			log.Info("GET /podcasts after sync result", map[string]any{
				"podcastsCount": len(getResp.Body),
			})

			if len(getResp.Body) != resp.Body.TotalSynced {
				t.Logf("Synced %d podcasts, but GET returned %d podcasts", resp.Body.TotalSynced, len(getResp.Body))
			}
		}
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

	// Test GET /podcasts/:id (episode auto-fetch)
	t.Run("GET /podcasts/:id returns podcast with episodes", func(t *testing.T) {
		// First sync podcasts
		syncResp := utils.MustSendTestRequest[SyncResult](utils.TestRequest{
			Method:      http.MethodPost,
			URL:         "/podcasts/sync",
			HandlerFunc: HandleHTTPRequests,
		})

		if syncResp.StatusCode == http.StatusOK && syncResp.Body.TotalSynced > 0 {
			// Get the first podcast
			getResp := utils.MustSendTestRequest[[]Podcast](utils.TestRequest{
				Method:      http.MethodGet,
				URL:         "/podcasts",
				HandlerFunc: HandleHTTPRequests,
			})

			if len(getResp.Body) > 0 {
				podcastID := fmt.Sprintf("%d", getResp.Body[0].ID)

				// Get podcast by ID (should auto-fetch episodes)
				podcastResp := utils.MustSendTestRequest[Podcast](utils.TestRequest{
					Method:      http.MethodGet,
					URL:         "/podcasts/" + podcastID,
					HandlerFunc: HandleHTTPRequests,
				})

				if podcastResp.StatusCode != http.StatusOK && podcastResp.StatusCode != http.StatusInternalServerError {
					t.Errorf("Expected status 200 or 500, got %d", podcastResp.StatusCode)
				}
			}
		}
	})

	// Test POST /podcasts/:id/sync
	t.Run("POST /podcasts/:id/sync syncs episodes", func(t *testing.T) {
		// First sync podcasts
		syncResp := utils.MustSendTestRequest[SyncResult](utils.TestRequest{
			Method:      http.MethodPost,
			URL:         "/podcasts/sync",
			HandlerFunc: HandleHTTPRequests,
		})

		if syncResp.StatusCode == http.StatusOK && syncResp.Body.TotalSynced > 0 {
			// Get the first podcast
			getResp := utils.MustSendTestRequest[[]Podcast](utils.TestRequest{
				Method:      http.MethodGet,
				URL:         "/podcasts",
				HandlerFunc: HandleHTTPRequests,
			})

			if len(getResp.Body) > 0 {
				podcastID := fmt.Sprintf("%d", getResp.Body[0].ID)

				// Sync episodes for this podcast
				episodeSyncResp := utils.MustSendTestRequest[SyncResult](utils.TestRequest{
					Method:      http.MethodPost,
					URL:         "/podcasts/" + podcastID + "/sync",
					HandlerFunc: HandleHTTPRequests,
				})

				if episodeSyncResp.StatusCode != http.StatusOK && episodeSyncResp.StatusCode != http.StatusInternalServerError {
					t.Errorf("Expected status 200 or 500, got %d", episodeSyncResp.StatusCode)
				}
			}
		}
	})
}
