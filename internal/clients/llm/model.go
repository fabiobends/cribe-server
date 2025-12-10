package llm

import (
	"net/http"

	"cribeapp.com/cribe-server/internal/core/logger"
)

const (
	// Default models for different use cases
	DefaultChatModel   = "gpt-4o-mini" // For general chat/inference
	DefaultTemperature = 0.7           // Default creativity level
)

// Client handles LLM API interactions
type Client struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
	log        *logger.ContextualLogger
}

// Message represents a chat message
type Message struct {
	Role    string `json:"role"` // "system", "user", "assistant"
	Content string `json:"content"`
}

// ChatRequest is the service-level request (infrastructure-agnostic)
type ChatRequest struct {
	Messages  []Message `json:"messages"`
	MaxTokens int       `json:"max_tokens,omitempty"`
}

// chatCompletionRequest is the internal OpenAI-specific request
type chatCompletionRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Temperature float64   `json:"temperature,omitempty"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
}

// Choice represents a completion choice
type Choice struct {
	Index        int     `json:"index"`
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason"`
}

// Usage represents token usage
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// ChatCompletionResponse represents the LLM response
type ChatCompletionResponse struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
	Usage   Usage    `json:"usage"`
}
