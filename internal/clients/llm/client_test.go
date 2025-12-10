package llm

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

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

func TestChat(t *testing.T) {
	t.Run("should send chat request successfully with correct defaults", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/chat/completions" {
				t.Errorf("Expected path '/chat/completions', got %s", r.URL.Path)
			}

			if r.Header.Get("Authorization") != "Bearer test-key" {
				t.Errorf("Expected Authorization header 'Bearer test-key', got %s", r.Header.Get("Authorization"))
			}

			// Verify request body contains correct defaults
			var reqBody chatCompletionRequest
			if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
				t.Errorf("Failed to decode request body: %v", err)
			}

			if reqBody.Model != DefaultChatModel {
				t.Errorf("Expected model %s, got %s", DefaultChatModel, reqBody.Model)
			}

			if reqBody.Temperature != DefaultTemperature {
				t.Errorf("Expected temperature %f, got %f", DefaultTemperature, reqBody.Temperature)
			}

			if reqBody.MaxTokens != 150 {
				t.Errorf("Expected MaxTokens 150, got %d", reqBody.MaxTokens)
			}

			if len(reqBody.Messages) != 1 || reqBody.Messages[0].Content != "Hello" {
				t.Error("Expected messages to be passed through correctly")
			}

			response := ChatCompletionResponse{
				Choices: []Choice{
					{
						Message: Message{
							Content: "This is the AI response",
						},
					},
				},
				Usage: Usage{
					TotalTokens: 100,
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

		request := ChatRequest{
			Messages: []Message{
				{Role: "user", Content: "Hello"},
			},
			MaxTokens: 150,
		}

		response, err := client.Chat(context.Background(), request)

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if len(response.Choices) == 0 {
			t.Error("Expected at least one choice in response")
		}

		if response.Choices[0].Message.Content != "This is the AI response" {
			t.Errorf("Expected 'This is the AI response', got %s", response.Choices[0].Message.Content)
		}

		if response.Usage.TotalTokens != 100 {
			t.Errorf("Expected 100 tokens used, got %d", response.Usage.TotalTokens)
		}
	})

	t.Run("should return error when API returns non-200 status", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"error": "Invalid request"}`))
		}))
		defer server.Close()

		client := &Client{
			apiKey:     "test-key",
			baseURL:    server.URL,
			httpClient: &http.Client{},
			log:        logger.NewServiceLogger("TestLLMClient"),
		}

		request := ChatRequest{
			Messages:  []Message{{Role: "user", Content: "Hello"}},
			MaxTokens: 150,
		}

		_, err := client.Chat(context.Background(), request)

		if err == nil {
			t.Error("Expected error, got nil")
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

		request := ChatRequest{
			Messages:  []Message{{Role: "user", Content: "Hello"}},
			MaxTokens: 150,
		}

		_, err := client.Chat(context.Background(), request)

		if err == nil {
			t.Error("Expected error for invalid JSON, got nil")
		}
	})

	t.Run("should handle HTTP client error", func(t *testing.T) {
		client := &Client{
			apiKey:     "test-key",
			baseURL:    "http://invalid-url-that-does-not-exist.local",
			httpClient: &http.Client{},
			log:        logger.NewServiceLogger("TestLLMClient"),
		}

		request := ChatRequest{
			Messages:  []Message{{Role: "user", Content: "Hello"}},
			MaxTokens: 150,
		}

		_, err := client.Chat(context.Background(), request)

		if err == nil {
			t.Error("Expected error for failed HTTP request, got nil")
		}
	})

	t.Run("should handle context cancellation", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Simulate slow response
			time.Sleep(100 * time.Millisecond)
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		client := &Client{
			apiKey:     "test-key",
			baseURL:    server.URL,
			httpClient: &http.Client{},
			log:        logger.NewServiceLogger("TestLLMClient"),
		}

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		request := ChatRequest{
			Messages:  []Message{{Role: "user", Content: "Hello"}},
			MaxTokens: 150,
		}

		_, err := client.Chat(ctx, request)

		if err == nil {
			t.Error("Expected error for cancelled context, got nil")
		}
	})

	t.Run("should handle response with empty choices", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			response := ChatCompletionResponse{
				Choices: []Choice{},
				Usage: Usage{
					TotalTokens: 0,
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

		request := ChatRequest{
			Messages:  []Message{{Role: "user", Content: "Hello"}},
			MaxTokens: 150,
		}

		response, err := client.Chat(context.Background(), request)

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if len(response.Choices) != 0 {
			t.Error("Expected empty choices array")
		}
	})

	t.Run("should handle multiple response choices", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			response := ChatCompletionResponse{
				Choices: []Choice{
					{Message: Message{Content: "Response 1"}},
					{Message: Message{Content: "Response 2"}},
				},
				Usage: Usage{
					TotalTokens: 200,
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

		request := ChatRequest{
			Messages:  []Message{{Role: "user", Content: "Hello"}},
			MaxTokens: 150,
		}

		response, err := client.Chat(context.Background(), request)

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if len(response.Choices) != 2 {
			t.Errorf("Expected 2 choices, got %d", len(response.Choices))
		}

		if response.Choices[0].Message.Content != "Response 1" {
			t.Errorf("Expected 'Response 1', got %s", response.Choices[0].Message.Content)
		}

		if response.Choices[1].Message.Content != "Response 2" {
			t.Errorf("Expected 'Response 2', got %s", response.Choices[1].Message.Content)
		}
	})
}
