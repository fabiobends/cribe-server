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
	defer cancel()

	eventCh := make(chan string, 512)
	writeErrCh := make(chan error, 1)

	// Writer goroutine: consumes formatted SSE strings and writes them
	// to the ResponseWriter. On error it reports via writeErrCh and
	// cancels the context.
	go func() {
		for {
			select {
			case <-streamCtx.Done():
				return
			case s, ok := <-eventCh:
				if !ok {
					return
				}
				if _, err := fmt.Fprint(w, s); err != nil {
					// Report the write error and cancel the stream.
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
	enqueue := func(payload string) error {
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
			return enqueue(payload)
		},
		// Speaker callback
		func(speaker *Speaker) error {
			data, _ := json.Marshal(speaker)
			payload := fmt.Sprintf("event: speaker\ndata: %s\n\n", data)
			return enqueue(payload)
		},
	)

	// Close event channel to allow writer goroutine to exit if it's idle.
	close(eventCh)

	// If writer reported an error, prefer that as the root cause.
	select {
	case wErr := <-writeErrCh:
		if wErr != nil {
			err = wErr
		}
	default:
	}

	if err != nil {
		h.log.Error("Stream error", map[string]any{"error": err})
		errorMsg := map[string]string{"error": err.Error()}
		data, _ := json.Marshal(errorMsg)
		if _, writeErr := fmt.Fprintf(w, "event: error\ndata: %s\n\n", data); writeErr != nil {
			h.log.Error("Failed to write error event", map[string]any{"error": writeErr})
		}
		flusher.Flush()
		return
	}

	// Send complete event
	if _, err := fmt.Fprintf(w, "event: complete\ndata: {}\n\n"); err != nil {
		h.log.Error("Failed to write complete event", map[string]any{"error": err})
		return
	}
	flusher.Flush()
}
