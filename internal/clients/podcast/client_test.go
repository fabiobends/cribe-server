package podcast

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

// ErrorReader is a mock io.ReadCloser that always returns an error
type ErrorReader struct{}

func (e *ErrorReader) Read(p []byte) (n int, err error) {
	return 0, io.ErrUnexpectedEOF
}

func (e *ErrorReader) Close() error {
	return nil
}

func TestNewClient(t *testing.T) {
	oldUserID := os.Getenv("TADDY_USER_ID")
	oldAPIKey := os.Getenv("TADDY_API_KEY")
	defer func() {
		_ = os.Setenv("TADDY_USER_ID", oldUserID)
		_ = os.Setenv("TADDY_API_KEY", oldAPIKey)
	}()

	_ = os.Setenv("TADDY_USER_ID", "test-user-id")
	_ = os.Setenv("TADDY_API_KEY", "test-api-key")

	client := NewClient()

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
	mockHTTPClient := &MockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			response := GraphQLResponse{
				Data: GraphQLData{
					GetPopularContent: PopularContent{
						PodcastSeries: []ExternalPodcastSeries{
							{
								UUID:        "uuid-1",
								Name:        "Test Podcast",
								AuthorName:  "Author",
								ImageURL:    "http://example.com/image.jpg",
								Description: "Description",
							},
						},
					},
				},
			}

			responseBody, _ := json.Marshal(response)
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewBuffer(responseBody)),
				Header:     make(http.Header),
			}, nil
		},
	}

	client := &Client{
		userID:     "test-user",
		apiKey:     "test-key",
		httpClient: mockHTTPClient,
		baseURL:    "http://test.com",
		logger:     logger.NewServiceLogger("PodcastClientTest"),
	}

	podcasts, err := client.GetTopPodcasts()

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(podcasts) != 1 {
		t.Errorf("Expected 1 podcast, got %d", len(podcasts))
	}

	if podcasts[0].Name != "Test Podcast" {
		t.Errorf("Expected 'Test Podcast', got '%s'", podcasts[0].Name)
	}
}

func TestGetTopPodcasts_HTTPError(t *testing.T) {
	mockHTTPClient := &MockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return nil, http.ErrNotSupported
		},
	}

	client := &Client{
		userID:     "test-user",
		apiKey:     "test-key",
		httpClient: mockHTTPClient,
		baseURL:    "http://test.com",
		logger:     logger.NewServiceLogger("PodcastClientTest"),
	}

	_, err := client.GetTopPodcasts()

	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestGetPodcastByID_Success(t *testing.T) {
	mockHTTPClient := &MockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			type GetPodcastResponse struct {
				Data struct {
					GetPodcastSeries PodcastWithEpisodes `json:"getPodcastSeries"`
				} `json:"data"`
			}

			response := GetPodcastResponse{
				Data: struct {
					GetPodcastSeries PodcastWithEpisodes `json:"getPodcastSeries"`
				}{
					GetPodcastSeries: PodcastWithEpisodes{
						UUID:       "uuid-1",
						Name:       "Test Podcast",
						AuthorName: "Author",
						Episodes: []PodcastEpisode{
							{UUID: "ep-1", Name: "Episode 1"},
						},
					},
				},
			}

			responseBody, _ := json.Marshal(response)
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewBuffer(responseBody)),
				Header:     make(http.Header),
			}, nil
		},
	}

	client := &Client{
		userID:     "test-user",
		apiKey:     "test-key",
		httpClient: mockHTTPClient,
		baseURL:    "http://test.com",
		logger:     logger.NewServiceLogger("PodcastClientTest"),
	}

	podcast, err := client.GetPodcastByID("uuid-1")

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if podcast.Name != "Test Podcast" {
		t.Errorf("Expected 'Test Podcast', got '%s'", podcast.Name)
	}

	if len(podcast.Episodes) != 1 {
		t.Errorf("Expected 1 episode, got %d", len(podcast.Episodes))
	}
}

func TestGetEpisodeByID_Success(t *testing.T) {
	mockHTTPClient := &MockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			type GetEpisodeResponse struct {
				Data struct {
					GetPodcastEpisode PodcastEpisode `json:"getPodcastEpisode"`
				} `json:"data"`
			}

			response := GetEpisodeResponse{
				Data: struct {
					GetPodcastEpisode PodcastEpisode `json:"getPodcastEpisode"`
				}{
					GetPodcastEpisode: PodcastEpisode{
						UUID: "ep-1",
						Name: "Episode 1",
					},
				},
			}

			responseBody, _ := json.Marshal(response)
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewBuffer(responseBody)),
				Header:     make(http.Header),
			}, nil
		},
	}

	client := &Client{
		userID:     "test-user",
		apiKey:     "test-key",
		httpClient: mockHTTPClient,
		baseURL:    "http://test.com",
		logger:     logger.NewServiceLogger("PodcastClientTest"),
	}

	episode, err := client.GetEpisodeByID("podcast-1", "ep-1")

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if episode.Name != "Episode 1" {
		t.Errorf("Expected 'Episode 1', got '%s'", episode.Name)
	}
}

// Error cases tests
func TestGetTopPodcasts_Errors(t *testing.T) {
	tests := []struct {
		name   string
		doFunc func(req *http.Request) (*http.Response, error)
	}{
		{"ReadBodyError", func(req *http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: http.StatusOK, Body: &ErrorReader{}}, nil
		}},
		{"Non200Status", func(req *http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: http.StatusInternalServerError, Body: io.NopCloser(bytes.NewBufferString("error"))}, nil
		}},
		{"DecodeError", func(req *http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(bytes.NewBufferString("invalid"))}, nil
		}},
		{"GraphQLErrors", func(req *http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(bytes.NewBufferString(`{"errors":[{"message":"error"}]}`))}, nil
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &Client{
				userID:     "test-user",
				apiKey:     "test-key",
				httpClient: &MockHTTPClient{DoFunc: tt.doFunc},
				baseURL:    "http://test.com",
				logger:     logger.NewServiceLogger("PodcastClientTest"),
			}
			if _, err := client.GetTopPodcasts(); err == nil {
				t.Fatal("Expected error, got nil")
			}
		})
	}
}

func TestGetPodcastByID_Errors(t *testing.T) {
	tests := []struct {
		name   string
		doFunc func(req *http.Request) (*http.Response, error)
	}{
		{"ReadBodyError", func(req *http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: http.StatusOK, Body: &ErrorReader{}}, nil
		}},
		{"Non200Status", func(req *http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: http.StatusNotFound, Body: io.NopCloser(bytes.NewBufferString("error"))}, nil
		}},
		{"DecodeError", func(req *http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(bytes.NewBufferString("invalid"))}, nil
		}},
		{"GraphQLErrors", func(req *http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(bytes.NewBufferString(`{"errors":[{"message":"error"}]}`))}, nil
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &Client{
				userID:     "test-user",
				apiKey:     "test-key",
				httpClient: &MockHTTPClient{DoFunc: tt.doFunc},
				baseURL:    "http://test.com",
				logger:     logger.NewServiceLogger("PodcastClientTest"),
			}
			if _, err := client.GetPodcastByID("uuid-1"); err == nil {
				t.Fatal("Expected error, got nil")
			}
		})
	}
}

func TestGetEpisodeByID_Errors(t *testing.T) {
	tests := []struct {
		name   string
		doFunc func(req *http.Request) (*http.Response, error)
	}{
		{"ReadBodyError", func(req *http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: http.StatusOK, Body: &ErrorReader{}}, nil
		}},
		{"Non200Status", func(req *http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: http.StatusNotFound, Body: io.NopCloser(bytes.NewBufferString("error"))}, nil
		}},
		{"DecodeError", func(req *http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(bytes.NewBufferString("invalid"))}, nil
		}},
		{"GraphQLErrors", func(req *http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(bytes.NewBufferString(`{"errors":[{"message":"error"}]}`))}, nil
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &Client{
				userID:     "test-user",
				apiKey:     "test-key",
				httpClient: &MockHTTPClient{DoFunc: tt.doFunc},
				baseURL:    "http://test.com",
				logger:     logger.NewServiceLogger("PodcastClientTest"),
			}
			if _, err := client.GetEpisodeByID("podcast-1", "ep-1"); err == nil {
				t.Fatal("Expected error, got nil")
			}
		})
	}
}
