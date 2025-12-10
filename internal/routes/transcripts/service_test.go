package transcripts

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"cribeapp.com/cribe-server/internal/clients/transcription"
	"cribeapp.com/cribe-server/internal/utils"
)

func TestTranscriptService_StreamTranscript(t *testing.T) {
	service := NewService(&MockTranscriptionClient{}, &MockLLMClient{})

	t.Run("should stream from DB when transcript exists", func(t *testing.T) {
		service.repo.transcriptRepo.Executor = utils.QueryExecutor[Transcript]{
			QueryItem: func(query string, args ...any) (Transcript, error) {
				return Transcript{ID: 1, Status: string(TranscriptStatusComplete)}, nil
			},
		}
		service.repo.speakerRepo.Executor = utils.QueryExecutor[TranscriptSpeaker]{
			QueryList: func(query string, args ...any) ([]TranscriptSpeaker, error) {
				return []TranscriptSpeaker{{SpeakerIndex: 0, SpeakerName: "Speaker 0"}}, nil
			},
		}
		service.repo.chunkRepo.Executor = utils.QueryExecutor[TranscriptChunk]{
			QueryList: func(query string, args ...any) ([]TranscriptChunk, error) {
				return []TranscriptChunk{{Position: 0, Text: "Hello"}}, nil
			},
		}

		chunkCount := 0
		err := service.StreamTranscript(context.Background(), 1,
			func(chunk *Chunk) error {
				chunkCount++
				return nil
			},
			func(speaker *Speaker) error { return nil },
		)

		if err != nil || chunkCount != 1 {
			t.Errorf("Expected no error and 1 chunk, got error=%v chunks=%d", err, chunkCount)
		}
	})

	t.Run("should stream from API when transcript not exists", func(t *testing.T) {
		service.repo.transcriptRepo.Executor = utils.QueryExecutor[Transcript]{
			QueryItem: func(query string, args ...any) (Transcript, error) {
				if strings.Contains(query, "WHERE episode_id") {
					return Transcript{}, fmt.Errorf("no rows in result set")
				}
				return Transcript{ID: 1}, nil
			},
			Exec: func(query string, args ...any) error { return nil },
		}
		service.repo.episodeRepo.Executor = utils.QueryExecutor[Episode]{
			QueryItem: func(query string, args ...any) (Episode, error) {
				return Episode{ID: 1, AudioURL: "test.mp3"}, nil
			},
		}
		service.repo.chunkRepo.Executor = utils.QueryExecutor[TranscriptChunk]{
			Exec: func(query string, args ...any) error { return nil },
		}
		service.repo.speakerRepo.Executor = utils.QueryExecutor[TranscriptSpeaker]{
			Exec: func(query string, args ...any) error { return nil },
		}

		chunkCount := 0
		err := service.StreamTranscript(context.Background(), 1,
			func(chunk *Chunk) error {
				chunkCount++
				return nil
			},
			func(speaker *Speaker) error { return nil },
		)

		if err != nil || chunkCount < 1 {
			t.Errorf("Expected no error and chunks, got error=%v chunks=%d", err, chunkCount)
		}
	})
}

type mockFailingTranscriptionClient struct{}

func (m *mockFailingTranscriptionClient) StreamAudioURL(audioURL string, callback transcription.StreamCallback) error {
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
			if len(args) >= 2 && args[0] == string(TranscriptStatusFailed) {
				statusUpdated = true
			}
			return nil
		},
	}

	err := service.streamFromTranscriptionAPI(1, "test.mp3", "Test",
		func(chunk *Chunk) error { return nil },
		func(speaker *Speaker) error { return nil },
	)

	if err == nil || !statusUpdated {
		t.Errorf("Expected error and status update, got error=%v updated=%v", err, statusUpdated)
	}
}

func TestTranscriptService_BuildEarlySpeakerContext(t *testing.T) {
	service := NewService(&MockTranscriptionClient{}, &MockLLMClient{})

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
		t.Errorf("Expected %d words, got %d", len(expected), len(contextWords))
	}
	for i, word := range expected {
		if contextWords[i] != word {
			t.Errorf("Position %d: expected '%s', got '%s'", i, word, contextWords[i])
		}
	}
}

func TestTranscriptService_EarlyInference(t *testing.T) {
	mockTranscriptionClient := &customMockTranscriptionClient{wordCount: 60}
	service := NewService(mockTranscriptionClient, &MockLLMClient{})

	service.repo.transcriptRepo.Executor = utils.QueryExecutor[Transcript]{
		QueryItem: func(query string, args ...any) (Transcript, error) {
			return Transcript{ID: 1}, nil
		},
		Exec: func(query string, args ...any) error { return nil },
	}
	service.repo.episodeRepo.Executor = utils.QueryExecutor[Episode]{
		QueryItem: func(query string, args ...any) (Episode, error) {
			return Episode{ID: 1, AudioURL: "test.mp3", Description: "Test"}, nil
		},
	}
	service.repo.chunkRepo.Executor = utils.QueryExecutor[TranscriptChunk]{
		Exec: func(query string, args ...any) error { return nil },
	}
	service.repo.speakerRepo.Executor = utils.QueryExecutor[TranscriptSpeaker]{
		Exec: func(query string, args ...any) error { return nil },
	}

	speakerCallbackCount := 0
	speakerNames := []string{}

	err := service.streamFromTranscriptionAPI(1, "test.mp3", "Test episode",
		func(chunk *Chunk) error { return nil },
		func(speaker *Speaker) error {
			speakerCallbackCount++
			speakerNames = append(speakerNames, speaker.Name)
			return nil
		},
	)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	if speakerCallbackCount < 2 {
		t.Errorf("Expected at least 2 speaker callbacks, got %d", speakerCallbackCount)
	}
	if len(speakerNames) >= 1 && speakerNames[0] != "Speaker 0" {
		t.Errorf("Expected first speaker name to be 'Speaker 0', got '%s'", speakerNames[0])
	}
}

type customMockTranscriptionClient struct {
	wordCount int
}

func (m *customMockTranscriptionClient) StreamAudioURL(audioURL string, callback transcription.StreamCallback) error {
	for i := 0; i < m.wordCount; i++ {
		response := &transcription.StreamResponse{
			Type: "Results",
			Channel: transcription.Channel{
				Alternatives: []transcription.Alternative{
					{
						Words: []transcription.Word{
							{
								Word:           fmt.Sprintf("word%d", i),
								PunctuatedWord: fmt.Sprintf("word%d", i),
								Start:          float64(i),
								End:            float64(i) + 0.5,
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
	}
	return nil
}
