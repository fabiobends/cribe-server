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

// InferSpeakerName uses AI to infer the speaker's name
func (c *Client) InferSpeakerName(ctx context.Context, episodeDescription string, speakerIndex int, transcriptChunks []string) (string, error) {
	// Build context from transcript chunks (includes before/during/after speaker talks)
	chunksText := ""
	maxChunks := min(len(transcriptChunks), 200) // Limit to avoid token limits
	for i := range maxChunks {
		chunksText += transcriptChunks[i] + " "
	}

	// Create prompt for speaker inference
	systemPrompt := "You are an expert at identifying speakers in podcast transcripts. Look for explicit name mentions in the text (e.g., 'this is John', 'I'm Sarah', 'talking with Mike'). Return ONLY the person's name."

	userPrompt := fmt.Sprintf(`Episode description:
%s

Transcript excerpt with context around speaker %d:
%s

Instructions:
- Look for the speaker's name mentioned BEFORE they speak (introductions)
- Look for the speaker's name mentioned WHILE they speak (self-introduction)
- Look for the speaker's name mentioned AFTER they speak (references)
- Common patterns: "I'm [name]", "this is [name]", "with [name]", "[name] said"

Who is speaker %d? Return only their full name (e.g., "John Smith"). If uncertain, return "Speaker %d".`,
		episodeDescription, speakerIndex, chunksText, speakerIndex, speakerIndex)

	// Create request
	reqBody := ChatCompletionRequest{
		Model: "gpt-4o-mini", // Cheapest model
		Messages: []Message{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
		Temperature: 0.3, // Low temperature for more deterministic results
		MaxTokens:   50,  // We only need a name
	}

	jsonBody, err := utils.EncodeToJSON(reqBody)
	if err != nil {
		c.log.Error("Failed to marshal LLM request", map[string]any{
			"error": err.Error(),
		})
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/chat/completions", bytes.NewReader(jsonBody))
	if err != nil {
		c.log.Error("Failed to create LLM request", map[string]any{
			"error": err.Error(),
		})
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	c.log.Info("Inferring speaker name", map[string]any{
		"speakerIndex": speakerIndex,
		"chunksCount":  len(transcriptChunks),
	})

	// Send request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.log.Error("Failed to send inference request", map[string]any{
			"error": err.Error(),
		})
		return "", fmt.Errorf("failed to send request: %w", err)
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
		return "", fmt.Errorf("LLM API error: status=%d, body=%s", resp.StatusCode, string(body))
	}

	// Parse response
	var chatResp ChatCompletionResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		c.log.Error("Failed to decode LLM response", map[string]any{
			"error": err.Error(),
		})
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if len(chatResp.Choices) == 0 {
		c.log.Error("LLM API returned no choices", nil)
		return "", fmt.Errorf("no choices returned from LLM API")
	}

	speakerName := chatResp.Choices[0].Message.Content

	c.log.Info("Speaker name inferred", map[string]any{
		"speakerIndex": speakerIndex,
		"speakerName":  speakerName,
		"tokensUsed":   chatResp.Usage.TotalTokens,
	})

	return speakerName, nil
}
