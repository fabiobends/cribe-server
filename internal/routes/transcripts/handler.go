package transcripts

import (
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

	// Stream transcript
	err = h.service.StreamTranscript(r.Context(), episodeID,
		// Chunk callback
		func(chunk *Chunk) error {
			data, _ := json.Marshal(chunk)
			if _, err := fmt.Fprintf(w, "event: chunk\ndata: %s\n\n", data); err != nil {
				return err
			}
			flusher.Flush()
			return nil
		},
		// Speaker callback
		func(speaker *Speaker) error {
			data, _ := json.Marshal(speaker)
			if _, err := fmt.Fprintf(w, "event: speaker\ndata: %s\n\n", data); err != nil {
				return err
			}
			flusher.Flush()
			return nil
		},
	)

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
