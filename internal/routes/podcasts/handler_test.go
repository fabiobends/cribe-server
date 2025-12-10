package podcasts

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewPodcastHandler(t *testing.T) {
	mockRepo := MockPodcastRepo{}
	mockAPIClient := &MockAPIClient{}

	service := NewPodcastService(mockRepo, mockAPIClient)
	handler := NewPodcastHandler(service)

	if handler == nil {
		t.Error("Expected handler to be initialized")
	}
}

func TestHandlerHandleRequest_MethodNotAllowed(t *testing.T) {
	mockRepo := MockPodcastRepo{}
	mockAPIClient := &MockAPIClient{}

	service := NewPodcastService(mockRepo, mockAPIClient)
	handler := NewPodcastHandler(service)

	req := httptest.NewRequest(http.MethodPut, "/podcasts", nil)
	w := httptest.NewRecorder()

	handler.HandleRequest(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", w.Code)
	}
}

func TestHandlerHandleRequest_PostInvalidPath(t *testing.T) {
	mockRepo := MockPodcastRepo{}
	mockAPIClient := &MockAPIClient{}

	service := NewPodcastService(mockRepo, mockAPIClient)
	handler := NewPodcastHandler(service)

	req := httptest.NewRequest(http.MethodPost, "/podcasts/invalid", nil)
	w := httptest.NewRecorder()

	handler.HandleRequest(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}
}

func TestHandleGetByID_Success(t *testing.T) {
	mockRepo := MockPodcastRepo{
		getPodcastByIDFunc:         func(id int) (Podcast, error) { return Podcast{ID: 1, Name: "Test"}, nil },
		getEpisodesByPodcastIDFunc: func(podcastID int) ([]Episode, error) { return []Episode{{ID: 1}}, nil },
	}

	service := NewPodcastService(mockRepo, &MockAPIClient{})
	handler := NewPodcastHandler(service)

	req := httptest.NewRequest(http.MethodGet, "/podcasts/1", nil)
	w := httptest.NewRecorder()

	handler.handleGetByID(w, req, "1")

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestHandleGetByID_Error(t *testing.T) {
	mockRepo := MockPodcastRepo{
		getPodcastByIDFunc: func(id int) (Podcast, error) { return Podcast{}, fmt.Errorf("db error") },
	}

	service := NewPodcastService(mockRepo, &MockAPIClient{})
	handler := NewPodcastHandler(service)

	req := httptest.NewRequest(http.MethodGet, "/podcasts/1", nil)
	w := httptest.NewRecorder()

	handler.handleGetByID(w, req, "1")

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", w.Code)
	}
}

func TestHandleSyncEpisodes_Success(t *testing.T) {
	mockRepo := MockPodcastRepo{
		getPodcastByIDFunc:         func(id int) (Podcast, error) { return Podcast{ID: 1, ExternalID: "uuid-1"}, nil },
		getEpisodesByPodcastIDFunc: func(podcastID int) ([]Episode, error) { return []Episode{}, nil },
		upsertEpisodeFunc:          func(episode PodcastEpisode, podcastID int) (Episode, error) { return Episode{ID: 1}, nil },
	}

	mockAPIClient := &MockAPIClient{
		getPodcastByIDFunc: func(podcastID string) (*PodcastWithEpisodes, error) {
			return &PodcastWithEpisodes{
				Episodes: []PodcastEpisode{{UUID: "ep-1", Name: "Ep1", AudioURL: "url", DatePublished: 123, Duration: 100}},
			}, nil
		},
	}

	service := NewPodcastService(mockRepo, mockAPIClient)
	handler := NewPodcastHandler(service)

	req := httptest.NewRequest(http.MethodPost, "/podcasts/1/sync", nil)
	w := httptest.NewRecorder()

	handler.handleSyncEpisodes(w, req, "1")

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestHandleSyncEpisodes_Error(t *testing.T) {
	mockRepo := MockPodcastRepo{
		getPodcastByIDFunc: func(id int) (Podcast, error) { return Podcast{}, fmt.Errorf("db error") },
	}

	service := NewPodcastService(mockRepo, &MockAPIClient{})
	handler := NewPodcastHandler(service)

	req := httptest.NewRequest(http.MethodPost, "/podcasts/1/sync", nil)
	w := httptest.NewRecorder()

	handler.handleSyncEpisodes(w, req, "1")

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", w.Code)
	}
}

func TestHandleSync_Success(t *testing.T) {
	mockRepo := MockPodcastRepo{
		getPodcastByExternalIDFunc: func(externalID string) (Podcast, error) {
			return Podcast{}, fmt.Errorf("not found") // New podcasts
		},
		upsertPodcastFunc: func(podcast ExternalPodcast) (Podcast, error) {
			return Podcast{ID: 1, ExternalID: podcast.UUID}, nil
		},
	}

	mockAPIClient := &MockAPIClient{
		getTopPodcastsFunc: func() ([]ExternalPodcast, error) {
			return []ExternalPodcast{
				{UUID: "podcast-1", Name: "Test Podcast 1", AuthorName: "Author 1", ImageURL: "https://example.com/1.jpg", Description: "Test 1"},
				{UUID: "podcast-2", Name: "Test Podcast 2", AuthorName: "Author 2", ImageURL: "https://example.com/2.jpg", Description: "Test 2"},
			}, nil
		},
	}

	service := NewPodcastService(mockRepo, mockAPIClient)
	handler := NewPodcastHandler(service)

	req := httptest.NewRequest(http.MethodPost, "/podcasts/sync", nil)
	w := httptest.NewRecorder()

	handler.handleSync(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestHandleSync_Error(t *testing.T) {
	mockRepo := MockPodcastRepo{}

	mockAPIClient := &MockAPIClient{
		getTopPodcastsFunc: func() ([]ExternalPodcast, error) {
			return nil, fmt.Errorf("API error")
		},
	}

	service := NewPodcastService(mockRepo, mockAPIClient)
	handler := NewPodcastHandler(service)

	req := httptest.NewRequest(http.MethodPost, "/podcasts/sync", nil)
	w := httptest.NewRecorder()

	handler.handleSync(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", w.Code)
	}
}
