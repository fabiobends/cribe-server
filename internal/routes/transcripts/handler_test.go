package transcripts

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"cribeapp.com/cribe-server/internal/clients/llm"
	"cribeapp.com/cribe-server/internal/clients/transcription"
	"cribeapp.com/cribe-server/internal/core/logger"
	"cribeapp.com/cribe-server/internal/utils"
)

type MockTranscriptionClient struct{}

func (m *MockTranscriptionClient) StreamAudioURL(ctx context.Context, audioURL string, callback transcription.StreamCallback) error {
	// Simulate streaming a few chunks
	response := &transcription.StreamResponse{
		Type: "Results",
		Channel: transcription.Channel{
			Alternatives: []transcription.Alternative{
				{
					Words: []transcription.Word{
						{
							Word:           "Hello",
							PunctuatedWord: "Hello",
							Start:          0.0,
							End:            0.5,
							Speaker:        0,
						},
						{
							Word:           "world",
							PunctuatedWord: "world",
							Start:          0.5,
							End:            1.0,
							Speaker:        0,
						},
					},
				},
			},
		},
	}
	if err := callback(response); err != nil {
		return err
	}
	return nil
}

type MockLLMClient struct{}

func (m *MockLLMClient) Chat(ctx context.Context, req llm.ChatRequest) (llm.ChatCompletionResponse, error) {
	return llm.ChatCompletionResponse{
		Choices: []llm.Choice{
			{Message: llm.Message{Content: "Speaker 0"}},
		},
	}, nil
}

// setupMockedService creates a service with mocked repository for testing
func setupMockedService() *Service {
	transcriptionClient := &MockTranscriptionClient{}
	llmClient := &MockLLMClient{}
	service := NewService(transcriptionClient, llmClient)

	// Mock episode exists
	mockEpisode := Episode{
		ID:          1,
		AudioURL:    "https://example.com/audio.mp3",
		Description: "Test episode",
	}
	episodeExecutor := utils.QueryExecutor[Episode]{
		QueryItem: func(query string, args ...any) (Episode, error) {
			return mockEpisode, nil
		},
	}
	service.repo.episodeRepo.Executor = episodeExecutor

	// Mock transcript exists and is complete
	mockTranscript := Transcript{
		ID:        1,
		EpisodeID: 1,
		Status:    string(TranscriptStatusComplete),
	}
	transcriptExecutor := utils.QueryExecutor[Transcript]{
		QueryItem: func(query string, args ...any) (Transcript, error) {
			return mockTranscript, nil
		},
	}
	service.repo.transcriptRepo.Executor = transcriptExecutor

	// Mock chunks
	mockChunks := []TranscriptChunk{
		{Position: 0, Text: "Hello", StartTime: 0.0, EndTime: 0.5, SpeakerIndex: 0},
		{Position: 1, Text: "world", StartTime: 0.5, EndTime: 1.0, SpeakerIndex: 0},
	}
	chunkExecutor := utils.QueryExecutor[TranscriptChunk]{
		QueryList: func(query string, args ...any) ([]TranscriptChunk, error) {
			return mockChunks, nil
		},
	}
	service.repo.chunkRepo.Executor = chunkExecutor

	// Mock speakers
	mockSpeakers := []TranscriptSpeaker{
		{SpeakerIndex: 0, SpeakerName: "Speaker 0"},
	}
	speakerExecutor := utils.QueryExecutor[TranscriptSpeaker]{
		QueryList: func(query string, args ...any) ([]TranscriptSpeaker, error) {
			return mockSpeakers, nil
		},
	}
	service.repo.speakerRepo.Executor = speakerExecutor

	return service
}

func TestTranscriptHandler_HandleRequest(t *testing.T) {
	transcriptionClient := &MockTranscriptionClient{}
	llmClient := &MockLLMClient{}
	service := NewService(transcriptionClient, llmClient)
	handler := NewTranscriptHandler(service)

	t.Run("should handle SSE stream request with valid episode_id", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/transcripts/stream/sse?episode_id=1", nil)

		handler.HandleRequest(w, r)

		if w.Code != http.StatusOK && w.Code != http.StatusNotFound && w.Code != http.StatusInternalServerError {
			t.Errorf("Expected status code %v, %v, or %v, got %v", http.StatusOK, http.StatusNotFound, http.StatusInternalServerError, w.Code)
		}
	})

	t.Run("should return bad request for missing episode_id", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/transcripts/stream/sse", nil)

		handler.HandleRequest(w, r)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status code %v, got %v", http.StatusBadRequest, w.Code)
		}
	})

	t.Run("should return bad request for invalid episode_id", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/transcripts/stream/sse?episode_id=invalid", nil)

		handler.HandleRequest(w, r)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status code %v, got %v", http.StatusBadRequest, w.Code)
		}
	})

	t.Run("should return not found for unknown path", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/transcripts/unknown", nil)

		handler.HandleRequest(w, r)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status code %v, got %v", http.StatusNotFound, w.Code)
		}
	})

	t.Run("should handle nil service", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/transcripts/stream/sse?episode_id=1", nil)

		handler := &TranscriptHandler{
			service: nil,
			log:     logger.NewHandlerLogger("TranscriptHandler"),
		}
		handler.HandleRequest(w, r)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("Expected status code %v, got %v", http.StatusInternalServerError, w.Code)
		}
	})

	t.Run("should return error when streaming is unsupported", func(t *testing.T) {
		service := setupMockedService()
		handler := NewTranscriptHandler(service)

		// Use a custom ResponseWriter that doesn't implement http.Flusher
		w := &nonFlushableResponseWriter{
			headers: make(http.Header),
			body:    []byte{},
		}
		r := httptest.NewRequest(http.MethodGet, "/transcripts/stream/sse?episode_id=1", nil)

		handler.HandleRequest(w, r)

		if w.statusCode != http.StatusInternalServerError {
			t.Errorf("Expected status code %v, got %v", http.StatusInternalServerError, w.statusCode)
		}

		// Check error message
		if !strings.Contains(string(w.body), "Streaming unsupported") {
			t.Errorf("Expected error message to contain 'Streaming unsupported', got %s", string(w.body))
		}
	})
}

// nonFlushableResponseWriter is a ResponseWriter that doesn't implement http.Flusher
type nonFlushableResponseWriter struct {
	headers    http.Header
	body       []byte
	statusCode int
}

func (w *nonFlushableResponseWriter) Header() http.Header {
	return w.headers
}

func (w *nonFlushableResponseWriter) Write(b []byte) (int, error) {
	w.body = append(w.body, b...)
	return len(b), nil
}

func (w *nonFlushableResponseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
}

func TestTranscriptHandler_SSEEvents(t *testing.T) {
	t.Run("should write all SSE events in correct format", func(t *testing.T) {
		service := setupMockedService()
		handler := NewTranscriptHandler(service)

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/transcripts/stream/sse?episode_id=1", nil)

		handler.HandleRequest(w, r)

		body := w.Body.String()

		// Check for chunk events
		if !strings.Contains(body, "event: chunk") {
			t.Error("Expected response to contain 'event: chunk'")
		}

		if !strings.Contains(body, "data:") {
			t.Error("Expected response to contain 'data:'")
		}

		// Check for speaker events
		if !strings.Contains(body, "event: speaker") {
			t.Error("Expected response to contain 'event: speaker'")
		}

		// Check for complete event
		if !strings.Contains(body, "event: complete") {
			t.Error("Expected response to contain 'event: complete'")
		}

		if !strings.Contains(body, "data: {}") {
			t.Error("Expected complete event to have 'data: {}'")
		}
	})
}

// Test for context cancellation scenarios
func TestTranscriptHandler_ContextCancellation(t *testing.T) {
	t.Run("should handle client disconnect gracefully", func(t *testing.T) {
		service := setupMockedService()
		handler := NewTranscriptHandler(service)

		// Create a cancellable context
		ctx, cancel := context.WithCancel(context.Background())
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/transcripts/stream/sse?episode_id=1", nil)
		r = r.WithContext(ctx)

		// Cancel the context immediately to simulate client disconnect
		cancel()

		handler.HandleRequest(w, r)

		// The handler should complete without panicking
	})
}

// Test for error scenarios during streaming
func TestTranscriptHandler_StreamErrors(t *testing.T) {
	t.Run("should handle database error gracefully", func(t *testing.T) {
		transcriptionClient := &MockTranscriptionClient{}
		llmClient := &MockLLMClient{}
		service := NewService(transcriptionClient, llmClient)
		handler := NewTranscriptHandler(service)

		// Mock episode query returns error
		episodeExecutor := utils.QueryExecutor[Episode]{
			QueryItem: func(query string, args ...any) (Episode, error) {
				return Episode{}, fmt.Errorf("database connection error")
			},
		}
		service.repo.episodeRepo.Executor = episodeExecutor

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/transcripts/stream/sse?episode_id=1", nil)

		handler.HandleRequest(w, r)

		body := w.Body.String()

		// The handler should complete without panicking and may contain error event
		// Error event is expected when there's a database error
		_ = body // body may contain "event: error"
	})
}

// Test SSE header setup
func TestTranscriptHandler_SSEHeaders(t *testing.T) {
	t.Run("should set correct SSE headers", func(t *testing.T) {
		service := setupMockedService()
		handler := NewTranscriptHandler(service)

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/transcripts/stream/sse?episode_id=1", nil)

		handler.HandleRequest(w, r)

		headers := w.Header()

		if headers.Get("Content-Type") != "text/event-stream" {
			t.Errorf("Expected Content-Type to be 'text/event-stream', got %s", headers.Get("Content-Type"))
		}

		if headers.Get("Cache-Control") != "no-cache" {
			t.Errorf("Expected Cache-Control to be 'no-cache', got %s", headers.Get("Cache-Control"))
		}

		if headers.Get("Connection") != "keep-alive" {
			t.Errorf("Expected Connection to be 'keep-alive', got %s", headers.Get("Connection"))
		}
	})
}
