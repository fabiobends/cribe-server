package transcripts

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"cribeapp.com/cribe-server/internal/core/logger"
	"cribeapp.com/cribe-server/internal/utils"
)

type TranscriptHandler struct {
	service *Service
	log     *logger.ContextualLogger
}

func NewTranscriptHandler(service *Service) *TranscriptHandler {
	return &TranscriptHandler{
		service: service,
		log:     logger.NewHandlerLogger("TranscriptHandler"),
	}
}

func (h *TranscriptHandler) HandleRequest(w http.ResponseWriter, r *http.Request) {
	h.log.Debug("Processing transcript request", map[string]any{
		"method": r.Method,
		"path":   r.URL.Path,
	})

	if h.service == nil {
		h.log.Error("Transcript service not configured", nil)
		utils.EncodeResponse(w, http.StatusInternalServerError, map[string]string{
			"error": "Service configuration error",
		})
		return
	}

	// Route based on path
	path := strings.TrimPrefix(r.URL.Path, "/transcripts")

	switch {
	case path == "/stream/sse" && r.Method == "GET":
		h.handleSSEStream(w, r)
	default:
		utils.NotFound(w, r)
	}
}

// handleSSEStream handles SSE streaming of transcript chunks
func (h *TranscriptHandler) handleSSEStream(w http.ResponseWriter, r *http.Request) {
	// Get episode ID from query params
	episodeIDStr := r.URL.Query().Get("episode_id")

	if episodeIDStr == "" {
		utils.EncodeResponse(w, http.StatusBadRequest, map[string]string{
			"error": "episode_id required",
		})
		return
	}

	episodeID, err := strconv.Atoi(episodeIDStr)
	if err != nil {
		utils.EncodeResponse(w, http.StatusBadRequest, map[string]string{
			"error": "Invalid episode_id",
		})
		return
	}

	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		utils.EncodeResponse(w, http.StatusInternalServerError, map[string]string{
			"error": "Streaming unsupported",
		})
		return
	}

	h.log.Info("Starting SSE stream", map[string]any{
		"episodeID": episodeID,
	})

	// Use a buffered channel and a dedicated writer goroutine for SSE
	// to avoid blocking the transcription callbacks when the HTTP client
	// is slow to read. If a write fails, cancel the context so the
	// transcription stream can stop promptly.
	streamCtx, cancel := context.WithCancel(r.Context())
	defer func() {
		cancel()
		select {
		case <-r.Context().Done():
			h.log.Info("SSE stream ended (client disconnected)", map[string]any{
				"episodeID": episodeID,
			})
		default:
			h.log.Info("SSE stream ended (completed normally)", map[string]any{
				"episodeID": episodeID,
			})
		}
	}()

	eventCh := make(chan string, 512)
	writeErrCh := make(chan error, 1)
	writerDone := make(chan struct{})

	// Writer goroutine: consumes formatted SSE strings and writes them
	// to the ResponseWriter. On error it reports via writeErrCh and
	// cancels the context.
	go func() {
		defer close(writerDone)
		for {
			select {
			case <-streamCtx.Done():
				h.log.Info("Writer goroutine: context cancelled", map[string]any{
					"episodeID": episodeID,
					"reason":    streamCtx.Err(),
				})
				return
			case s, ok := <-eventCh:
				if !ok {
					h.log.Debug("Writer goroutine: event channel closed", map[string]any{
						"episodeID": episodeID,
					})
					return
				}
				if _, err := fmt.Fprint(w, s); err != nil {
					// Report the write error and cancel the stream.
					h.log.Warn("Writer goroutine: write failed (client disconnected?)", map[string]any{
						"episodeID": episodeID,
						"error":     err,
					})
					select {
					case writeErrCh <- err:
					default:
					}
					cancel()
					return
				}
				flusher.Flush()
			}
		}
	}()

	// Helper to enqueue SSE event payloads without blocking indefinitely.
	// Returns error if context is cancelled or send fails.
	enqueue := func(payload string) (err error) {
		// Recover from panic if channel is closed
		defer func() {
			if r := recover(); r != nil {
				h.log.Debug("Attempted to send to closed channel (race condition)", map[string]any{
					"episodeID": episodeID,
				})
				err = fmt.Errorf("channel closed")
			}
		}()

		select {
		case eventCh <- payload:
			return nil
		case <-streamCtx.Done():
			return streamCtx.Err()
		}
	}

	// Stream transcript: push events to the channel instead of writing
	// directly to the response.
	err = h.service.StreamTranscript(streamCtx, episodeID,
		// Chunk callback
		func(chunk *Chunk) error {
			data, _ := json.Marshal(chunk)
			payload := fmt.Sprintf("event: chunk\ndata: %s\n\n", data)
			if err := enqueue(payload); err != nil {
				// Context cancelled or channel closed - stop processing
				return err
			}
			return nil
		},
		// Speaker callback
		func(speaker *Speaker) error {
			data, _ := json.Marshal(speaker)
			payload := fmt.Sprintf("event: speaker\ndata: %s\n\n", data)
			if err := enqueue(payload); err != nil {
				// Context cancelled or channel closed - stop processing
				return err
			}
			return nil
		},
	)

	h.log.Debug("StreamTranscript completed", map[string]any{
		"episodeID": episodeID,
		"error":     err,
	})

	// Check if context was cancelled before trying to send final events
	if streamCtx.Err() != nil {
		h.log.Info("Context cancelled, skipping final event", map[string]any{
			"episodeID": episodeID,
		})
		// Close channel and wait for writer to finish
		close(eventCh)
		<-writerDone
		return
	}

	// Send complete or error event through the channel before closing
	if err != nil {
		h.log.Error("Stream error", map[string]any{
			"episodeID": episodeID,
			"error":     err,
		})
		errorMsg := map[string]string{"error": err.Error()}
		data, _ := json.Marshal(errorMsg)
		payload := fmt.Sprintf("event: error\ndata: %s\n\n", data)
		// Try to send error event, but don't block if channel is full or context cancelled
		select {
		case eventCh <- payload:
		case <-streamCtx.Done():
		default:
		}
	} else {
		h.log.Info("Stream completed successfully", map[string]any{
			"episodeID": episodeID,
		})
		// Send complete event through the channel
		payload := "event: complete\ndata: {}\n\n"
		select {
		case eventCh <- payload:
		case <-streamCtx.Done():
		default:
		}
	}

	// Close event channel to signal writer goroutine to exit
	close(eventCh)

	h.log.Debug("Event channel closed, waiting for writer to finish", map[string]any{
		"episodeID": episodeID,
	})

	// Wait for writer goroutine to finish processing all events
	<-writerDone

	// Check if writer reported an error
	select {
	case wErr := <-writeErrCh:
		if wErr != nil {
			h.log.Error("Writer error", map[string]any{"error": wErr})
		}
	default:
	}
}
