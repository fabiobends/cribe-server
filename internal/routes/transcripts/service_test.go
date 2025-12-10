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
	err := service.streamFromTranscriptionAPI(1, "test.mp3", "Test episode",
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

func (m *customMockTranscriptionClient) StreamAudioURL(audioURL string, callback transcription.StreamCallback) error {
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
				if status, ok := args[0].(string); ok && status == string(TranscriptStatusFailed) {
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
		func(speaker *Speaker) error { return nil },
	)

	if statusUpdateError != "database connection error" {
		t.Errorf("got error '%s', want 'database connection error'", statusUpdateError)
	}
}
