package podcasts

import (
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
