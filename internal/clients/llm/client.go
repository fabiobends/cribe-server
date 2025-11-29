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
	"time"

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
		log.Error("Missing required environment variables for LLM client", map[string]interface{}{
			"has_base_url": baseURL != "",
			"has_api_key":  apiKey != "",
		})
		return nil
	}

	log.Info("LLM client initialized", map[string]interface{}{
		"apiKey":  strings.Split(apiKey, "")[0] + "...",
		"baseURL": baseURL,
	})

	return &Client{
		apiKey:  apiKey,
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
		log: log,
	}
}

// InferSpeakerName uses AI to infer the speaker's name
func (c *Client) InferSpeakerName(ctx context.Context, episodeDescription string, speakerIndex int, transcriptChunks []string) (string, error) {
	// Build context from transcript chunks
	chunksText := ""
	for i, chunk := range transcriptChunks {
		chunksText += fmt.Sprintf("\n[Segment %d] %s", i+1, chunk)
		if i >= 9 { // Limit to first 10 chunks to save tokens
			break
		}
	}

	// Create prompt for speaker inference
	systemPrompt := "You are a helpful assistant that identifies speakers in podcast transcripts. Return ONLY the person's name, nothing else."

	userPrompt := fmt.Sprintf(`Given this podcast episode description:
%s

And these transcript segments from speaker %d:
%s

Who is speaker %d? Return only the person's full name (e.g., "John Smith" or "Jane Doe"). If you cannot determine the name, return "Speaker %d".`,
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
		c.log.Error("Failed to marshal LLM request", map[string]interface{}{
			"error": err.Error(),
		})
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/chat/completions", bytes.NewReader(jsonBody))
	if err != nil {
		c.log.Error("Failed to create LLM request", map[string]interface{}{
			"error": err.Error(),
		})
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	c.log.Info("Inferring speaker name", map[string]interface{}{
		"speakerIndex": speakerIndex,
		"chunksCount":  len(transcriptChunks),
	})

	// Send request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.log.Error("Failed to send inference request", map[string]interface{}{
			"error": err.Error(),
		})
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			c.log.Error("Failed to close response body", map[string]interface{}{
				"error": closeErr.Error(),
			})
		}
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		c.log.Error("LLM API returned error", map[string]interface{}{
			"statusCode": resp.StatusCode,
			"response":   string(body),
		})
		return "", fmt.Errorf("LLM API error: status=%d, body=%s", resp.StatusCode, string(body))
	}

	// Parse response
	var chatResp ChatCompletionResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		c.log.Error("Failed to decode LLM response", map[string]interface{}{
			"error": err.Error(),
		})
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if len(chatResp.Choices) == 0 {
		c.log.Error("LLM API returned no choices", nil)
		return "", fmt.Errorf("no choices returned from LLM API")
	}

	speakerName := chatResp.Choices[0].Message.Content

	c.log.Info("Speaker name inferred", map[string]interface{}{
		"speakerIndex": speakerIndex,
		"speakerName":  speakerName,
		"tokensUsed":   chatResp.Usage.TotalTokens,
	})

	return speakerName, nil
}
