package transcripts

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"cribeapp.com/cribe-server/internal/clients/transcription"
	"cribeapp.com/cribe-server/internal/core/logger"
	"cribeapp.com/cribe-server/internal/utils"
)

type MockTranscriptionClient struct{}

func (m *MockTranscriptionClient) StreamAudioURL(ctx context.Context, audioURL string, opts transcription.StreamOptions, callback transcription.StreamCallback) error {
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

func (m *MockTranscriptionClient) StreamAudioURLWebSocket(ctx context.Context, audioURL string, opts transcription.StreamOptions, callback transcription.StreamCallback) error {
	// Use the same mock implementation as StreamAudioURL
	return m.StreamAudioURL(ctx, audioURL, opts, callback)
}

type MockLLMClient struct{}

func (m *MockLLMClient) InferSpeakerName(ctx context.Context, episodeDescription string, speakerIndex int, transcriptChunks []string) (string, error) {
	return "Speaker 0", nil
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
}

func TestTranscriptHandler_SSEEvents(t *testing.T) {
	t.Run("should write chunk events in SSE format", func(t *testing.T) {
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
	})

	t.Run("should write speaker events in SSE format", func(t *testing.T) {
		service := setupMockedService()
		handler := NewTranscriptHandler(service)

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/transcripts/stream/sse?episode_id=1", nil)

		handler.HandleRequest(w, r)

		body := w.Body.String()

		// Check for speaker events
		if !strings.Contains(body, "event: speaker") {
			t.Error("Expected response to contain 'event: speaker'")
		}
	})

	t.Run("should write complete event at the end", func(t *testing.T) {
		service := setupMockedService()
		handler := NewTranscriptHandler(service)

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/transcripts/stream/sse?episode_id=1", nil)

		handler.HandleRequest(w, r)

		body := w.Body.String()

		// Check for complete event
		if !strings.Contains(body, "event: complete") {
			t.Error("Expected response to contain 'event: complete'")
		}

		if !strings.Contains(body, "data: {}") {
			t.Error("Expected complete event to have 'data: {}'")
		}
	})
}
