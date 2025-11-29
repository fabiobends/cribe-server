package transcripts

import (
	"context"
	"fmt"
	"testing"

	"cribeapp.com/cribe-server/internal/clients/transcription"
	"cribeapp.com/cribe-server/internal/utils"
)

func TestTranscriptService_NewService(t *testing.T) {
	t.Run("should create a new service with dependencies", func(t *testing.T) {
		transcriptionClient := &MockTranscriptionClient{}
		llmClient := &MockLLMClient{}

		service := NewService(transcriptionClient, llmClient)

		if service == nil {
			t.Fatal("Expected service to be created, got nil")
		}

		if service.transcriptionClient == nil {
			t.Error("Expected transcriptionClient to be set")
		}

		if service.llmClient == nil {
			t.Error("Expected llmClient to be set")
		}

		if service.repo == nil {
			t.Error("Expected repo to be set")
		}
	})
}

func TestTranscriptService_StreamTranscript(t *testing.T) {
	t.Run("should stream transcript for non-existent episode and return error", func(t *testing.T) {
		transcriptionClient := &MockTranscriptionClient{}
		llmClient := &MockLLMClient{}
		service := NewService(transcriptionClient, llmClient)

		ctx := context.Background()
		chunkCount := 0

		err := service.StreamTranscript(ctx, 999999,
			func(chunk *Chunk) error {
				chunkCount++
				return nil
			},
			func(speaker *Speaker) error {
				return nil
			},
		)

		if err == nil {
			t.Error("Expected error for non-existent episode, got nil")
		}

		if chunkCount > 0 {
			t.Errorf("Expected 0 chunks for non-existent episode, got %d", chunkCount)
		}
	})

	t.Run("should stream from DB when transcript exists and is complete", func(t *testing.T) {
		transcriptionClient := &MockTranscriptionClient{}
		llmClient := &MockLLMClient{}
		service := NewService(transcriptionClient, llmClient)

		// Mock existing complete transcript
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

		// Mock chunks
		mockChunks := []TranscriptChunk{
			{Position: 0, Text: "Hello", StartTime: 0.0, EndTime: 0.5, SpeakerIndex: 0},
		}
		chunkExecutor := utils.QueryExecutor[TranscriptChunk]{
			QueryList: func(query string, args ...any) ([]TranscriptChunk, error) {
				return mockChunks, nil
			},
		}
		service.repo.chunkRepo.Executor = chunkExecutor

		ctx := context.Background()
		chunkCount := 0

		err := service.StreamTranscript(ctx, 1,
			func(chunk *Chunk) error {
				chunkCount++
				return nil
			},
			func(speaker *Speaker) error {
				return nil
			},
		)

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if chunkCount != 1 {
			t.Errorf("Expected 1 chunk from DB, got %d", chunkCount)
		}
	})

	t.Run("should create transcript and stream from API when not exists", func(t *testing.T) {
		transcriptionClient := &MockTranscriptionClient{}
		llmClient := &MockLLMClient{}
		service := NewService(transcriptionClient, llmClient)

		// Mock no existing transcript
		transcriptQueryExecutor := utils.QueryExecutor[Transcript]{
			QueryItem: func(query string, args ...any) (Transcript, error) {
				if query == "SELECT id, episode_id, status, error_message, created_at, completed_at FROM transcripts WHERE episode_id = $1" {
					return Transcript{}, fmt.Errorf("no rows in result set")
				}
				// For CreateTranscript
				return Transcript{ID: 1, EpisodeID: 1}, nil
			},
			Exec: func(query string, args ...any) error {
				return nil
			},
		}
		service.repo.transcriptRepo.Executor = transcriptQueryExecutor

		// Mock episode
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

		// Mock chunk executor
		chunkExecutor := utils.QueryExecutor[TranscriptChunk]{
			Exec: func(query string, args ...any) error {
				return nil
			},
		}
		service.repo.chunkRepo.Executor = chunkExecutor

		// Mock speaker executor
		speakerExecutor := utils.QueryExecutor[TranscriptSpeaker]{
			Exec: func(query string, args ...any) error {
				return nil
			},
		}
		service.repo.speakerRepo.Executor = speakerExecutor

		ctx := context.Background()
		chunkCount := 0

		err := service.StreamTranscript(ctx, 1,
			func(chunk *Chunk) error {
				chunkCount++
				return nil
			},
			func(speaker *Speaker) error {
				return nil
			},
		)

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if chunkCount < 1 {
			t.Errorf("Expected at least 1 chunk from API, got %d", chunkCount)
		}
	})
}

type MockFailingTranscriptionClient struct{}

func (m *MockFailingTranscriptionClient) StreamAudioURL(ctx context.Context, audioURL string, opts transcription.StreamOptions, callback transcription.StreamCallback) error {
	return fmt.Errorf("transcription API error")
}

func TestTranscriptService_StreamFromTranscriptionAPI_Error(t *testing.T) {
	t.Run("should handle transcription API error and update status to failed", func(t *testing.T) {
		transcriptionClient := &MockFailingTranscriptionClient{}
		llmClient := &MockLLMClient{}
		service := NewService(transcriptionClient, llmClient)

		// Mock transcript executor to track UpdateTranscriptStatus calls
		statusUpdated := false
		transcriptExecutor := utils.QueryExecutor[Transcript]{
			Exec: func(query string, args ...any) error {
				if len(args) >= 2 && args[0] == string(TranscriptStatusFailed) {
					statusUpdated = true
				}
				return nil
			},
		}
		service.repo.transcriptRepo.Executor = transcriptExecutor

		ctx := context.Background()

		err := service.streamFromTranscriptionAPI(ctx, 1, "https://example.com/audio.mp3", "Test episode",
			func(chunk *Chunk) error {
				return nil
			},
			func(speaker *Speaker) error {
				return nil
			},
		)

		if err == nil {
			t.Error("Expected error from failing transcription client, got nil")
		}

		if !statusUpdated {
			t.Error("Expected transcript status to be updated to failed")
		}
	})

	t.Run("should handle SaveChunks error", func(t *testing.T) {
		transcriptionClient := &MockTranscriptionClient{}
		llmClient := &MockLLMClient{}
		service := NewService(transcriptionClient, llmClient)

		// Mock chunk executor to return error
		chunkExecutor := utils.QueryExecutor[TranscriptChunk]{
			Exec: func(query string, args ...any) error {
				return fmt.Errorf("database error")
			},
		}
		service.repo.chunkRepo.Executor = chunkExecutor

		// Mock transcript executor for status updates
		transcriptExecutor := utils.QueryExecutor[Transcript]{
			Exec: func(query string, args ...any) error {
				return nil
			},
		}
		service.repo.transcriptRepo.Executor = transcriptExecutor

		ctx := context.Background()

		err := service.streamFromTranscriptionAPI(ctx, 1, "https://example.com/audio.mp3", "Test episode",
			func(chunk *Chunk) error {
				return nil
			},
			func(speaker *Speaker) error {
				return nil
			},
		)

		if err == nil {
			t.Error("Expected error from SaveChunks, got nil")
		}

		if err.Error() != "failed to save chunks: database error" {
			t.Errorf("Expected 'failed to save chunks: database error', got %v", err)
		}
	})

	t.Run("should handle UpdateTranscriptStatus error at completion", func(t *testing.T) {
		transcriptionClient := &MockTranscriptionClient{}
		llmClient := &MockLLMClient{}
		service := NewService(transcriptionClient, llmClient)

		// Mock chunk executor to succeed
		chunkExecutor := utils.QueryExecutor[TranscriptChunk]{
			Exec: func(query string, args ...any) error {
				return nil
			},
		}
		service.repo.chunkRepo.Executor = chunkExecutor

		// Mock transcript executor to fail on completion status update
		updateCallCount := 0
		transcriptExecutor := utils.QueryExecutor[Transcript]{
			Exec: func(query string, args ...any) error {
				// Check if this is an update (has status as first arg)
				if len(args) >= 2 {
					status, ok := args[0].(string)
					if ok && status == string(TranscriptStatusComplete) {
						updateCallCount++
						return fmt.Errorf("status update error")
					}
				}
				return nil
			},
		}
		service.repo.transcriptRepo.Executor = transcriptExecutor

		// Mock speaker executor to succeed
		speakerExecutor := utils.QueryExecutor[TranscriptSpeaker]{
			Exec: func(query string, args ...any) error {
				return nil
			},
		}
		service.repo.speakerRepo.Executor = speakerExecutor

		ctx := context.Background()

		err := service.streamFromTranscriptionAPI(ctx, 1, "https://example.com/audio.mp3", "Test episode",
			func(chunk *Chunk) error {
				return nil
			},
			func(speaker *Speaker) error {
				return nil
			},
		)

		if err == nil {
			t.Error("Expected error from UpdateTranscriptStatus, got nil")
		}

		if err.Error() != "status update error" {
			t.Errorf("Expected 'status update error', got %v", err)
		}

		if updateCallCount == 0 {
			t.Error("Expected UpdateTranscriptStatus to be called for completion")
		}
	})
}

func TestTranscriptService_GetExistingTranscript(t *testing.T) {

	t.Run("should return existing complete transcript", func(t *testing.T) {
		transcriptionClient := &MockTranscriptionClient{}
		llmClient := &MockLLMClient{}
		service := NewService(transcriptionClient, llmClient)

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

		id, exists, err := service.getExistingTranscript(1)

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if !exists {
			t.Error("Expected transcript to exist")
		}

		if id != 1 {
			t.Errorf("Expected transcript ID 1, got %d", id)
		}
	})

	t.Run("should return false for incomplete transcript", func(t *testing.T) {
		transcriptionClient := &MockTranscriptionClient{}
		llmClient := &MockLLMClient{}
		service := NewService(transcriptionClient, llmClient)

		mockTranscript := Transcript{
			ID:        1,
			EpisodeID: 1,
			Status:    string(TranscriptStatusProcessing),
		}

		transcriptExecutor := utils.QueryExecutor[Transcript]{
			QueryItem: func(query string, args ...any) (Transcript, error) {
				return mockTranscript, nil
			},
		}
		service.repo.transcriptRepo.Executor = transcriptExecutor

		_, exists, err := service.getExistingTranscript(1)

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if exists {
			t.Error("Expected transcript to not exist (incomplete)")
		}
	})

	t.Run("should return false when transcript not found", func(t *testing.T) {
		transcriptionClient := &MockTranscriptionClient{}
		llmClient := &MockLLMClient{}
		service := NewService(transcriptionClient, llmClient)

		transcriptExecutor := utils.QueryExecutor[Transcript]{
			QueryItem: func(query string, args ...any) (Transcript, error) {
				return Transcript{}, fmt.Errorf("no rows in result set")
			},
		}
		service.repo.transcriptRepo.Executor = transcriptExecutor

		_, exists, err := service.getExistingTranscript(1)

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if exists {
			t.Error("Expected transcript to not exist")
		}
	})
}

func TestTranscriptService_CreateTranscript(t *testing.T) {
	t.Run("should create new transcript", func(t *testing.T) {
		transcriptionClient := &MockTranscriptionClient{}
		llmClient := &MockLLMClient{}
		service := NewService(transcriptionClient, llmClient)

		transcriptExecutor := utils.QueryExecutor[Transcript]{
			QueryItem: func(query string, args ...any) (Transcript, error) {
				return Transcript{ID: 1, EpisodeID: 1}, nil
			},
		}
		service.repo.transcriptRepo.Executor = transcriptExecutor

		id, err := service.createTranscript(1)

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if id != 1 {
			t.Errorf("Expected transcript ID 1, got %d", id)
		}
	})
}
