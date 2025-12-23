package transcripts

import (
	"errors"
	"testing"
	"time"

	"cribeapp.com/cribe-server/internal/utils"
)

func TestTranscriptRepository_GetTranscriptByEpisodeID(t *testing.T) {
	t.Run("should get transcript by episode ID", func(t *testing.T) {
		mockTranscript := Transcript{
			ID:        1,
			EpisodeID: 1,
			Status:    string(TranscriptStatusComplete),
			CreatedAt: time.Now(),
		}

		mockExecutor := utils.QueryExecutor[Transcript]{
			QueryItem: func(query string, args ...any) (Transcript, error) {
				return mockTranscript, nil
			},
		}

		repo := NewTranscriptRepository()
		repo.transcriptRepo.Executor = mockExecutor

		result, err := repo.GetTranscriptByEpisodeID(1)

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if result.ID != mockTranscript.ID {
			t.Errorf("Expected transcript ID %v, got %v", mockTranscript.ID, result.ID)
		}

		if result.EpisodeID != mockTranscript.EpisodeID {
			t.Errorf("Expected episode ID %v, got %v", mockTranscript.EpisodeID, result.EpisodeID)
		}
	})

	t.Run("should return error when transcript not found", func(t *testing.T) {
		mockExecutor := utils.QueryExecutor[Transcript]{
			QueryItem: func(query string, args ...any) (Transcript, error) {
				return Transcript{}, errors.New("no rows in result set")
			},
		}

		repo := NewTranscriptRepository()
		repo.transcriptRepo.Executor = mockExecutor

		_, err := repo.GetTranscriptByEpisodeID(999)

		if err == nil {
			t.Error("Expected error for non-existent transcript, got nil")
		}
	})
}

func TestTranscriptRepository_GetEpisodeByID(t *testing.T) {
	t.Run("should get episode by ID", func(t *testing.T) {
		mockEpisode := Episode{
			ID:          1,
			AudioURL:    "https://example.com/audio.mp3",
			Description: "Test episode",
		}

		mockExecutor := utils.QueryExecutor[Episode]{
			QueryItem: func(query string, args ...any) (Episode, error) {
				return mockEpisode, nil
			},
		}

		repo := NewTranscriptRepository()
		repo.episodeRepo.Executor = mockExecutor

		result, err := repo.GetEpisodeByID(1)

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if result.ID != mockEpisode.ID {
			t.Errorf("Expected episode ID %v, got %v", mockEpisode.ID, result.ID)
		}

		if result.AudioURL != mockEpisode.AudioURL {
			t.Errorf("Expected audio URL %v, got %v", mockEpisode.AudioURL, result.AudioURL)
		}
	})
}

func TestTranscriptRepository_UpdateTranscriptStatus(t *testing.T) {
	t.Run("should update transcript status to complete", func(t *testing.T) {
		mockExecutor := utils.QueryExecutor[Transcript]{
			Exec: func(query string, args ...any) error {
				return nil
			},
		}

		repo := NewTranscriptRepository()
		repo.transcriptRepo.Executor = mockExecutor

		err := repo.UpdateTranscriptStatus(1, TranscriptStatusComplete, "")

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})

	t.Run("should update transcript status to failed with error message", func(t *testing.T) {
		mockExecutor := utils.QueryExecutor[Transcript]{
			Exec: func(query string, args ...any) error {
				return nil
			},
		}

		repo := NewTranscriptRepository()
		repo.transcriptRepo.Executor = mockExecutor

		err := repo.UpdateTranscriptStatus(1, TranscriptStatusFailed, "Test error")

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})

	t.Run("should return error when update fails", func(t *testing.T) {
		mockExecutor := utils.QueryExecutor[Transcript]{
			Exec: func(query string, args ...any) error {
				return errors.New("database error")
			},
		}

		repo := NewTranscriptRepository()
		repo.transcriptRepo.Executor = mockExecutor

		err := repo.UpdateTranscriptStatus(1, TranscriptStatusComplete, "")

		if err == nil {
			t.Error("Expected error, got nil")
		}

		if err.Error() != "database error" {
			t.Errorf("Expected 'database error', got %v", err.Error())
		}
	})
}

func TestTranscriptRepository_CreateTranscript(t *testing.T) {
	t.Run("should create transcript successfully", func(t *testing.T) {
		mockExecutor := utils.QueryExecutor[Transcript]{
			QueryItem: func(query string, args ...any) (Transcript, error) {
				return Transcript{ID: 1, EpisodeID: 1}, nil
			},
		}

		repo := NewTranscriptRepository()
		repo.transcriptRepo.Executor = mockExecutor

		id, err := repo.CreateTranscript(1)

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if id != 1 {
			t.Errorf("Expected transcript ID 1, got %v", id)
		}
	})
}

func TestTranscriptRepository_GetSpeakersByTranscriptID(t *testing.T) {
	t.Run("should get speakers successfully", func(t *testing.T) {
		mockSpeakers := []TranscriptSpeaker{
			{SpeakerIndex: 0, SpeakerName: "Speaker 0"},
			{SpeakerIndex: 1, SpeakerName: "Speaker 1"},
		}

		mockExecutor := utils.QueryExecutor[TranscriptSpeaker]{
			QueryList: func(query string, args ...any) ([]TranscriptSpeaker, error) {
				return mockSpeakers, nil
			},
		}

		repo := NewTranscriptRepository()
		repo.speakerRepo.Executor = mockExecutor

		speakers, err := repo.GetSpeakersByTranscriptID(1)

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if len(speakers) != 2 {
			t.Errorf("Expected 2 speakers, got %v", len(speakers))
		}
	})
}

func TestTranscriptRepository_GetChunksByTranscriptID(t *testing.T) {
	t.Run("should get chunks successfully", func(t *testing.T) {
		mockChunks := []TranscriptChunk{
			{Position: 0, Text: "Hello", StartTime: 0.0, EndTime: 0.5, SpeakerIndex: 0},
			{Position: 1, Text: "world", StartTime: 0.5, EndTime: 1.0, SpeakerIndex: 0},
		}

		mockExecutor := utils.QueryExecutor[TranscriptChunk]{
			QueryList: func(query string, args ...any) ([]TranscriptChunk, error) {
				return mockChunks, nil
			},
		}

		repo := NewTranscriptRepository()
		repo.chunkRepo.Executor = mockExecutor

		chunks, err := repo.GetChunksByTranscriptID(1)

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if len(chunks) != 2 {
			t.Errorf("Expected 2 chunks, got %v", len(chunks))
		}
	})
}

func TestTranscriptRepository_SaveChunks(t *testing.T) {
	t.Run("should save chunks successfully", func(t *testing.T) {
		mockExecutor := utils.QueryExecutor[TranscriptChunk]{
			Exec: func(query string, args ...any) error {
				return nil
			},
		}

		repo := NewTranscriptRepository()
		repo.chunkRepo.Executor = mockExecutor

		chunks := []Chunk{
			{Position: 0, Text: "Hello", Start: 0.0, End: 0.5, SpeakerIndex: 0},
		}
		err := repo.SaveChunks(1, chunks)

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})

	t.Run("should return nil for empty chunks", func(t *testing.T) {
		repo := NewTranscriptRepository()
		err := repo.SaveChunks(1, []Chunk{})

		if err != nil {
			t.Errorf("Expected no error for empty chunks, got %v", err)
		}
	})
}

func TestTranscriptRepository_UpsertSpeaker(t *testing.T) {
	t.Run("should upsert speaker successfully", func(t *testing.T) {
		mockExecutor := utils.QueryExecutor[TranscriptSpeaker]{
			Exec: func(query string, args ...any) error {
				return nil
			},
		}

		repo := NewTranscriptRepository()
		repo.speakerRepo.Executor = mockExecutor

		err := repo.UpsertSpeaker(1, 0, "Speaker 0")

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})

	t.Run("should return error when upsert fails", func(t *testing.T) {
		mockExecutor := utils.QueryExecutor[TranscriptSpeaker]{
			Exec: func(query string, args ...any) error {
				return errors.New("database error")
			},
		}

		repo := NewTranscriptRepository()
		repo.speakerRepo.Executor = mockExecutor

		err := repo.UpsertSpeaker(1, 0, "Speaker 0")

		if err == nil {
			t.Error("Expected error, got nil")
		}

		if err.Error() != "database error" {
			t.Errorf("Expected 'database error', got %v", err.Error())
		}
	})
}
