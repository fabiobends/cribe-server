package podcasts

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"testing"

	"cribeapp.com/cribe-server/internal/core/logger"
)

// MockHTTPClient is a mock implementation of HTTPClient for testing
type MockHTTPClient struct {
	DoFunc func(req *http.Request) (*http.Response, error)
}

func (m *MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	if m.DoFunc != nil {
		return m.DoFunc(req)
	}
	return nil, nil
}

func TestNewPodcastAPIClient(t *testing.T) {
	oldUserID := os.Getenv("TADDY_USER_ID")
	oldAPIKey := os.Getenv("TADDY_API_KEY")
	defer func() {
		_ = os.Setenv("TADDY_USER_ID", oldUserID)
		_ = os.Setenv("TADDY_API_KEY", oldAPIKey)
	}()

	_ = os.Setenv("TADDY_USER_ID", "test-user-id")
	_ = os.Setenv("TADDY_API_KEY", "test-api-key")

	client := NewPodcastAPIClient()

	if client.userID != "test-user-id" {
		t.Errorf("Expected userID 'test-user-id', got '%s'", client.userID)
	}

	if client.apiKey != "test-api-key" {
		t.Errorf("Expected apiKey 'test-api-key', got '%s'", client.apiKey)
	}

	if client.logger == nil {
		t.Error("Expected logger to be initialized")
	}
}

func TestGetTopPodcasts_Success(t *testing.T) {
	// Create a mock HTTP client
	mockHTTPClient := &MockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			// Verify headers
			if req.Header.Get("X-USER-ID") != "test-user" {
				t.Errorf("Expected X-USER-ID 'test-user', got '%s'", req.Header.Get("X-USER-ID"))
			}
			if req.Header.Get("X-API-KEY") != "test-key" {
				t.Errorf("Expected X-API-KEY 'test-key', got '%s'", req.Header.Get("X-API-KEY"))
			}
			if req.Header.Get("Content-Type") != "application/json" {
				t.Errorf("Expected Content-Type 'application/json', got '%s'", req.Header.Get("Content-Type"))
			}

			// Return mock response using actual structs for type safety
			response := GraphQLResponse{
				Data: GraphQLData{
					GetPopularContent: PopularContent{
						PodcastSeries: []ExternalPodcastSeries{
							{
								UUID:        "uuid-1",
								Name:        "Test Podcast 1",
								AuthorName:  "Author 1",
								ImageURL:    "http://example.com/image1.jpg",
								Description: "Description 1",
							},
							{
								UUID:        "uuid-2",
								Name:        "Test Podcast 2",
								AuthorName:  "Author 2",
								ImageURL:    "http://example.com/image2.jpg",
								Description: "Description 2",
							},
						},
					},
				},
			}

			responseBody, err := json.Marshal(response)
			if err != nil {
				t.Fatalf("Failed to marshal mock response: %v", err)
			}

			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewBuffer(responseBody)),
				Header:     make(http.Header),
			}, nil
		},
	}

	// Create client with mock HTTP client
	client := &PodcastAPIClient{
		userID:     "test-user",
		apiKey:     "test-key",
		logger:     logger.NewServiceLogger("PodcastAPIClient"),
		httpClient: mockHTTPClient,
		baseURL:    "http://mock-api.com",
	}

	podcasts, err := client.GetTopPodcasts()

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(podcasts) != 2 {
		t.Errorf("Expected 2 podcasts, got %d", len(podcasts))
	}

	if podcasts[0].UUID != "uuid-1" {
		t.Errorf("Expected UUID 'uuid-1', got '%s'", podcasts[0].UUID)
	}

	if podcasts[0].Name != "Test Podcast 1" {
		t.Errorf("Expected name 'Test Podcast 1', got '%s'", podcasts[0].Name)
	}
}

func TestGetTopPodcasts_HTTPError(t *testing.T) {
	// Create a mock HTTP client that returns an error status
	mockHTTPClient := &MockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusInternalServerError,
				Body:       io.NopCloser(bytes.NewBufferString("Internal Server Error")),
				Header:     make(http.Header),
			}, nil
		},
	}

	client := &PodcastAPIClient{
		userID:     "test-user",
		apiKey:     "test-key",
		logger:     logger.NewServiceLogger("PodcastAPIClient"),
		httpClient: mockHTTPClient,
		baseURL:    "http://mock-api.com",
	}

	podcasts, err := client.GetTopPodcasts()

	if err == nil {
		t.Error("Expected error for HTTP 500, got nil")
	}

	if len(podcasts) != 0 {
		t.Errorf("Expected 0 podcasts on error, got %d", len(podcasts))
	}
}

func TestGetTopPodcasts_GraphQLError(t *testing.T) {
	// Create a mock HTTP client that returns GraphQL errors
	mockHTTPClient := &MockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			// Return mock response using actual structs for type safety
			response := GraphQLResponse{
				Errors: []map[string]interface{}{
					{"message": "GraphQL error occurred"},
				},
			}

			responseBody, err := json.Marshal(response)
			if err != nil {
				t.Fatalf("Failed to marshal mock response: %v", err)
			}

			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewBuffer(responseBody)),
				Header:     make(http.Header),
			}, nil
		},
	}

	client := &PodcastAPIClient{
		userID:     "test-user",
		apiKey:     "test-key",
		logger:     logger.NewServiceLogger("PodcastAPIClient"),
		httpClient: mockHTTPClient,
		baseURL:    "http://mock-api.com",
	}

	podcasts, err := client.GetTopPodcasts()

	if err == nil {
		t.Error("Expected error for GraphQL errors, got nil")
	}

	if len(podcasts) != 0 {
		t.Errorf("Expected 0 podcasts on error, got %d", len(podcasts))
	}
}

func TestGetTopPodcasts_InvalidJSON(t *testing.T) {
	// Create a mock HTTP client that returns invalid JSON
	mockHTTPClient := &MockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewBufferString("invalid json")),
				Header:     make(http.Header),
			}, nil
		},
	}

	client := &PodcastAPIClient{
		userID:     "test-user",
		apiKey:     "test-key",
		logger:     logger.NewServiceLogger("PodcastAPIClient"),
		httpClient: mockHTTPClient,
		baseURL:    "http://mock-api.com",
	}

	podcasts, err := client.GetTopPodcasts()

	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}

	if len(podcasts) != 0 {
		t.Errorf("Expected 0 podcasts on error, got %d", len(podcasts))
	}
}

func TestGetTopPodcasts_NetworkError(t *testing.T) {
	// Create a mock HTTP client that returns a network error
	mockHTTPClient := &MockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return nil, io.EOF
		},
	}

	client := &PodcastAPIClient{
		userID:     "test-user",
		apiKey:     "test-key",
		logger:     logger.NewServiceLogger("PodcastAPIClient"),
		httpClient: mockHTTPClient,
		baseURL:    "http://mock-api.com",
	}

	podcasts, err := client.GetTopPodcasts()

	if err == nil {
		t.Error("Expected error for network failure, got nil")
	}

	if len(podcasts) != 0 {
		t.Errorf("Expected 0 podcasts on error, got %d", len(podcasts))
	}
}
