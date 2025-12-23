package quizzes

import (
	"fmt"
	"strings"
	"testing"

	"cribeapp.com/cribe-server/internal/routes/transcripts"
)

func TestQuizService_generateQuestions(t *testing.T) {
	mockTranscript := &MockTranscriptRepository{
		GetTranscriptByEpisodeIDFunc: func(id int) (transcripts.Transcript, error) {
			return transcripts.Transcript{ID: 1, EpisodeID: id, Status: "complete"}, nil
		},
		GetChunksByTranscriptIDFunc: func(id int) ([]transcripts.TranscriptChunk, error) {
			return []transcripts.TranscriptChunk{
				{TranscriptID: id, Text: "Speaker 1: Hello world"},
				{TranscriptID: id, Text: "Speaker 2: Hi there"},
			}, nil
		},
	}

	t.Run("success", func(t *testing.T) {
		svc := setupQuizService(&MockLLMClient{ChatResponse: makeLLMResponse(makeQuizJSON(3))}, mockTranscript)

		questions, errResp := svc.generateQuestions(1)
		if errResp != nil {
			t.Fatalf("unexpected error: %v", errResp.Details)
		}

		if len(questions) != 3 {
			t.Errorf("got %d questions, want 3", len(questions))
		}
	})

	t.Run("transcript not found", func(t *testing.T) {
		failingTranscript := &MockTranscriptRepository{
			GetTranscriptByEpisodeIDFunc: func(id int) (transcripts.Transcript, error) {
				return transcripts.Transcript{}, fmt.Errorf("not found")
			},
		}
		svc := setupQuizService(&MockLLMClient{}, failingTranscript)

		_, errResp := svc.generateQuestions(999)
		if errResp == nil || !strings.Contains(errResp.Details, "Failed to fetch transcript") {
			t.Errorf("got error %v, want error containing 'Failed to fetch transcript'", errResp)
		}
	})

	t.Run("incomplete transcript", func(t *testing.T) {
		incompleteTranscript := &MockTranscriptRepository{
			GetTranscriptByEpisodeIDFunc: func(id int) (transcripts.Transcript, error) {
				return transcripts.Transcript{ID: 1, Status: "processing"}, nil
			},
		}
		svc := setupQuizService(&MockLLMClient{}, incompleteTranscript)

		_, errResp := svc.generateQuestions(1)
		if errResp == nil || !strings.Contains(errResp.Details, "must be complete") {
			t.Errorf("got error %v, want error containing 'must be complete'", errResp)
		}
	})

	t.Run("LLM error", func(t *testing.T) {
		svc := setupQuizService(&MockLLMClient{ChatError: fmt.Errorf("LLM failed")}, mockTranscript)

		_, errResp := svc.generateQuestions(1)
		if errResp == nil || !strings.Contains(errResp.Details, "Failed to generate questions") {
			t.Errorf("got error %v, want error containing 'Failed to generate questions'", errResp)
		}
	})
}

func TestQuizService_getTranscriptText(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockTranscript := &MockTranscriptRepository{
			GetTranscriptByEpisodeIDFunc: func(id int) (transcripts.Transcript, error) {
				return transcripts.Transcript{ID: 42, EpisodeID: id, Status: "complete"}, nil
			},
			GetChunksByTranscriptIDFunc: func(id int) ([]transcripts.TranscriptChunk, error) {
				return []transcripts.TranscriptChunk{
					{Text: "Hello"},
					{Text: "World"},
				}, nil
			},
		}
		svc := setupQuizService(nil, mockTranscript)

		text, errResp := svc.getTranscriptText(1)
		if errResp != nil {
			t.Fatalf("unexpected error: %v", errResp.Details)
		}

		expected := "Hello World "
		if text != expected {
			t.Errorf("got %q, want %q", text, expected)
		}
	})

	t.Run("transcript not found", func(t *testing.T) {
		mockTranscript := &MockTranscriptRepository{
			GetTranscriptByEpisodeIDFunc: func(id int) (transcripts.Transcript, error) {
				return transcripts.Transcript{}, fmt.Errorf("not found")
			},
		}
		svc := setupQuizService(nil, mockTranscript)

		_, errResp := svc.getTranscriptText(999)
		if errResp == nil || !strings.Contains(errResp.Details, "Failed to fetch transcript") {
			t.Errorf("got error %v, want error containing 'Failed to fetch transcript'", errResp)
		}
	})

	t.Run("incomplete transcript", func(t *testing.T) {
		mockTranscript := &MockTranscriptRepository{
			GetTranscriptByEpisodeIDFunc: func(id int) (transcripts.Transcript, error) {
				return transcripts.Transcript{ID: 1, Status: "processing"}, nil
			},
		}
		svc := setupQuizService(nil, mockTranscript)

		_, errResp := svc.getTranscriptText(1)
		if errResp == nil || !strings.Contains(errResp.Details, "must be complete") {
			t.Errorf("got error %v, want error containing 'must be complete'", errResp)
		}
	})

	t.Run("empty chunks", func(t *testing.T) {
		mockTranscript := &MockTranscriptRepository{
			GetTranscriptByEpisodeIDFunc: func(id int) (transcripts.Transcript, error) {
				return transcripts.Transcript{ID: 1, Status: "complete"}, nil
			},
			GetChunksByTranscriptIDFunc: func(id int) ([]transcripts.TranscriptChunk, error) {
				return []transcripts.TranscriptChunk{}, nil
			},
		}
		svc := setupQuizService(nil, mockTranscript)

		text, errResp := svc.getTranscriptText(1)
		if errResp != nil {
			t.Fatalf("unexpected error: %v", errResp.Details)
		}

		if text != "" {
			t.Errorf("got %q, want empty string", text)
		}
	})

	t.Run("database error fetching chunks", func(t *testing.T) {
		mockTranscript := &MockTranscriptRepository{
			GetTranscriptByEpisodeIDFunc: func(id int) (transcripts.Transcript, error) {
				return transcripts.Transcript{ID: 1, EpisodeID: id, Status: "complete"}, nil
			},
			GetChunksByTranscriptIDFunc: func(id int) ([]transcripts.TranscriptChunk, error) {
				return nil, fmt.Errorf("database connection error")
			},
		}
		svc := setupQuizService(nil, mockTranscript)

		_, errResp := svc.getTranscriptText(1)
		if errResp == nil || !strings.Contains(errResp.Details, "Failed to fetch transcript chunks") {
			t.Errorf("got error %v, want error containing 'Failed to fetch transcript chunks'", errResp)
		}
	})
}
