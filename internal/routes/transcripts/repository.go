package transcripts

import (
	"fmt"
	"time"

	"cribeapp.com/cribe-server/internal/core/logger"
	"cribeapp.com/cribe-server/internal/utils"
)

type TranscriptRepository struct {
	transcriptRepo *utils.Repository[Transcript]
	chunkRepo      *utils.Repository[TranscriptChunk]
	speakerRepo    *utils.Repository[TranscriptSpeaker]
	episodeRepo    *utils.Repository[Episode]
	logger         *logger.ContextualLogger
}

func NewTranscriptRepository() *TranscriptRepository {
	return &TranscriptRepository{
		transcriptRepo: utils.NewRepository[Transcript](),
		chunkRepo:      utils.NewRepository[TranscriptChunk](),
		speakerRepo:    utils.NewRepository[TranscriptSpeaker](),
		episodeRepo:    utils.NewRepository[Episode](),
		logger:         logger.NewRepositoryLogger("TranscriptRepository"),
	}
}

func (r *TranscriptRepository) GetTranscriptByEpisodeID(episodeID int) (Transcript, error) {
	r.logger.Debug("Fetching transcript by episode ID", map[string]any{
		"episodeID": episodeID,
	})

	query := `SELECT id, episode_id, status, error_message, created_at, completed_at FROM transcripts WHERE episode_id = $1`
	result, err := r.transcriptRepo.Executor.QueryItem(query, episodeID)

	if err != nil {
		r.logger.Error("Failed to fetch transcript", map[string]any{
			"episodeID": episodeID,
			"error":     err.Error(),
		})
		return Transcript{}, err
	}

	r.logger.Debug("Transcript found", map[string]any{
		"episodeID":    episodeID,
		"transcriptID": result.ID,
		"status":       result.Status,
	})

	return result, nil
}

func (r *TranscriptRepository) GetEpisodeByID(episodeID int) (Episode, error) {
	r.logger.Debug("Fetching episode by ID", map[string]any{
		"episodeID": episodeID,
	})

	query := `SELECT id, audio_url, description FROM episodes WHERE id = $1`
	result, err := r.episodeRepo.Executor.QueryItem(query, episodeID)

	if err != nil {
		r.logger.Error("Failed to fetch episode", map[string]any{
			"episodeID": episodeID,
			"error":     err.Error(),
		})
		return Episode{}, err
	}

	r.logger.Debug("Episode found", map[string]any{
		"episodeID": episodeID,
	})

	return result, nil
}

func (r *TranscriptRepository) CreateTranscript(episodeID int) (int, error) {
	r.logger.Debug("Creating transcript record", map[string]any{
		"episodeID": episodeID,
	})

	query := `
		INSERT INTO transcripts (episode_id, status, created_at)
		VALUES ($1, 'processing', NOW())
		ON CONFLICT (episode_id) DO UPDATE
		SET status = 'processing'
		RETURNING id
	`

	rows, err := r.transcriptRepo.Executor.QueryItem(query, episodeID)
	if err != nil {
		r.logger.Error("Failed to create transcript", map[string]any{
			"episodeID": episodeID,
			"error":     err.Error(),
		})
		return 0, err
	}

	transcriptID := rows.ID
	r.logger.Info("Transcript created", map[string]any{
		"episodeID":    episodeID,
		"transcriptID": transcriptID,
	})

	return transcriptID, nil
}

func (r *TranscriptRepository) UpdateTranscriptStatus(transcriptID int, status string, errorMessage string) error {
	r.logger.Debug("Updating transcript status", map[string]any{
		"transcriptID": transcriptID,
		"status":       status,
	})

	var err error
	switch {
	case status == string(TranscriptStatusFailed) && errorMessage != "":
		err = r.transcriptRepo.Executor.Exec(
			`UPDATE transcripts SET status = $1, error_message = $2 WHERE id = $3`,
			status, errorMessage, transcriptID,
		)
	case status == string(TranscriptStatusComplete):
		err = r.transcriptRepo.Executor.Exec(
			`UPDATE transcripts SET status = $1, completed_at = NOW() WHERE id = $2`,
			string(TranscriptStatusComplete), transcriptID,
		)
	default:
		err = r.transcriptRepo.Executor.Exec(
			`UPDATE transcripts SET status = $1 WHERE id = $2`,
			status, transcriptID,
		)
	}

	if err != nil {
		r.logger.Error("Failed to update transcript status", map[string]any{
			"transcriptID": transcriptID,
			"status":       status,
			"error":        err.Error(),
		})
		return err
	}

	r.logger.Info("Transcript status updated", map[string]any{
		"transcriptID": transcriptID,
		"status":       status,
	})

	return nil
}

func (r *TranscriptRepository) GetSpeakersByTranscriptID(transcriptID int) ([]TranscriptSpeaker, error) {
	r.logger.Debug("Fetching speakers for transcript", map[string]any{
		"transcriptID": transcriptID,
	})

	query := `SELECT speaker_index, speaker_name FROM transcript_speakers WHERE transcript_id = $1 ORDER BY speaker_index`
	speakers, err := r.speakerRepo.Executor.QueryList(query, transcriptID)

	if err != nil {
		r.logger.Error("Failed to fetch speakers", map[string]any{
			"transcriptID": transcriptID,
			"error":        err.Error(),
		})
		return nil, err
	}

	r.logger.Debug("Speakers fetched", map[string]any{
		"transcriptID": transcriptID,
		"count":        len(speakers),
	})

	return speakers, nil
}

func (r *TranscriptRepository) GetChunksByTranscriptID(transcriptID int) ([]TranscriptChunk, error) {
	r.logger.Debug("Fetching chunks for transcript", map[string]any{
		"transcriptID": transcriptID,
	})

	query := `
		SELECT position, speaker_index, start_time, end_time, text
		FROM transcript_chunks
		WHERE transcript_id = $1
		ORDER BY position ASC
	`
	chunks, err := r.chunkRepo.Executor.QueryList(query, transcriptID)

	if err != nil {
		r.logger.Error("Failed to fetch chunks", map[string]any{
			"transcriptID": transcriptID,
			"error":        err.Error(),
		})
		return nil, err
	}

	r.logger.Debug("Chunks fetched", map[string]any{
		"transcriptID": transcriptID,
		"count":        len(chunks),
	})

	return chunks, nil
}

func (r *TranscriptRepository) SaveChunksBatched(transcriptID int, chunks []Chunk) error {
	if len(chunks) == 0 {
		return nil
	}

	r.logger.Debug("Saving chunks to database with batching", map[string]any{
		"transcriptID": transcriptID,
		"chunkCount":   len(chunks),
	})

	// Use batched inserts to reduce connection time
	const batchSize = 500
	for i := 0; i < len(chunks); i += batchSize {
		end := min(i+batchSize, len(chunks))
		batch := chunks[i:end]

		// Build multi-row INSERT
		query := `INSERT INTO transcript_chunks (transcript_id, position, speaker_index, start_time, end_time, text) VALUES `
		args := make([]any, 0, len(batch)*6)

		for j, chunk := range batch {
			if j > 0 {
				query += ", "
			}
			offset := j * 6
			query += fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d)", offset+1, offset+2, offset+3, offset+4, offset+5, offset+6)
			args = append(args, transcriptID, chunk.Position, chunk.SpeakerIndex, chunk.Start, chunk.End, chunk.Text)
		}
		query += " ON CONFLICT (transcript_id, position) DO NOTHING"

		err := r.chunkRepo.Executor.Exec(query, args...)
		if err != nil {
			r.logger.Error("Failed to save chunk batch", map[string]any{
				"transcriptID": transcriptID,
				"batchStart":   i,
				"batchEnd":     end,
				"error":        err.Error(),
			})
			return err
		}
	}

	r.logger.Info("Chunks saved successfully", map[string]any{
		"transcriptID": transcriptID,
		"chunkCount":   len(chunks),
	})

	return nil
}

func (r *TranscriptRepository) SaveChunks(transcriptID int, chunks []Chunk) error {
	if len(chunks) == 0 {
		return nil
	}

	r.logger.Debug("Saving chunks to database", map[string]any{
		"transcriptID": transcriptID,
		"chunkCount":   len(chunks),
	})

	for _, chunk := range chunks {
		err := r.chunkRepo.Executor.Exec(
			`INSERT INTO transcript_chunks (transcript_id, position, speaker_index, start_time, end_time, text)
			 VALUES ($1, $2, $3, $4, $5, $6)
			 ON CONFLICT (transcript_id, position) DO NOTHING`,
			transcriptID, chunk.Position, chunk.SpeakerIndex, chunk.Start, chunk.End, chunk.Text,
		)
		if err != nil {
			r.logger.Error("Failed to save chunk", map[string]any{
				"transcriptID": transcriptID,
				"position":     chunk.Position,
				"error":        err.Error(),
			})
			return err
		}
	}

	r.logger.Info("Chunks saved successfully", map[string]any{
		"transcriptID": transcriptID,
		"chunkCount":   len(chunks),
	})

	return nil
}

func (r *TranscriptRepository) UpsertSpeaker(transcriptID int, speakerIndex int, speakerName string) error {
	r.logger.Debug("Upserting speaker", map[string]any{
		"transcriptID": transcriptID,
		"speakerIndex": speakerIndex,
		"speakerName":  speakerName,
	})

	err := r.speakerRepo.Executor.Exec(
		`INSERT INTO transcript_speakers (transcript_id, speaker_index, speaker_name, inferred_at)
		 VALUES ($1, $2, $3, NOW())
		 ON CONFLICT (transcript_id, speaker_index)
		 DO UPDATE SET speaker_name = $3, inferred_at = NOW()`,
		transcriptID, speakerIndex, speakerName,
	)

	if err != nil {
		r.logger.Error("Failed to upsert speaker", map[string]any{
			"transcriptID": transcriptID,
			"speakerIndex": speakerIndex,
			"error":        err.Error(),
		})
		return err
	}

	r.logger.Info("Speaker upserted successfully", map[string]any{
		"transcriptID": transcriptID,
		"speakerIndex": speakerIndex,
		"speakerName":  speakerName,
	})

	return nil
}

func (r *TranscriptRepository) UpsertSpeakerWithRetry(transcriptID int, speakerIndex int, speakerName string) error {
	const maxRetries = 3
	const retryDelayMs = 100

	var lastErr error
	for attempt := 1; attempt <= maxRetries; attempt++ {
		err := r.UpsertSpeaker(transcriptID, speakerIndex, speakerName)
		if err == nil {
			return nil
		}

		lastErr = err
		if attempt < maxRetries {
			r.logger.Debug("Retrying speaker upsert", map[string]any{
				"attempt":      attempt,
				"transcriptID": transcriptID,
				"speakerIndex": speakerIndex,
			})
			// Simple exponential backoff
			time.Sleep(time.Duration(retryDelayMs*attempt) * time.Millisecond)
		}
	}

	return fmt.Errorf("failed after %d attempts: %w", maxRetries, lastErr)
}
