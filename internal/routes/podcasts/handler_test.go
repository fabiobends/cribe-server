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
