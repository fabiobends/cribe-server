package transcripts

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"cribeapp.com/cribe-server/internal/clients/llm"
	"cribeapp.com/cribe-server/internal/clients/transcription"
	"cribeapp.com/cribe-server/internal/utils"
)

// Test helpers
func setupService() *Service {
	return NewService(&MockTranscriptionClient{}, &MockLLMClient{})
}

func setupMockRepos(service *Service, transcriptExists bool) {
	service.repo.transcriptRepo.Executor = utils.QueryExecutor[Transcript]{
		QueryItem: func(query string, args ...any) (Transcript, error) {
			if !transcriptExists && strings.Contains(query, "WHERE episode_id") {
				return Transcript{}, fmt.Errorf("no rows in result set")
			}
			return Transcript{ID: 1, Status: string(TranscriptStatusComplete)}, nil
		},
		Exec: func(query string, args ...any) error { return nil },
	}
	service.repo.speakerRepo.Executor = utils.QueryExecutor[TranscriptSpeaker]{
		QueryList: func(query string, args ...any) ([]TranscriptSpeaker, error) {
			return []TranscriptSpeaker{{SpeakerIndex: 0, SpeakerName: "Speaker 0"}}, nil
		},
		Exec: func(query string, args ...any) error { return nil },
	}
	service.repo.chunkRepo.Executor = utils.QueryExecutor[TranscriptChunk]{
		QueryList: func(query string, args ...any) ([]TranscriptChunk, error) {
			return []TranscriptChunk{{Position: 0, Text: "Hello"}}, nil
		},
		Exec: func(query string, args ...any) error { return nil },
	}
	service.repo.episodeRepo.Executor = utils.QueryExecutor[Episode]{
		QueryItem: func(query string, args ...any) (Episode, error) {
			return Episode{ID: 1, AudioURL: "test.mp3", Description: "Test"}, nil
		},
	}
}

func TestTranscriptService_StreamTranscript(t *testing.T) {
	tests := []struct {
		name             string
		transcriptExists bool
		wantChunks       int
		wantError        bool
	}{
		{"stream from DB", true, 1, false},
		{"stream from API", false, 1, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := setupService()
			setupMockRepos(service, tt.transcriptExists)

			chunkCount := 0
			err := service.StreamTranscript(context.Background(), 1,
				func(chunk *Chunk) error {
					chunkCount++
					return nil
				},
				func(speaker *Speaker) error { return nil },
			)

			if (err != nil) != tt.wantError {
				t.Errorf("error = %v, wantError %v", err, tt.wantError)
			}
			if chunkCount < tt.wantChunks {
				t.Errorf("got %d chunks, want at least %d", chunkCount, tt.wantChunks)
			}
		})
	}
}

type mockFailingTranscriptionClient struct{}

func (m *mockFailingTranscriptionClient) StreamAudioURL(ctx context.Context, audioURL string, callback transcription.StreamCallback) error {
	return fmt.Errorf("transcription API error")
}

func TestTranscriptService_ErrorHandling(t *testing.T) {
	service := NewService(&mockFailingTranscriptionClient{}, &MockLLMClient{})
	statusUpdated := false

	service.repo.transcriptRepo.Executor = utils.QueryExecutor[Transcript]{
		QueryItem: func(query string, args ...any) (Transcript, error) {
			return Transcript{ID: 1}, nil
		},
		Exec: func(query string, args ...any) error {
			if len(args) >= 3 {
				// Check for failed status update (3 args: status, errorMessage, transcriptID)
				if status, ok := args[0].(TranscriptStatus); ok && status == TranscriptStatusFailed {
					statusUpdated = true
				}
			}
			return nil
		},
	}

	err := service.streamFromTranscriptionAPI(context.Background(), 1, "test.mp3", "Test",
		func(chunk *Chunk) error { return nil },
		func(speaker *Speaker) error { return nil },
	)

	if err == nil || !statusUpdated {
		t.Errorf("Expected error and status update, got error=%v updated=%v", err, statusUpdated)
	}
}

func TestTranscriptService_BuildEarlySpeakerContext(t *testing.T) {
	service := setupService()
	chunks := []Chunk{
		{Position: 0, SpeakerIndex: 1, Text: "Before1"},
		{Position: 1, SpeakerIndex: 1, Text: "Before2"},
		{Position: 2, SpeakerIndex: 0, Text: "Target"},
		{Position: 3, SpeakerIndex: 1, Text: "After1"},
		{Position: 4, SpeakerIndex: 1, Text: "After2"},
	}

	contextWords := service.buildEarlySpeakerContext(chunks, 0)
	expected := []string{"Before1", "Before2", "Target", "After1", "After2"}

	if len(contextWords) != len(expected) {
		t.Fatalf("got %d words, want %d", len(contextWords), len(expected))
	}
	for i, want := range expected {
		if contextWords[i] != want {
			t.Errorf("pos %d: got '%s', want '%s'", i, contextWords[i], want)
		}
	}
}

func TestTranscriptService_EarlyInference(t *testing.T) {
	service := NewService(&customMockTranscriptionClient{wordCount: 60}, &MockLLMClient{})
	setupMockRepos(service, false)

	var speakerNames []string
	err := service.streamFromTranscriptionAPI(context.Background(), 1, "test.mp3", "Test episode",
		func(chunk *Chunk) error { return nil },
		func(speaker *Speaker) error {
			speakerNames = append(speakerNames, speaker.Name)
			return nil
		},
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	if len(speakerNames) < 2 {
		t.Errorf("got %d speaker callbacks, want at least 2", len(speakerNames))
	}
	if len(speakerNames) >= 1 && speakerNames[0] != "Speaker 0" {
		t.Errorf("got first speaker '%s', want 'Speaker 0'", speakerNames[0])
	}
}

type customMockTranscriptionClient struct {
	wordCount int
}

func (m *customMockTranscriptionClient) StreamAudioURL(ctx context.Context, audioURL string, callback transcription.StreamCallback) error {
	for i := 0; i < m.wordCount; i++ {
		if err := callback(&transcription.StreamResponse{
			Type: "Results",
			Channel: transcription.Channel{
				Alternatives: []transcription.Alternative{{
					Words: []transcription.Word{{
						Word:           fmt.Sprintf("word%d", i),
						PunctuatedWord: fmt.Sprintf("word%d", i),
						Start:          float64(i),
						End:            float64(i) + 0.5,
						Speaker:        0,
					}},
				}},
			},
		}); err != nil {
			return err
		}
	}
	return nil
}

func TestTranscriptService_SaveChunksBatchedError(t *testing.T) {
	service := setupService()
	var statusUpdateError string

	service.repo.transcriptRepo.Executor = utils.QueryExecutor[Transcript]{
		Exec: func(query string, args ...any) error {
			if len(args) >= 3 {
				// Check for failed status update (3 args: status, errorMessage, transcriptID)
				if status, ok := args[0].(TranscriptStatus); ok && status == TranscriptStatusFailed {
					if errMsg, ok := args[1].(string); ok {
						statusUpdateError = errMsg
					}
				}
			}
			return nil
		},
	}
	service.repo.chunkRepo.Executor = utils.QueryExecutor[TranscriptChunk]{
		Exec: func(query string, args ...any) error {
			return fmt.Errorf("database connection error")
		},
	}

	service.saveTranscriptInBackground(1,
		[]Chunk{{Position: 0, Text: "Test", SpeakerIndex: 0, Start: 0.0, End: 1.0}},
		make(map[int][]string), "test", make(map[int]bool),
	)

	if statusUpdateError != "database connection error" {
		t.Errorf("got error '%s', want 'database connection error'", statusUpdateError)
	}
}

func TestTranscriptService_SpeakerInferenceError(t *testing.T) {
	t.Run("continues after speaker inference failure", func(t *testing.T) {
		// Create a mock LLM client that fails
		failingLLM := &mockFailingLLMClient{shouldFail: true}
		service := NewService(&MockTranscriptionClient{}, failingLLM)

		var upsertCalled bool
		service.repo.speakerRepo.Executor = utils.QueryExecutor[TranscriptSpeaker]{
			Exec: func(query string, args ...any) error {
				upsertCalled = true
				return nil
			},
		}
		service.repo.transcriptRepo.Executor = utils.QueryExecutor[Transcript]{
			Exec: func(query string, args ...any) error {
				return nil
			},
		}
		service.repo.chunkRepo.Executor = utils.QueryExecutor[TranscriptChunk]{
			Exec: func(query string, args ...any) error {
				return nil
			},
		}

		// Run background saving with speaker inference
		chunks := []Chunk{
			{Position: 0, Text: "Test", SpeakerIndex: 0, Start: 0.0, End: 1.0},
			{Position: 1, Text: "More", SpeakerIndex: 0, Start: 1.0, End: 2.0},
		}
		service.saveTranscriptInBackground(1, chunks, make(map[int][]string), "test", make(map[int]bool))

		// Wait for goroutine to complete
		time.Sleep(100 * time.Millisecond)

		// Verify that UpsertSpeaker was NOT called since inference failed
		if upsertCalled {
			t.Error("Expected UpsertSpeaker to not be called after inference failure")
		}
	})
}

func TestTranscriptService_UpsertSpeakerError(t *testing.T) {
	t.Run("logs error when speaker save fails", func(t *testing.T) {
		service := setupService()

		var upsertAttempted bool
		service.repo.speakerRepo.Executor = utils.QueryExecutor[TranscriptSpeaker]{
			Exec: func(query string, args ...any) error {
				upsertAttempted = true
				return fmt.Errorf("database write error")
			},
		}
		service.repo.transcriptRepo.Executor = utils.QueryExecutor[Transcript]{
			Exec: func(query string, args ...any) error {
				return nil
			},
		}
		service.repo.chunkRepo.Executor = utils.QueryExecutor[TranscriptChunk]{
			Exec: func(query string, args ...any) error {
				return nil
			},
		}

		// Run background saving with speaker inference
		chunks := []Chunk{
			{Position: 0, Text: "Test word1", SpeakerIndex: 0, Start: 0.0, End: 1.0},
			{Position: 1, Text: "Test word2", SpeakerIndex: 0, Start: 1.0, End: 2.0},
		}
		service.saveTranscriptInBackground(1, chunks, make(map[int][]string), "test", make(map[int]bool))

		// Wait for goroutine to complete
		time.Sleep(100 * time.Millisecond)

		// Verify that UpsertSpeaker was attempted
		if !upsertAttempted {
			t.Error("Expected UpsertSpeaker to be called")
		}
	})
}

func TestTranscriptService_SpeakerInferenceWithRetry(t *testing.T) {
	t.Run("successfully saves speaker after inference", func(t *testing.T) {
		service := setupService()

		var savedSpeakers []string
		service.repo.speakerRepo.Executor = utils.QueryExecutor[TranscriptSpeaker]{
			Exec: func(query string, args ...any) error {
				if len(args) >= 3 {
					if name, ok := args[2].(string); ok {
						savedSpeakers = append(savedSpeakers, name)
					}
				}
				return nil
			},
		}
		service.repo.transcriptRepo.Executor = utils.QueryExecutor[Transcript]{
			Exec: func(query string, args ...any) error {
				return nil
			},
		}
		service.repo.chunkRepo.Executor = utils.QueryExecutor[TranscriptChunk]{
			Exec: func(query string, args ...any) error {
				return nil
			},
		}

		// Run background saving with speaker inference
		chunks := []Chunk{
			{Position: 0, Text: "Hello from speaker", SpeakerIndex: 0, Start: 0.0, End: 1.0},
			{Position: 1, Text: "More from speaker", SpeakerIndex: 0, Start: 1.0, End: 2.0},
		}
		service.saveTranscriptInBackground(1, chunks, make(map[int][]string), "test", make(map[int]bool))

		// Wait for goroutine to complete
		time.Sleep(100 * time.Millisecond)

		// Verify speaker was saved
		if len(savedSpeakers) == 0 {
			t.Error("Expected at least one speaker to be saved")
		}
		if len(savedSpeakers) > 0 && savedSpeakers[0] != "Speaker 0" {
			t.Errorf("Expected 'Speaker 0', got '%s'", savedSpeakers[0])
		}
	})

	t.Run("skips already inferred speakers", func(t *testing.T) {
		service := setupService()

		var upsertCount int
		service.repo.speakerRepo.Executor = utils.QueryExecutor[TranscriptSpeaker]{
			Exec: func(query string, args ...any) error {
				upsertCount++
				return nil
			},
		}
		service.repo.transcriptRepo.Executor = utils.QueryExecutor[Transcript]{
			Exec: func(query string, args ...any) error {
				return nil
			},
		}
		service.repo.chunkRepo.Executor = utils.QueryExecutor[TranscriptChunk]{
			Exec: func(query string, args ...any) error {
				return nil
			},
		}

		// Run with speaker already inferred
		chunks := []Chunk{
			{Position: 0, Text: "Test", SpeakerIndex: 0, Start: 0.0, End: 1.0},
			{Position: 1, Text: "More", SpeakerIndex: 1, Start: 1.0, End: 2.0},
		}
		speakerInferred := map[int]bool{0: true} // Speaker 0 already inferred

		service.saveTranscriptInBackground(1, chunks, make(map[int][]string), "test", speakerInferred)

		// Wait for goroutines to complete
		time.Sleep(100 * time.Millisecond)

		// Should only save speaker 1 (speaker 0 was already inferred)
		if upsertCount != 1 {
			t.Errorf("Expected 1 upsert call, got %d", upsertCount)
		}
	})
}

type mockFailingLLMClient struct {
	shouldFail bool
}

func (m *mockFailingLLMClient) Chat(ctx context.Context, req llm.ChatRequest) (llm.ChatCompletionResponse, error) {
	if m.shouldFail {
		return llm.ChatCompletionResponse{}, fmt.Errorf("LLM API error")
	}
	return llm.ChatCompletionResponse{
		Choices: []llm.Choice{
			{Message: llm.Message{Content: "Speaker 0"}},
		},
	}, nil
}
