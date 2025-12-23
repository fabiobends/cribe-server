package quizzes

import (
	"database/sql"
	"fmt"
	"strings"

	"cribeapp.com/cribe-server/internal/errors"
)

// getTranscriptText retrieves and compiles the transcript text for an episode
func (s *QuizService) getTranscriptText(episodeID int) (string, *errors.ErrorResponse) {
	// Get transcript
	transcript, err := s.transcriptRepo.GetTranscriptByEpisodeID(episodeID)
	if err != nil {
		s.logger.Error("Failed to fetch transcript", map[string]any{
			"episode_id": episodeID,
			"error":      err.Error(),
		})
		return "", &errors.ErrorResponse{
			Message: errors.DatabaseNotFound,
			Details: "Failed to fetch transcript for this episode",
		}
	}

	// Check transcript status
	if transcript.Status != "complete" {
		return "", &errors.ErrorResponse{
			Message: errors.ValidationError,
			Details: "Transcript must be complete before generating questions",
		}
	}

	// Get transcript chunks
	chunks, err := s.transcriptRepo.GetChunksByTranscriptID(transcript.ID)
	if err != nil && err != sql.ErrNoRows {
		s.logger.Error("Failed to fetch transcript chunks", map[string]any{
			"transcript_id": transcript.ID,
			"error":         err.Error(),
		})
		return "", &errors.ErrorResponse{
			Message: errors.DatabaseError,
			Details: "Failed to fetch transcript chunks",
		}
	}

	// Compile transcript text
	var transcriptText strings.Builder
	for _, chunk := range chunks {
		transcriptText.WriteString(chunk.Text)
		transcriptText.WriteString(" ")
	}

	return transcriptText.String(), nil
}

// generateQuestions generates quiz questions for an episode using LLM
func (s *QuizService) generateQuestions(episodeID int) ([]Question, *errors.ErrorResponse) {
	s.logger.Info("Generating questions for episode", map[string]any{
		"episode_id": episodeID,
	})

	transcriptText, errResp := s.getTranscriptText(episodeID)
	if errResp != nil {
		return nil, errResp
	}

	// Delete existing questions for this episode (regenerate)
	if err := s.repo.DeleteQuestionsByEpisodeID(episodeID); err != nil {
		s.logger.Error("Failed to delete existing questions", map[string]any{
			"episode_id": episodeID,
			"error":      err.Error(),
		})
		return nil, &errors.ErrorResponse{
			Message: errors.DatabaseError,
			Details: "Failed to delete existing questions before regeneration",
		}
	}

	// Generate questions using LLM
	llmQuestions, err := s.generateQuestionsWithLLM(transcriptText)
	if err != nil {
		return nil, &errors.ErrorResponse{
			Message: errors.InternalServerError,
			Details: fmt.Sprintf("Failed to generate questions: %v", err),
		}
	}

	// Save questions to database
	var savedQuestions []Question
	for i, llmQ := range llmQuestions {
		// Validate question type before creating
		questionType := QuestionType(llmQ.Type)
		if questionType != MultipleChoice && questionType != TrueFalse && questionType != OpenEnded {
			s.logger.Error("Invalid question type from LLM, skipping question", map[string]any{
				"type":          llmQ.Type,
				"question_text": llmQ.QuestionText,
			})
			continue
		}

		question := Question{
			EpisodeID:    episodeID,
			QuestionText: llmQ.QuestionText,
			Type:         questionType,
			Position:     i,
		}

		savedQuestion, err := s.repo.CreateQuestion(question)
		if err != nil {
			s.logger.Error("Failed to save question", map[string]any{
				"error": err.Error(),
			})
			continue
		}

		// Save options if applicable
		var savedOptions []QuestionOption
		if questionType == MultipleChoice || questionType == TrueFalse {
			for j, opt := range llmQ.Options {
				option := QuestionOption{
					QuestionID: savedQuestion.ID,
					OptionText: opt.Text,
					Position:   j,
					IsCorrect:  opt.IsCorrect,
				}

				savedOption, err := s.repo.CreateQuestionOption(option)
				if err != nil {
					s.logger.Error("Failed to save option", map[string]any{
						"error": err.Error(),
					})
					continue
				}
				savedOptions = append(savedOptions, savedOption)
			}
		}

		savedQuestion.Options = savedOptions
		savedQuestions = append(savedQuestions, savedQuestion)
	}

	s.logger.Info("Questions generated successfully", map[string]any{
		"episode_id": episodeID,
		"count":      len(savedQuestions),
	})

	return savedQuestions, nil
}
