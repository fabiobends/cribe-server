package transcription

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/gorilla/websocket"

	"cribeapp.com/cribe-server/internal/core/logger"
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
		apiKey:     apiKey,
		baseURL:    baseURL,
		httpClient: &http.Client{},
		log:        log,
	}
}

// StreamAudioURL is the service-level method - streams audio from URL with default settings
// Service only needs to pass audioURL and callback, client handles all infrastructure
func (c *Client) StreamAudioURL(audioURL string, callback StreamCallback) error {
	// Use background context so transcription continues even if client disconnects
	ctx := context.Background()

	// Default options configured in client
	opts := StreamOptions{
		Model:      "nova-3",
		Language:   "en",
		Diarize:    true,
		Punctuate:  true,
		Utterances: false,
	}

	return c.StreamAudioURLWebSocket(ctx, audioURL, opts, callback)
}

// buildWebSocketURL constructs a WebSocket URL with query parameters for Deepgram API
func (c *Client) buildWebSocketURL(opts StreamOptions) (string, error) {
	wsURL := strings.Replace(c.baseURL, "https://", "wss://", 1)

	u, err := url.Parse(wsURL + "/listen")
	if err != nil {
		return "", fmt.Errorf("failed to parse WebSocket URL: %w", err)
	}

	q := u.Query()
	q.Set("model", opts.Model)
	q.Set("language", opts.Language)
	q.Set("diarize", fmt.Sprintf("%t", opts.Diarize))
	q.Set("punctuate", fmt.Sprintf("%t", opts.Punctuate))
	q.Set("utterances", fmt.Sprintf("%t", opts.Utterances))
	q.Set("interim_results", "false")
	u.RawQuery = q.Encode()

	return u.String(), nil
}

// streamAudioToWebSocket downloads audio from URL and streams it to the WebSocket connection
func (c *Client) streamAudioToWebSocket(ctx context.Context, conn *websocket.Conn, audioURL string, errCh chan<- error, doneCh chan<- struct{}) {
	req, err := http.NewRequestWithContext(ctx, "GET", audioURL, nil)
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

	reader := bufio.NewReader(resp.Body)
	buffer := make([]byte, 8192)
	totalBytesSent := 0

	for {
		select {
		case <-ctx.Done():
			errCh <- ctx.Err()
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
		}

		if err != nil {
			if err == io.EOF {
				c.log.Info("Finished streaming audio to WebSocket", map[string]any{
					"totalBytesSent": totalBytesSent,
				})
				closeMsg := []byte("{\"type\":\"CloseStream\"}")
				if err := conn.WriteMessage(websocket.TextMessage, closeMsg); err != nil {
					c.log.Error("Failed to send close stream message", map[string]any{
						"error": err.Error(),
					})
				}
				doneCh <- struct{}{}
				return
			}
			errCh <- fmt.Errorf("failed to read audio: %w", err)
			return
		}
	}
}

// readTranscriptionResults reads and processes messages from the WebSocket connection
func (c *Client) readTranscriptionResults(ctx context.Context, conn *websocket.Conn, callback StreamCallback, errCh chan<- error, doneCh chan<- struct{}) {
	defer func() {
		if err := conn.Close(); err != nil {
			c.log.Error("Failed to close WebSocket in reader goroutine", map[string]any{
				"error": err.Error(),
			})
		}
	}()

	messageCount := 0
	for {
		select {
		case <-ctx.Done():
			c.log.Info("WebSocket reader stopping due to context cancellation", map[string]any{
				"messagesReceived": messageCount,
			})
			errCh <- ctx.Err()
			return
		default:
		}

		_, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
				c.log.Info("WebSocket closed normally", map[string]any{
					"messagesReceived": messageCount,
				})
				doneCh <- struct{}{}
				return
			}
			errCh <- fmt.Errorf("failed to read WebSocket message: %w", err)
			return
		}

		messageCount++
		if messageCount%10 == 0 {
			c.log.Debug("WebSocket progress", map[string]any{
				"messagesReceived": messageCount,
			})
		}

		var resp StreamResponse
		if err := json.Unmarshal(message, &resp); err != nil {
			c.log.Error("Failed to unmarshal WebSocket message", map[string]any{
				"error":   err.Error(),
				"message": string(message),
			})
			continue
		}

		if resp.Type != "" && resp.Type != "Results" {
			continue
		}

		if len(resp.Channel.Alternatives) == 0 {
			c.log.Debug("Received message with no alternatives", map[string]any{
				"type": resp.Type,
			})
			continue
		}

		if !resp.IsFinal {
			continue
		}

		if err := callback(&resp); err != nil {
			c.log.Error("Callback error processing WebSocket response", map[string]any{
				"error": err.Error(),
			})
			errCh <- err
			return
		}
	}
}

// StreamAudioURLWebSocket streams audio from a URL to Deepgram's WebSocket endpoint
// for real-time transcription. This enables progressive transcription of long-form
// audio (e.g., 2+ hour podcasts) by downloading and streaming audio chunks concurrently
// with receiving transcription results.
func (c *Client) StreamAudioURLWebSocket(ctx context.Context, audioURL string, opts StreamOptions, callback StreamCallback) error {
	wsURLString, err := c.buildWebSocketURL(opts)
	if err != nil {
		return err
	}

	headers := http.Header{}
	headers.Set("Authorization", "Token "+c.apiKey)

	c.log.Info("Connecting to Deepgram WebSocket", map[string]any{
		"url":      wsURLString,
		"audioURL": audioURL,
	})

	conn, _, err := websocket.DefaultDialer.Dial(wsURLString, headers)
	if err != nil {
		c.log.Error("Failed to connect to WebSocket", map[string]any{
			"error": err.Error(),
		})
		return fmt.Errorf("failed to connect to WebSocket: %w", err)
	}

	streamCtx, cancel := context.WithCancel(ctx)
	errCh := make(chan error, 2)
	doneCh := make(chan struct{}, 2)

	go c.streamAudioToWebSocket(streamCtx, conn, audioURL, errCh, doneCh)
	go func() {
		defer cancel()
		c.readTranscriptionResults(streamCtx, conn, callback, errCh, doneCh)
	}()

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
			cancel()
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

		c.log.Debug("Processing streamed transcription response chunk", map[string]any{"words": len(words), "isFinal": resp.IsFinal})

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
