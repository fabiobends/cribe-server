package transcripts

import (
	"context"
	"fmt"
	"sync"

	"cribeapp.com/cribe-server/internal/clients/llm"
	"cribeapp.com/cribe-server/internal/clients/transcription"
	"cribeapp.com/cribe-server/internal/core/logger"
)

// TranscriptionClientInterface defines the contract for transcription clients
type TranscriptionClientInterface interface {
	StreamAudioURL(ctx context.Context, audioURL string, opts transcription.StreamOptions, callback transcription.StreamCallback) error
}

// LLMClientInterface defines the contract for LLM clients
type LLMClientInterface interface {
	InferSpeakerName(ctx context.Context, episodeDescription string, speakerIndex int, transcriptChunks []string) (string, error)
}

// Service handles transcript business logic
type Service struct {
	repo                *TranscriptRepository
	transcriptionClient TranscriptionClientInterface
	llmClient           LLMClientInterface
	log                 *logger.ContextualLogger
}

// NewServiceReady creates a new transcript service with environment configuration
func NewServiceReady() *Service {

	transcriptionClient := transcription.NewClient()
	llmClient := llm.NewClient()

	return NewService(transcriptionClient, llmClient)
}

// NewService creates a new transcript service
func NewService(transcriptionClient TranscriptionClientInterface, llmClient LLMClientInterface) *Service {
	return &Service{
		repo:                NewTranscriptRepository(),
		transcriptionClient: transcriptionClient,
		llmClient:           llmClient,
		log:                 logger.NewServiceLogger("TranscriptService"),
	}
}

// ChunkCallback is called for each transcript chunk
type ChunkCallback func(chunk *Chunk) error

// SpeakerCallback is called when a speaker is identified
type SpeakerCallback func(speaker *Speaker) error

// StreamTranscript streams a transcript for an episode
func (s *Service) StreamTranscript(ctx context.Context, episodeID int, chunkCB ChunkCallback, speakerCB SpeakerCallback) error {
	s.log.Info("Starting transcript stream", map[string]any{
		"episodeID": episodeID,
	})

	// Check if transcript already exists in DB
	transcriptID, exists, err := s.getExistingTranscript(episodeID)
	if err != nil {
		return fmt.Errorf("failed to check existing transcript: %w", err)
	}

	if exists {
		s.log.Info("Streaming cached transcript from DB", map[string]any{
			"episodeID":    episodeID,
			"transcriptID": transcriptID,
		})
		return s.streamFromDB(transcriptID, chunkCB, speakerCB)
	}

	// Get episode info from DB
	episode, err := s.repo.GetEpisodeByID(episodeID)
	if err != nil {
		return fmt.Errorf("failed to get episode: %w", err)
	}

	s.log.Info("Streaming from transcription API", map[string]any{
		"audioURL": episode.AudioURL,
	})

	// Create new transcript record
	transcriptID, err = s.createTranscript(episodeID)
	if err != nil {
		return fmt.Errorf("failed to create transcript: %w", err)
	}

	// Stream from transcription API and save to DB
	return s.streamFromTranscriptionAPI(ctx, transcriptID, episode.AudioURL, episode.Description, chunkCB, speakerCB)
}

// getExistingTranscript checks if a transcript exists for the episode
func (s *Service) getExistingTranscript(episodeID int) (int, bool, error) {
	transcript, err := s.repo.GetTranscriptByEpisodeID(episodeID)

	if err != nil {
		if err.Error() == "no rows in result set" {
			return 0, false, nil
		}
		return 0, false, err
	}

	// Only return existing if complete
	return transcript.ID, transcript.Status == string(TranscriptStatusComplete), nil
}

// createTranscript creates a new transcript record
func (s *Service) createTranscript(episodeID int) (int, error) {
	return s.repo.CreateTranscript(episodeID)
}

// streamFromDB streams a cached transcript from the database
func (s *Service) streamFromDB(transcriptID int, chunkCB ChunkCallback, speakerCB SpeakerCallback) error {
	// First, send all speakers
	speakers, err := s.repo.GetSpeakersByTranscriptID(transcriptID)
	if err != nil {
		return fmt.Errorf("failed to get speakers: %w", err)
	}

	for _, speaker := range speakers {
		if err := speakerCB(&Speaker{
			Index: speaker.SpeakerIndex,
			Name:  speaker.SpeakerName,
		}); err != nil {
			return err
		}
	}

	// Then stream all chunks
	chunks, err := s.repo.GetChunksByTranscriptID(transcriptID)
	if err != nil {
		return fmt.Errorf("failed to query chunks: %w", err)
	}

	for _, chunk := range chunks {
		if err := chunkCB(&Chunk{
			Position:     chunk.Position,
			SpeakerIndex: chunk.SpeakerIndex,
			Start:        chunk.StartTime,
			End:          chunk.EndTime,
			Text:         chunk.Text,
		}); err != nil {
			return err
		}
	}

	return nil
}

// streamFromTranscriptionAPI streams from transcription API and saves to DB
func (s *Service) streamFromTranscriptionAPI(ctx context.Context, transcriptID int, audioURL, episodeDesc string, chunkCB ChunkCallback, speakerCB SpeakerCallback) error {
	var (
		chunks        []Chunk
		speakersSeen  = make(map[int]bool)
		position      = 0
		mu            sync.Mutex
		speakerChunks = make(map[int][]string)
	)

	// Stream from transcription API
	err := s.transcriptionClient.StreamAudioURL(ctx, audioURL, transcription.StreamOptions{
		Model:      "nova-3",
		Language:   "en",
		Diarize:    true,
		Punctuate:  true,
		Utterances: false,
	}, func(response *transcription.StreamResponse) error {
		// Skip responses without alternatives (empty chunks)
		if len(response.Results.Channels) == 0 || len(response.Results.Channels[0].Alternatives) == 0 {
			return nil
		}

		alt := response.Results.Channels[0].Alternatives[0]

		for _, word := range alt.Words {
			chunk := Chunk{
				Position:     position,
				SpeakerIndex: word.Speaker,
				Start:        word.Start,
				End:          word.End,
				Text:         word.PunctuatedWord,
			}

			mu.Lock()
			chunks = append(chunks, chunk)
			position++

			speakerChunks[word.Speaker] = append(speakerChunks[word.Speaker], word.PunctuatedWord)

			if !speakersSeen[word.Speaker] {
				speakersSeen[word.Speaker] = true

				if err := speakerCB(&Speaker{
					Index: word.Speaker,
					Name:  fmt.Sprintf("Speaker %d", word.Speaker),
				}); err != nil {
					mu.Unlock()
					return err
				}
			}
			mu.Unlock()

			if err := chunkCB(&chunk); err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		_ = s.repo.UpdateTranscriptStatus(transcriptID, string(TranscriptStatusFailed), err.Error())
		s.log.Error("Transcription streaming error", map[string]any{
			"error": err.Error(),
		})
		return fmt.Errorf("transcription streaming error: %w", err)
	}

	s.log.Info("Transcription completed, saving to DB", map[string]any{
		"totalChunks": len(chunks),
	})

	// Save chunks to DB
	if err := s.repo.SaveChunks(transcriptID, chunks); err != nil {
		return fmt.Errorf("failed to save chunks: %w", err)
	}

	// Infer speaker names in parallel
	var wg sync.WaitGroup
	for speakerIndex, speakerWords := range speakerChunks {
		wg.Add(1)
		go func(idx int, words []string) {
			defer wg.Done()

			// Use background context to prevent cancellation when client disconnects
			speakerName, err := s.llmClient.InferSpeakerName(context.Background(), episodeDesc, idx, words)
			if err != nil {
				s.log.Error("Failed to infer speaker name", map[string]any{
					"error":        err.Error(),
					"speakerIndex": idx,
				})
				return
			}

			// Update database
			if err := s.repo.UpsertSpeaker(transcriptID, idx, speakerName); err != nil {
				s.log.Error("Failed to save speaker", map[string]any{
					"error": err.Error(),
				})
				return
			}

			// Send updated speaker name to client (will fail gracefully if disconnected)
			if err := speakerCB(&Speaker{
				Index: idx,
				Name:  speakerName,
			}); err != nil {
				s.log.Error("Failed to send speaker update", map[string]any{
					"error": err.Error(),
				})
				return
			}

			s.log.Info("Speaker name inferred and updated", map[string]any{
				"speakerIndex": idx,
				"speakerName":  speakerName,
			})
		}(speakerIndex, speakerWords)
	}

	// Wait for all speaker inferences to complete
	wg.Wait()

	// Mark as complete after all speaker inference
	if err := s.repo.UpdateTranscriptStatus(transcriptID, string(TranscriptStatusComplete), ""); err != nil {
		return err
	}

	return nil
}
