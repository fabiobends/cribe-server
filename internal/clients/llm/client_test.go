package llm

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
		_ = os.Setenv("LLM_API_KEY", "test-key")
		_ = os.Setenv("LLM_API_BASE_URL", "https://api.example.com")
		defer func() { _ = os.Unsetenv("LLM_API_KEY") }()
		defer func() { _ = os.Unsetenv("LLM_API_BASE_URL") }()

		client := NewClient()

		if client == nil {
			t.Fatal("Expected client to be created, got nil")
		}

		if client.apiKey != "test-key" {
			t.Errorf("Expected apiKey to be 'test-key', got %s", client.apiKey)
		}

		if client.baseURL != "https://api.example.com" {
			t.Errorf("Expected baseURL to be 'https://api.example.com', got %s", client.baseURL)
		}
	})

	t.Run("should return nil when API key is missing", func(t *testing.T) {
		_ = os.Unsetenv("LLM_API_KEY")
		_ = os.Setenv("LLM_API_BASE_URL", "https://api.example.com")
		defer func() { _ = os.Unsetenv("LLM_API_BASE_URL") }()

		client := NewClient()

		if client != nil {
			t.Error("Expected client to be nil when API key is missing")
		}
	})

	t.Run("should return nil when base URL is missing", func(t *testing.T) {
		_ = os.Setenv("LLM_API_KEY", "test-key")
		_ = os.Unsetenv("LLM_API_BASE_URL")
		defer func() { _ = os.Unsetenv("LLM_API_KEY") }()

		client := NewClient()

		if client != nil {
			t.Error("Expected client to be nil when base URL is missing")
		}
	})
}

func TestInferSpeakerName(t *testing.T) {
	t.Run("should infer speaker name successfully", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/chat/completions" {
				t.Errorf("Expected path '/chat/completions', got %s", r.URL.Path)
			}

			if r.Header.Get("Authorization") != "Bearer test-key" {
				t.Errorf("Expected Authorization header 'Bearer test-key', got %s", r.Header.Get("Authorization"))
			}

			response := ChatCompletionResponse{
				Choices: []Choice{
					{
						Message: Message{
							Content: "John Doe",
						},
					},
				},
				Usage: Usage{
					TotalTokens: 50,
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
			log:        logger.NewServiceLogger("TestLLMClient"),
		}

		name, err := client.InferSpeakerName(
			context.Background(),
			"Test episode about technology",
			0,
			[]string{"Hello world", "This is a test"},
		)

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if name != "John Doe" {
			t.Errorf("Expected 'John Doe', got %s", name)
		}
	})

	t.Run("should return error when API returns non-200 status", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("Internal server error"))
		}))
		defer server.Close()

		client := &Client{
			apiKey:     "test-key",
			baseURL:    server.URL,
			httpClient: &http.Client{},
			log:        logger.NewServiceLogger("TestLLMClient"),
		}

		_, err := client.InferSpeakerName(
			context.Background(),
			"Test episode",
			0,
			[]string{"Hello"},
		)

		if err == nil {
			t.Error("Expected error, got nil")
		}
	})

	t.Run("should return error when response has no choices", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			response := ChatCompletionResponse{
				Choices: []Choice{},
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		client := &Client{
			apiKey:     "test-key",
			baseURL:    server.URL,
			httpClient: &http.Client{},
			log:        logger.NewServiceLogger("TestLLMClient"),
		}

		_, err := client.InferSpeakerName(
			context.Background(),
			"Test episode",
			0,
			[]string{"Hello"},
		)

		if err == nil {
			t.Error("Expected error, got nil")
		}
	})
}

func TestInferSpeakerName_ErrorPaths(t *testing.T) {
	t.Run("should handle HTTP client error", func(t *testing.T) {
		client := &Client{
			apiKey:     "test-key",
			baseURL:    "http://invalid-url-that-does-not-exist.local",
			httpClient: &http.Client{},
			log:        logger.NewServiceLogger("TestLLMClient"),
		}

		_, err := client.InferSpeakerName(
			context.Background(),
			"Test episode",
			0,
			[]string{"Hello"},
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
			log:        logger.NewServiceLogger("TestLLMClient"),
		}

		_, err := client.InferSpeakerName(
			context.Background(),
			"Test episode",
			0,
			[]string{"Hello"},
		)

		if err == nil {
			t.Error("Expected error for invalid JSON, got nil")
		}
	})
}
