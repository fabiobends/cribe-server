package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"cribeapp.com/cribe-server/internal/core/logger"
	"cribeapp.com/cribe-server/internal/utils"
)

// LLMClient interface defines the methods required by services
type LLMClient interface {
	Chat(ctx context.Context, req ChatRequest) (ChatCompletionResponse, error)
}

// NewClient creates a new LLM client
func NewClient() *Client {
	log := logger.NewServiceLogger("LLMClient")

	// Get API key and base URL from environment
	apiKey := os.Getenv("LLM_API_KEY")
	baseURL := os.Getenv("LLM_API_BASE_URL")

	if baseURL == "" || apiKey == "" {
		log.Error("Missing required environment variables for LLM client", map[string]any{
			"has_base_url": baseURL != "",
			"has_api_key":  apiKey != "",
		})
		return nil
	}

	log.Info("LLM client initialized", map[string]any{
		"apiKey":  strings.Split(apiKey, "")[0] + "...",
		"baseURL": baseURL,
	})

	return &Client{
		apiKey:     apiKey,
		baseURL:    baseURL,
		httpClient: &http.Client{},
		log:        log,
	}
}

// setRequestHeaders sets common headers for LLM API requests
func (c *Client) setRequestHeaders(req *http.Request) {
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")
}

// Chat request to LLM client
func (c *Client) Chat(ctx context.Context, req ChatRequest) (ChatCompletionResponse, error) {
	// Build internal request with infrastructure details
	internalReq := chatCompletionRequest{
		Model:       DefaultChatModel,
		Messages:    req.Messages,
		Temperature: DefaultTemperature,
		MaxTokens:   req.MaxTokens,
	}

	jsonBody, err := utils.EncodeToJSON(internalReq)
	if err != nil {
		c.log.Error("Failed to marshal LLM request", map[string]any{
			"error": err.Error(),
		})
		return ChatCompletionResponse{}, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/chat/completions", bytes.NewReader(jsonBody))
	if err != nil {
		c.log.Error("Failed to create LLM request", map[string]any{
			"error": err.Error(),
		})
		return ChatCompletionResponse{}, fmt.Errorf("failed to create request: %w", err)
	}

	c.setRequestHeaders(httpReq)

	// Send request
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		c.log.Error("Failed to send LLM request", map[string]any{
			"error": err.Error(),
		})
		return ChatCompletionResponse{}, fmt.Errorf("failed to send request: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			c.log.Error("Failed to close response body", map[string]any{
				"error": closeErr.Error(),
			})
		}
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		c.log.Error("LLM API returned error", map[string]any{
			"statusCode": resp.StatusCode,
			"response":   string(body),
		})
		return ChatCompletionResponse{}, fmt.Errorf("LLM API error: status=%d, body=%s", resp.StatusCode, string(body))
	}

	// Parse response
	var chatResp ChatCompletionResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		c.log.Error("Failed to decode LLM response", map[string]any{
			"error": err.Error(),
		})
		return ChatCompletionResponse{}, fmt.Errorf("failed to decode response: %w", err)
	}

	return chatResp, nil
}
