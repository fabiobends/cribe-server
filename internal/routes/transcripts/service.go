package transcripts

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"

	"cribeapp.com/cribe-server/internal/clients/llm"
	"cribeapp.com/cribe-server/internal/clients/transcription"
	"cribeapp.com/cribe-server/internal/core/logger"
)

// TranscriptionClientInterface defines the contract for transcription clients
type TranscriptionClientInterface interface {
	StreamAudioURL(audioURL string, callback transcription.StreamCallback) error
}

// Service handles transcript business logic
type Service struct {
	repo                *TranscriptRepository
	transcriptionClient TranscriptionClientInterface
	llmClient           llm.LLMClient
	log                 *logger.ContextualLogger
}

// NewService creates a new transcript service
func NewService(transcriptionClient TranscriptionClientInterface, llmClient llm.LLMClient) *Service {
	return &Service{
		repo:                NewTranscriptRepository(),
		transcriptionClient: transcriptionClient,
		llmClient:           llmClient,
		log:                 logger.NewServiceLogger("TranscriptService"),
	}
}

// inferSpeakerName uses LLM to infer speaker name with proper timeout handling
func (s *Service) inferSpeakerName(episodeDescription string, speakerIndex int, transcriptChunks []string) (string, error) {
	// If LLM client is not available (nil interface or nil pointer), return default speaker name
	if s.llmClient == nil || reflect.ValueOf(s.llmClient).IsNil() {
		return fmt.Sprintf("Speaker %d", speakerIndex), nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Build context from transcript chunks
	var chunksText strings.Builder
	maxChunks := min(len(transcriptChunks), 200)
	for i := range maxChunks {
		chunksText.WriteString(transcriptChunks[i] + " ")
	}

	// Create request
	reqBody := llm.ChatRequest{
		Messages: []llm.Message{
			{Role: "system", Content: InferSpeakerNameSystemPrompt},
			{Role: "user", Content: InferSpeakerNameUserPrompt(episodeDescription, speakerIndex, chunksText.String())},
		},
		MaxTokens: 50,
	}

	response, err := s.llmClient.Chat(ctx, reqBody)
	if err != nil {
		return "", err
	}

	if len(response.Choices) == 0 {
		s.log.Warn("No choices returned from LLM", nil)
		return fmt.Sprintf("Speaker %d", speakerIndex), nil
	}

	// Extract and clean the name
	name := strings.TrimSpace(response.Choices[0].Message.Content)
	if name == "" || strings.HasPrefix(name, "Speaker") {
		return fmt.Sprintf("Speaker %d", speakerIndex), nil
	}

	return name, nil
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

	// Stream from transcription API and save to DB
	return s.streamFromTranscriptionAPI(episodeID, episode.AudioURL, episode.Description, chunkCB, speakerCB)
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

	s.log.Info("Streaming speakers from DB", map[string]any{
		"transcriptID": transcriptID,
		"speakerCount": len(speakers),
	})

	for _, speaker := range speakers {
		if err := speakerCB(&Speaker{
			Index: speaker.SpeakerIndex,
			Name:  speaker.SpeakerName,
		}); err != nil {
			s.log.Error("Failed to send speaker callback", map[string]any{
				"error":        err.Error(),
				"speakerIndex": speaker.SpeakerIndex,
			})
			return err
		}
	}

	// Then stream all chunks
	chunks, err := s.repo.GetChunksByTranscriptID(transcriptID)
	if err != nil {
		return fmt.Errorf("failed to query chunks: %w", err)
	}

	s.log.Info("Streaming chunks from DB", map[string]any{
		"transcriptID": transcriptID,
		"chunkCount":   len(chunks),
	})

	// For cached transcripts, send chunks in batches to reduce SSE events and prevent UI flickering
	const batchSize = 5
	for i := 0; i < len(chunks); i += batchSize {
		end := min(i+batchSize, len(chunks))

		// Send batch of chunks
		for _, chunk := range chunks[i:end] {
			if err := chunkCB(&Chunk{
				Position:     chunk.Position,
				SpeakerIndex: chunk.SpeakerIndex,
				Start:        chunk.StartTime,
				End:          chunk.EndTime,
				Text:         chunk.Text,
			}); err != nil {
				s.log.Error("Failed to send chunk callback", map[string]any{
					"error":    err.Error(),
					"position": chunk.Position,
				})
				return err
			}
		}

		// This prevents overwhelming the UI with rapid updates
		if end < len(chunks) {
			time.Sleep(20 * time.Millisecond)
		}
	}

	s.log.Info("Completed streaming from DB", map[string]any{
		"transcriptID":  transcriptID,
		"totalChunks":   len(chunks),
		"totalSpeakers": len(speakers),
	})

	return nil
}

// streamFromTranscriptionAPI streams from transcription API and saves to DB
func (s *Service) streamFromTranscriptionAPI(episodeId int, audioURL, episodeDesc string, chunkCB ChunkCallback, speakerCB SpeakerCallback) error {
	const (
		minSamplesForInference = 50 // Min words before inferring speaker name
	)

	var (
		chunks          []Chunk
		speakersSeen    = make(map[int]bool)
		speakerInferred = make(map[int]bool)
		position        = 0
		mu              sync.Mutex
		speakerChunks   = make(map[int][]string)
	)

	transcriptID, err := s.createTranscript(episodeId)

	if err != nil {
		s.log.Error("Failed to create transcript record", map[string]any{
			"error": err.Error(),
		})
		return fmt.Errorf("failed to create transcript record: %w", err)
	}

	// Stream from transcription API (client handles context and options internally)
	err = s.transcriptionClient.StreamAudioURL(audioURL, func(response *transcription.StreamResponse) error {
		if len(response.Channel.Alternatives) == 0 {
			return nil
		}

		words := response.Channel.Alternatives[0].Words
		if len(words) == 0 {
			return nil
		}

		for _, word := range words {
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
					s.log.Info("Client disconnected, continuing transcription", map[string]any{
						"error": err.Error(),
					})
				}
			}

			// Early speaker inference: when we have enough samples and haven't inferred yet
			if !speakerInferred[word.Speaker] && len(speakerChunks[word.Speaker]) >= minSamplesForInference {
				speakerInferred[word.Speaker] = true

				s.log.Info("Early inference for speaker", map[string]any{
					"speakerIndex": word.Speaker,
					"chunkCount":   len(speakerChunks[word.Speaker]),
					"threshold":    minSamplesForInference,
				})

				speakerIdx := word.Speaker
				contextWords := s.buildEarlySpeakerContext(chunks, speakerIdx)

				s.log.Info("Starting early inference goroutine", map[string]any{
					"speakerIndex": speakerIdx,
					"contextWords": len(contextWords),
				})

				go func(idx int, words []string) {
					name, err := s.inferSpeakerName(episodeDesc, idx, words)
					if err != nil {
						s.log.Error("Failed early speaker inference", map[string]any{
							"error":        err.Error(),
							"speakerIndex": idx,
						})
						return
					}

					s.log.Info("Success on speaker inference", map[string]any{
						"speakerIndex":    idx,
						"inferredName":    name,
						"sendingToClient": true,
					})

					if err := s.repo.UpsertSpeakerWithRetry(transcriptID, idx, name); err != nil {
						s.log.Error("Failed to save early inferred speaker", map[string]any{
							"error":        err.Error(),
							"speakerIndex": idx,
						})
						return
					}

					// Update client with real name
					if err := speakerCB(&Speaker{
						Index: idx,
						Name:  name,
					}); err != nil {
						s.log.Debug("Client disconnected during early inference", map[string]any{
							"error": err.Error(),
						})
					} else {
						s.log.Info("Speaker event sent to client", map[string]any{
							"speakerIndex": idx,
							"speakerName":  name,
						})
					}
				}(speakerIdx, contextWords)
			}
			mu.Unlock()

			if err := chunkCB(&chunk); err != nil {
				s.log.Info("Client disconnected, continuing transcription", map[string]any{
					"error": err.Error(),
				})
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

	// Streaming completed successfully, save to DB in background
	// Pass speakerInferred map to skip already-inferred speakers
	go s.saveTranscriptInBackground(transcriptID, chunks, speakerChunks, episodeDesc, speakerInferred, speakerCB)

	return nil
}

// saveTranscriptInBackground saves chunks and infers speaker names in the background
func (s *Service) saveTranscriptInBackground(transcriptID int, chunks []Chunk, speakerChunks map[int][]string, episodeDesc string, speakerInferred map[int]bool, speakerCB SpeakerCallback) {
	s.log.Info("Saving transcript to DB in background", map[string]any{
		"transcriptID": transcriptID,
		"totalChunks":  len(chunks),
	})

	// Save chunks to DB using batched inserts to reduce connection time
	if err := s.repo.SaveChunksBatched(transcriptID, chunks); err != nil {
		s.log.Error("Failed to save chunks", map[string]any{
			"error": err.Error(),
		})
		_ = s.repo.UpdateTranscriptStatus(transcriptID, string(TranscriptStatusFailed), err.Error())
		return
	}

	// Build context-aware speaker samples (includes words before/after speaker talks)
	speakerContexts := s.buildSpeakerContexts(chunks)

	// Infer speaker names in parallel (skip already-inferred speakers)
	var wg sync.WaitGroup
	speakersToInfer := 0
	for speakerIndex := range speakerContexts {
		if !speakerInferred[speakerIndex] {
			speakersToInfer++
		}
	}

	s.log.Info("Starting late speaker inference", map[string]any{
		"totalSpeakers":   len(speakerContexts),
		"speakersToInfer": speakersToInfer,
		"alreadyInferred": len(speakerContexts) - speakersToInfer,
	})

	for speakerIndex, contextWords := range speakerContexts {
		if speakerInferred[speakerIndex] {
			s.log.Debug("Skipping already-inferred speaker", map[string]any{
				"speakerIndex": speakerIndex,
			})
			continue // Already inferred during streaming
		}

		wg.Add(1)
		go func(idx int, words []string) {
			defer wg.Done()

			// Use background context to prevent cancellation
			speakerName, err := s.inferSpeakerName(episodeDesc, idx, words)
			if err != nil {
				s.log.Error("Failed to infer speaker name", map[string]any{
					"error":        err.Error(),
					"speakerIndex": idx,
				})
				return
			}

			// Update database with retry
			if err := s.repo.UpsertSpeakerWithRetry(transcriptID, idx, speakerName); err != nil {
				s.log.Error("Failed to save speaker", map[string]any{
					"error":        err.Error(),
					"speakerIndex": idx,
				})
				return
			}

			// Update client with real name
			if speakerCB != nil {
				if err := speakerCB(&Speaker{
					Index: idx,
					Name:  speakerName,
				}); err != nil {
					s.log.Debug("Client disconnected during late inference", map[string]any{
						"error": err.Error(),
					})
				}
			}
		}(speakerIndex, contextWords)
	}
	wg.Wait()

	// Update transcript status
	if err := s.repo.UpdateTranscriptStatus(transcriptID, string(TranscriptStatusComplete), ""); err != nil {
		s.log.Error("Failed to update transcript status", map[string]any{
			"error": err.Error(),
		})
		return
	}

	s.log.Info("Transcript saved successfully", map[string]any{
		"transcriptID": transcriptID,
	})
}

// buildSpeakerContexts creates context-aware samples for each speaker by including
// words spoken before, during, and after their speaking segments to capture name mentions
func (s *Service) buildSpeakerContexts(chunks []Chunk) map[int][]string {
	const (
		contextWindowBefore = 30 // words before speaker starts
		contextWindowAfter  = 30 // words after speaker finishes
		maxSamplesPerWindow = 3  // max speaker segments to sample
	)

	speakerContexts := make(map[int][]string)
	speakerSegments := make(map[int][][]int) // speaker -> list of [startIdx, endIdx]

	// Find all speaking segments for each speaker
	currentSpeaker := -1
	segmentStart := 0

	for i, chunk := range chunks {
		if chunk.SpeakerIndex != currentSpeaker {
			// Speaker changed
			if currentSpeaker != -1 {
				// Save previous segment
				speakerSegments[currentSpeaker] = append(speakerSegments[currentSpeaker], []int{segmentStart, i - 1})
			}
			currentSpeaker = chunk.SpeakerIndex
			segmentStart = i
		}
	}
	// Save last segment
	if currentSpeaker != -1 && len(chunks) > 0 {
		speakerSegments[currentSpeaker] = append(speakerSegments[currentSpeaker], []int{segmentStart, len(chunks) - 1})
	}

	// Build context for each speaker
	for speaker, segments := range speakerSegments {
		var contextWords []string

		// Sample up to maxSamplesPerWindow segments (beginning, middle, end)
		samplesToTake := min(len(segments), maxSamplesPerWindow)
		step := 1
		if len(segments) > maxSamplesPerWindow {
			step = len(segments) / maxSamplesPerWindow
		}

		for i := range samplesToTake {
			segmentIdx := i * step
			segment := segments[segmentIdx]
			startIdx := segment[0]
			endIdx := segment[1]

			// Include words BEFORE speaker starts
			beforeStart := max(0, startIdx-contextWindowBefore)
			for j := beforeStart; j < startIdx; j++ {
				contextWords = append(contextWords, chunks[j].Text)
			}

			// Include words DURING speaker talks
			for j := startIdx; j <= endIdx; j++ {
				contextWords = append(contextWords, chunks[j].Text)
			}

			// Include words AFTER speaker finishes
			afterEnd := min(len(chunks)-1, endIdx+contextWindowAfter)
			for j := endIdx + 1; j <= afterEnd; j++ {
				contextWords = append(contextWords, chunks[j].Text)
			}
		}

		speakerContexts[speaker] = contextWords
	}

	return speakerContexts
}

// buildEarlySpeakerContext creates context for a speaker during streaming (simpler than full context)
func (s *Service) buildEarlySpeakerContext(chunks []Chunk, targetSpeaker int) []string {
	const contextWindow = 30 // words around speaker segments

	var contextWords []string

	for i, chunk := range chunks {
		if chunk.SpeakerIndex == targetSpeaker {
			// Include words before
			beforeStart := max(0, i-contextWindow)
			for j := beforeStart; j < i; j++ {
				contextWords = append(contextWords, chunks[j].Text)
			}

			// Include this word
			contextWords = append(contextWords, chunk.Text)

			// Include words after
			afterEnd := min(len(chunks), i+contextWindow)
			for j := i + 1; j < afterEnd; j++ {
				contextWords = append(contextWords, chunks[j].Text)
			}

			break // Just get first occurrence with context
		}
	}

	return contextWords
}
