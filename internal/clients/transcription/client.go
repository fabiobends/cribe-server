package transcription

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

// NewClient creates a new transcription client
func NewClient() *Client {
	log := logger.NewServiceLogger("TranscriptionClient")

	// Get API key and base URL from environment
	apiKey := os.Getenv("TRANSCRIPTION_API_KEY")
	baseURL := os.Getenv("TRANSCRIPTION_API_BASE_URL")

	if baseURL == "" || apiKey == "" {
		log.Error("Missing required environment variables for Transcription client", map[string]interface{}{
			"has_base_url": baseURL != "",
			"has_api_key":  apiKey != "",
		})
		return nil
	}

	log.Info("Transcription client initialized", map[string]interface{}{
		"apiKey":  strings.Split(apiKey, "")[0] + "...",
		"baseURL": baseURL,
	})

	return &Client{
		apiKey:  apiKey,
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		log: log,
	}
}

// StreamAudioURL streams audio from a URL and processes transcription
func (c *Client) StreamAudioURL(ctx context.Context, audioURL string, opts StreamOptions, callback StreamCallback) error {
	// Build query parameters
	queryParams := fmt.Sprintf("model=%s&language=%s&diarize=%t&punctuate=%t&utterances=%t",
		opts.Model, opts.Language, opts.Diarize, opts.Punctuate, opts.Utterances)

	// Create request
	reqURL := fmt.Sprintf("%s/listen?%s", c.baseURL, queryParams)

	reqBody := map[string]string{"url": audioURL}
	jsonBody, err := utils.EncodeToJSON(reqBody)
	if err != nil {
		c.log.Error("Failed to marshal transcription request", map[string]interface{}{
			"error": err.Error(),
		})
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", reqURL, bytes.NewReader(jsonBody))
	if err != nil {
		c.log.Error("Failed to create transcription request", map[string]interface{}{
			"error": err.Error(),
		})
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Token "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	c.log.Info("Streaming audio transcription", map[string]interface{}{
		"audioURL": audioURL,
		"model":    opts.Model,
	})

	// Send request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.log.Error("Failed to send transcription request", map[string]interface{}{
			"error": err.Error(),
		})
		return fmt.Errorf("failed to send request: %w", err)
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
		c.log.Error("Transcription API returned error", map[string]interface{}{
			"statusCode": resp.StatusCode,
			"response":   string(body),
		})
		return fmt.Errorf("transcription API error: status=%d, body=%s", resp.StatusCode, string(body))
	}

	// Process streaming response
	return c.processStreamingResponse(resp.Body, callback)
}

// processStreamingResponse reads and parses the response, then streams it in chunks
func (c *Client) processStreamingResponse(body io.Reader, callback StreamCallback) error {
	// Read the entire response body
	responseBody, err := io.ReadAll(body)
	if err != nil {
		c.log.Error("Failed to read transcription response", map[string]interface{}{
			"error": err.Error(),
		})
		return fmt.Errorf("failed to read response: %w", err)
	}

	// Parse the complete response
	var response StreamResponse
	if err := json.Unmarshal(responseBody, &response); err != nil {
		c.log.Error("Failed to parse transcription response", map[string]interface{}{
			"error":      err.Error(),
			"bodySample": string(responseBody[:min(len(responseBody), 500)]),
		})
		return fmt.Errorf("failed to parse response: %w", err)
	}

	// Check if we have transcript data
	if len(response.Results.Channels) == 0 || len(response.Results.Channels[0].Alternatives) == 0 {
		c.log.Warn("No alternatives in transcription response", nil)
		return nil
	}

	alternative := response.Results.Channels[0].Alternatives[0]
	words := alternative.Words

	c.log.Info("Processing transcription words", map[string]interface{}{
		"totalWords": len(words),
	})

	// Stream words individually for better highlighting granularity
	for i, word := range words {
		isLastWord := i == len(words)-1

		// Create a response with single word
		wordResponse := StreamResponse{
			IsFinal:  isLastWord,
			Type:     "Results",
			Metadata: response.Metadata,
			Results: Results{
				Channels: []Channel{
					{
						Alternatives: []Alternative{
							{
								Transcript: word.PunctuatedWord,
								Confidence: alternative.Confidence,
								Words:      []Word{word},
							},
						},
					},
				},
			},
		}

		if err := callback(&wordResponse); err != nil {
			c.log.Error("Callback error processing word", map[string]interface{}{
				"error":     err.Error(),
				"wordIndex": i + 1,
				"word":      word.PunctuatedWord,
			})
			return err
		}
	}

	c.log.Info("Transcription stream completed successfully", map[string]interface{}{
		"totalWords": len(words),
	})
	return nil
}
