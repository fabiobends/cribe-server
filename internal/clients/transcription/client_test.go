package transcription

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"cribeapp.com/cribe-server/internal/core/logger"
)

func TestNewClient(t *testing.T) {
	t.Run("should create client with valid environment variables", func(t *testing.T) {
		_ = os.Setenv("TRANSCRIPTION_API_KEY", "test-key")
		_ = os.Setenv("TRANSCRIPTION_API_BASE_URL", "https://api.deepgram.com")
		defer func() { _ = os.Unsetenv("TRANSCRIPTION_API_KEY") }()
		defer func() { _ = os.Unsetenv("TRANSCRIPTION_API_BASE_URL") }()

		client := NewClient()

		if client == nil {
			t.Fatal("Expected client to be created, got nil")
		}

		if client.apiKey != "test-key" {
			t.Errorf("Expected apiKey to be 'test-key', got %s", client.apiKey)
		}

		if client.baseURL != "https://api.deepgram.com" {
			t.Errorf("Expected baseURL to be 'https://api.deepgram.com', got %s", client.baseURL)
		}
	})

	t.Run("should return nil when API key is missing", func(t *testing.T) {
		_ = os.Unsetenv("TRANSCRIPTION_API_KEY")
		_ = os.Setenv("TRANSCRIPTION_API_BASE_URL", "https://api.deepgram.com")
		defer func() { _ = os.Unsetenv("TRANSCRIPTION_API_BASE_URL") }()

		client := NewClient()

		if client != nil {
			t.Error("Expected client to be nil when API key is missing")
		}
	})

	t.Run("should return nil when base URL is missing", func(t *testing.T) {
		_ = os.Setenv("TRANSCRIPTION_API_KEY", "test-key")
		_ = os.Unsetenv("TRANSCRIPTION_API_BASE_URL")
		defer func() { _ = os.Unsetenv("TRANSCRIPTION_API_KEY") }()

		client := NewClient()

		if client != nil {
			t.Error("Expected client to be nil when base URL is missing")
		}
	})
}

func TestStreamAudioURL(t *testing.T) {
	t.Run("should stream audio transcription successfully", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("Authorization") != "Token test-key" {
				t.Errorf("Expected Authorization header 'Token test-key', got %s", r.Header.Get("Authorization"))
			}

			response := StreamResponse{
				Type: "Results",
				Results: Results{
					Channels: []Channel{
						{
							Alternatives: []Alternative{
								{
									Transcript: "Hello world",
									Confidence: 0.95,
									Words: []Word{
										{
											Word:           "Hello",
											PunctuatedWord: "Hello",
											Start:          0.0,
											End:            0.5,
											Speaker:        0,
										},
										{
											Word:           "world",
											PunctuatedWord: "world",
											Start:          0.5,
											End:            1.0,
											Speaker:        0,
										},
									},
								},
							},
						},
					},
				},
			}

			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		client := &Client{
			apiKey:     "test-key",
			baseURL:    server.URL,
			httpClient: &http.Client{},
			log:        logger.NewServiceLogger("TestTranscriptionClient"),
		}

		callbackCount := 0
		err := client.StreamAudioURL(
			context.Background(),
			"https://example.com/audio.mp3",
			StreamOptions{
				Model:      "nova-3",
				Language:   "en",
				Diarize:    true,
				Punctuate:  true,
				Utterances: true,
			},
			func(response *StreamResponse) error {
				callbackCount++
				return nil
			},
		)

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if callbackCount != 2 {
			t.Errorf("Expected callback to be called 2 times, got %d", callbackCount)
		}
	})

	t.Run("should return error when API returns non-200 status", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("Bad request"))
		}))
		defer server.Close()

		client := &Client{
			apiKey:     "test-key",
			baseURL:    server.URL,
			httpClient: &http.Client{},
			log:        logger.NewServiceLogger("TestTranscriptionClient"),
		}

		err := client.StreamAudioURL(
			context.Background(),
			"https://example.com/audio.mp3",
			StreamOptions{Model: "nova-3"},
			func(response *StreamResponse) error {
				return nil
			},
		)

		if err == nil {
			t.Error("Expected error, got nil")
		}
	})

	t.Run("should handle empty response gracefully", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			response := StreamResponse{
				Type: "Results",
				Results: Results{
					Channels: []Channel{},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		client := &Client{
			apiKey:     "test-key",
			baseURL:    server.URL,
			httpClient: &http.Client{},
			log:        logger.NewServiceLogger("TestTranscriptionClient"),
		}

		err := client.StreamAudioURL(
			context.Background(),
			"https://example.com/audio.mp3",
			StreamOptions{Model: "nova-3"},
			func(response *StreamResponse) error {
				return nil
			},
		)

		if err != nil {
			t.Errorf("Expected no error for empty response, got %v", err)
		}
	})
}

func TestStreamAudioURL_ErrorPaths(t *testing.T) {
	t.Run("should handle HTTP client error", func(t *testing.T) {
		client := &Client{
			apiKey:     "test-key",
			baseURL:    "http://invalid-url-that-does-not-exist.local",
			httpClient: &http.Client{},
			log:        logger.NewServiceLogger("TestTranscriptionClient"),
		}

		err := client.StreamAudioURL(
			context.Background(),
			"https://example.com/audio.mp3",
			StreamOptions{Model: "nova-3"},
			func(response *StreamResponse) error {
				return nil
			},
		)

		if err == nil {
			t.Error("Expected error for failed HTTP request, got nil")
		}
	})

	t.Run("should handle invalid JSON response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte("invalid json"))
		}))
		defer server.Close()

		client := &Client{
			apiKey:     "test-key",
			baseURL:    server.URL,
			httpClient: &http.Client{},
			log:        logger.NewServiceLogger("TestTranscriptionClient"),
		}

		err := client.StreamAudioURL(
			context.Background(),
			"https://example.com/audio.mp3",
			StreamOptions{Model: "nova-3"},
			func(response *StreamResponse) error {
				return nil
			},
		)

		if err == nil {
			t.Error("Expected error for invalid JSON, got nil")
		}
	})

	t.Run("should handle callback error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			response := StreamResponse{
				Type: "Results",
				Results: Results{
					Channels: []Channel{
						{
							Alternatives: []Alternative{
								{
									Words: []Word{
										{
											Word:           "test",
											PunctuatedWord: "test",
											Start:          0.0,
											End:            0.5,
											Speaker:        0,
										},
									},
								},
							},
						},
					},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		client := &Client{
			apiKey:     "test-key",
			baseURL:    server.URL,
			httpClient: &http.Client{},
			log:        logger.NewServiceLogger("TestTranscriptionClient"),
		}

		err := client.StreamAudioURL(
			context.Background(),
			"https://example.com/audio.mp3",
			StreamOptions{Model: "nova-3"},
			func(response *StreamResponse) error {
				return json.Unmarshal([]byte("invalid"), &response)
			},
		)

		if err == nil {
			t.Error("Expected callback error, got nil")
		}

	})
}
