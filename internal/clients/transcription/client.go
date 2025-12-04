package transcription

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/gorilla/websocket"

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
		log.Error("Missing required environment variables for Transcription client", map[string]any{
			"has_base_url": baseURL != "",
			"has_api_key":  apiKey != "",
		})
		return nil
	}

	log.Info("Transcription client initialized", map[string]any{
		"apiKey":  strings.Split(apiKey, "")[0] + "...",
		"baseURL": baseURL,
	})

	return &Client{
		apiKey:  apiKey,
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Minute,
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
		c.log.Error("Failed to marshal transcription request", map[string]any{
			"error": err.Error(),
		})
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", reqURL, bytes.NewReader(jsonBody))
	if err != nil {
		c.log.Error("Failed to create transcription request", map[string]any{
			"error": err.Error(),
		})
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Token "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	c.log.Info("Streaming audio transcription", map[string]any{
		"audioURL": audioURL,
		"model":    opts.Model,
	})

	// Send request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.log.Error("Failed to send transcription request", map[string]any{
			"error": err.Error(),
		})
		return fmt.Errorf("failed to send request: %w", err)
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
		c.log.Error("Transcription API returned error", map[string]any{
			"statusCode": resp.StatusCode,
			"response":   string(body),
		})
		return fmt.Errorf("transcription API error: status=%d, body=%s", resp.StatusCode, string(body))
	}

	// Process streaming response
	return c.processStreamingResponse(resp.Body, callback)
}

// StreamAudioURLWebSocket streams audio from a URL to Deepgram's WebSocket endpoint
// for real-time transcription. This enables progressive transcription of long-form
// audio (e.g., 2+ hour podcasts) by downloading and streaming audio chunks concurrently
// with receiving transcription results.
func (c *Client) StreamAudioURLWebSocket(ctx context.Context, audioURL string, opts StreamOptions, callback StreamCallback) error {
	// Build WebSocket URL with query parameters
	wsURL := strings.Replace(c.baseURL, "https://", "wss://", 1)
	wsURL = strings.Replace(wsURL, "http://", "ws://", 1)

	u, err := url.Parse(wsURL + "/listen")
	if err != nil {
		return fmt.Errorf("failed to parse WebSocket URL: %w", err)
	}

	q := u.Query()
	q.Set("model", opts.Model)
	q.Set("language", opts.Language)
	q.Set("diarize", fmt.Sprintf("%t", opts.Diarize))
	q.Set("punctuate", fmt.Sprintf("%t", opts.Punctuate))
	q.Set("utterances", fmt.Sprintf("%t", opts.Utterances))
	q.Set("interim_results", "false") // Only get final results
	u.RawQuery = q.Encode()

	// Create WebSocket connection
	headers := http.Header{}
	headers.Set("Authorization", "Token "+c.apiKey)

	c.log.Info("Connecting to Deepgram WebSocket", map[string]any{
		"url":      u.String(),
		"audioURL": audioURL,
	})

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), headers)
	if err != nil {
		c.log.Error("Failed to connect to WebSocket", map[string]any{
			"error": err.Error(),
		})
		return fmt.Errorf("failed to connect to WebSocket: %w", err)
	}
	// Note: Connection will be closed by the reader goroutine when done

	// Create context for managing goroutines - cancel will be called by reader goroutine on completion
	streamCtx, cancel := context.WithCancel(ctx)

	// Channel for errors and completion from goroutines
	errCh := make(chan error, 2)
	doneCh := make(chan struct{}, 2) // Signal successful completion

	// Goroutine 1: Download and stream audio to WebSocket
	go func() {
		// Download audio from URL
		req, err := http.NewRequestWithContext(streamCtx, "GET", audioURL, nil)
		if err != nil {
			errCh <- fmt.Errorf("failed to create audio download request: %w", err)
			return
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			errCh <- fmt.Errorf("failed to download audio: %w", err)
			return
		}
		defer func() {
			if err := resp.Body.Close(); err != nil {
				c.log.Error("Failed to close audio response body", map[string]any{
					"error": err.Error(),
				})
			}
		}()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			errCh <- fmt.Errorf("audio download failed: status=%d, body=%s", resp.StatusCode, string(body))
			return
		}

		c.log.Info("Started streaming audio to WebSocket", map[string]any{
			"audioURL": audioURL,
		})

		// Stream audio in chunks with throttling to simulate real-time
		// For live streaming, we need to pace the audio upload to match its duration
		reader := bufio.NewReader(resp.Body)
		buffer := make([]byte, 8192) // 8KB chunks
		totalBytesSent := 0
		startTime := time.Now()

		// Estimate bytes per second based on typical podcast bitrate
		// MP3 at 128kbps = 16KB/s, we'll use a conservative 20KB/s
		bytesPerSecond := 20000
		targetDuration := time.Second

		for {
			select {
			case <-streamCtx.Done():
				return
			default:
			}

			n, err := reader.Read(buffer)
			if n > 0 {
				if err := conn.WriteMessage(websocket.BinaryMessage, buffer[:n]); err != nil {
					errCh <- fmt.Errorf("failed to send audio chunk: %w", err)
					return
				}
				totalBytesSent += n

				// Throttle to simulate real-time streaming
				// Calculate how long this chunk should take based on bytes sent
				expectedElapsed := time.Duration(totalBytesSent/bytesPerSecond) * time.Second
				actualElapsed := time.Since(startTime)
				if expectedElapsed > actualElapsed {
					sleepDuration := expectedElapsed - actualElapsed
					// Cap sleep to avoid long pauses
					if sleepDuration > targetDuration {
						sleepDuration = targetDuration
					}
					time.Sleep(sleepDuration)
				}
			}

			if err != nil {
				if err == io.EOF {
					c.log.Info("Finished streaming audio to WebSocket", map[string]any{
						"totalBytesSent": totalBytesSent,
						"duration":       time.Since(startTime).String(),
					})
					// Send a close message to signal end of audio
					closeMsg := []byte("{\"type\":\"CloseStream\"}")
					if err := conn.WriteMessage(websocket.TextMessage, closeMsg); err != nil {
						c.log.Error("Failed to send close stream message", map[string]any{
							"error": err.Error(),
						})
					}
					doneCh <- struct{}{} // Signal audio streaming completed
					return
				}
				errCh <- fmt.Errorf("failed to read audio: %w", err)
				return
			}
		}
	}()

	// Goroutine 2: Read transcription results from WebSocket and save when complete
	go func() {
		defer func() {
			cancel() // Cancel context to stop audio goroutine if still running
			if err := conn.Close(); err != nil {
				c.log.Error("Failed to close WebSocket in reader goroutine", map[string]any{
					"error": err.Error(),
				})
			}
		}()

		messageCount := 0
		for {
			select {
			case <-streamCtx.Done():
				c.log.Info("WebSocket reader stopping due to context cancellation", map[string]any{
					"messagesReceived": messageCount,
				})
				return
			default:
			}

			_, message, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
					c.log.Info("WebSocket closed normally", map[string]any{
						"messagesReceived": messageCount,
					})
					doneCh <- struct{}{} // Signal WebSocket reading completed
					return
				}
				errCh <- fmt.Errorf("failed to read WebSocket message: %w", err)
				return
			}

			messageCount++
			if messageCount%10 == 0 {
				c.log.Info("WebSocket progress", map[string]any{
					"messagesReceived": messageCount,
				})
			}

			// Parse JSON message
			var resp StreamResponse
			if err := json.Unmarshal(message, &resp); err != nil {
				c.log.Error("Failed to unmarshal WebSocket message", map[string]any{
					"error":   err.Error(),
					"message": string(message),
				})
				continue
			}

			// Check for metadata or other message types
			if resp.Type != "" && resp.Type != "Results" {
				continue
			}

			// Handle both REST API (channels array) and WebSocket API (singular channel)
			var alternatives []Alternative
			if len(resp.Results.Channels) > 0 {
				// REST API uses "channels" array
				alternatives = resp.Results.Channels[0].Alternatives
			} else if len(resp.Channel.Alternatives) > 0 {
				// WebSocket API uses singular "channel"
				alternatives = resp.Channel.Alternatives
			}

			if len(alternatives) == 0 {
				continue
			}

			// Only process final results
			if !resp.IsFinal {
				continue
			}

			// Invoke callback with the decoded response
			if err := callback(&resp); err != nil {
				c.log.Error("Callback error processing WebSocket response", map[string]any{
					"error": err.Error(),
				})
				errCh <- err
				return
			}
		}
	}()

	// Wait for both goroutines to complete or for an error
	completedGoroutines := 0
	for completedGoroutines < 2 {
		select {
		case <-streamCtx.Done():
			c.log.Info("WebSocket stream context cancelled", map[string]any{
				"error": ctx.Err(),
			})
			return ctx.Err()
		case err := <-errCh:
			c.log.Error("WebSocket stream error", map[string]any{
				"error": err.Error(),
			})
			cancel() // Cancel other goroutine
			return err
		case <-doneCh:
			completedGoroutines++
			c.log.Info("WebSocket goroutine completed", map[string]any{
				"completed": completedGoroutines,
			})
		}
	}

	c.log.Info("WebSocket stream completed successfully", nil)
	return nil
}

// processStreamingResponse reads and parses NDJSON (newline-delimited JSON) response.
// The transcription API streams back multiple JSON objects, one per line, which we
// decode incrementally using json.Decoder. Each decoded object is passed to the callback
// for processing (typically word-by-word transcription with speaker diarization).
func (c *Client) processStreamingResponse(body io.Reader, callback StreamCallback) error {
	// Use a json.Decoder to decode successive JSON objects as they arrive.
	// Many streaming APIs emit a sequence of JSON objects (or newline-delimited
	// JSON). json.Decoder will handle consecutive objects separated by
	// whitespace/newlines.
	dec := json.NewDecoder(body)

	var totalProcessed int
	for {
		var resp StreamResponse
		if err := dec.Decode(&resp); err != nil {
			if err == io.EOF {
				// Stream ended cleanly
				c.log.Info("Reached end of transcription stream", map[string]any{"totalProcessed": totalProcessed})
				return nil
			}

			c.log.Error("Failed to decode streamed transcription response", map[string]any{
				"error": err.Error(),
			})
			return fmt.Errorf("failed to decode streamed response: %w", err)
		}

		// If the response doesn't contain alternatives, skip it
		if len(resp.Results.Channels) == 0 || len(resp.Results.Channels[0].Alternatives) == 0 {
			continue
		}

		alt := resp.Results.Channels[0].Alternatives[0]
		words := alt.Words

		c.log.Info("Processing streamed transcription response chunk", map[string]any{"words": len(words), "isFinal": resp.IsFinal})

		// Invoke callback with the decoded response directly so the caller
		// can handle grouping/word-level processing. Keep track of processed
		// words for diagnostics.
		if err := callback(&resp); err != nil {
			c.log.Error("Callback error processing streamed response", map[string]any{
				"error": err.Error(),
			})
			return err
		}

		totalProcessed += len(words)
	}
}
