package quizzes

import (
	"database/sql"
	"fmt"
	"time"

	"cribeapp.com/cribe-server/internal/clients/llm"
	"cribeapp.com/cribe-server/internal/core/logger"
	"cribeapp.com/cribe-server/internal/errors"
	"cribeapp.com/cribe-server/internal/routes/transcripts"
)

const (
	// Question generation settings
	TotalQuestionsPerEpisode = 3
	MultipleChoiceCount      = 1
	TrueFalseCount           = 1
	OpenEndedCount           = 1
	MultipleChoiceOptions    = 4
)

// TranscriptRepo interface defines methods needed from transcript repository
type TranscriptRepo interface {
	GetTranscriptByEpisodeID(episodeID int) (transcripts.Transcript, error)
	GetChunksByTranscriptID(transcriptID int) ([]transcripts.TranscriptChunk, error)
}

type QuizService struct {
	repo           QuizRepository
	transcriptRepo TranscriptRepo
	llmClient      llm.LLMClient
	logger         *logger.ContextualLogger
}

func NewQuizService(repo QuizRepository, transcriptRepo TranscriptRepo, llmClient llm.LLMClient) *QuizService {
	return &QuizService{
		repo:           repo,
		transcriptRepo: transcriptRepo,
		llmClient:      llmClient,
		logger:         logger.NewServiceLogger("QuizService"),
	}
}

// GetSessionWithDetailsByID retrieves a quiz session with its questions and answers
func (s *QuizService) GetSessionWithDetailsByID(sessionID int) (QuizSessionDetail, *errors.ErrorResponse) {
	s.logger.Info("Fetching session details", map[string]any{
		"session_id": sessionID,
	})

	session, err := s.repo.GetSessionByID(sessionID)
	if err != nil {
		return QuizSessionDetail{}, &errors.ErrorResponse{
			Message: errors.DatabaseNotFound,
			Details: "Session not found",
		}
	}

	questions, err := s.repo.GetQuestionsByEpisodeID(session.EpisodeID)
	if err != nil {
		return QuizSessionDetail{}, &errors.ErrorResponse{
			Message: errors.DatabaseError,
			Details: "Failed to fetch questions",
		}
	}

	answers, err := s.repo.GetAnswersBySessionID(session.ID)
	if err != nil {
		return QuizSessionDetail{}, &errors.ErrorResponse{
			Message: errors.DatabaseError,
			Details: "Failed to fetch answers",
		}
	}

	return QuizSessionDetail{
		Session:   session,
		Questions: questions,
		Answers:   answers,
	}, nil
}

// GetSessionsWithDetailsByUserID retrieves all quiz sessions with details for a user
func (s *QuizService) GetSessionsWithDetailsByUserID(userID int) ([]QuizSessionDetail, *errors.ErrorResponse) {
	result := []QuizSessionDetail{}
	sessions, err := s.repo.GetSessionsByUserID(userID)
	if err != nil {
		return nil, &errors.ErrorResponse{
			Message: errors.DatabaseError,
			Details: "Failed to fetch sessions",
		}
	}

	for _, session := range sessions {
		questions, err := s.repo.GetQuestionsByEpisodeID(session.EpisodeID)
		if err != nil && err != sql.ErrNoRows {
			s.logger.Error("Failed to fetch questions for session", map[string]any{
				"session_id": session.ID,
				"error":      err.Error(),
			})
			continue
		}

		answers, err := s.repo.GetAnswersBySessionID(session.ID)
		if err != nil && err != sql.ErrNoRows {
			s.logger.Error("Failed to fetch answers for session", map[string]any{
				"session_id": session.ID,
				"error":      err.Error(),
			})
			continue
		}

		result = append(result, QuizSessionDetail{
			Session:   session,
			Questions: questions,
			Answers:   answers,
		})
	}

	return result, nil
}

// StartSession starts a new quiz session for a user
func (s *QuizService) GetOrCreateSessionWithDetails(userID int, episodeID int) (QuizSessionDetail, *errors.ErrorResponse) {
	s.logger.Info("Starting quiz session", map[string]any{
		"user_id":    userID,
		"episode_id": episodeID,
	})

	// Get questions count
	questions, _ := s.repo.GetQuestionsByEpisodeID(episodeID)

	if len(questions) == 0 {
		s.logger.Error("No questions available for episode", map[string]any{
			"episode_id": episodeID,
		})

		transcriptText, errResp := s.getTranscriptText(episodeID)
		if errResp != nil {
			s.logger.Error("Failed to get transcript text for question generation", map[string]any{
				"error": errResp.Message,
			})
			return QuizSessionDetail{}, errResp
		}

		// Generate questions using LLM
		llmQuestions, err := s.generateQuestionsWithLLM(transcriptText)
		if err != nil {
			s.logger.Error("Failed to generate questions with LLM", map[string]any{
				"error": err.Error(),
			})
			return QuizSessionDetail{}, &errors.ErrorResponse{
				Message: errors.InternalServerError,
				Details: "Failed to generate questions",
			}
		}

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
				s.logger.Error("Failed to save generated question", map[string]any{
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
			questions = append(questions, savedQuestion)
		}
	}

	session, err := s.repo.GetActiveSessionByUserAndEpisode(userID, episodeID)

	if err != nil {
		s.logger.Error("Failed to check for existing session", map[string]any{
			"error": err.Error(),
		})
		session, err := s.repo.CreateSession(UserQuizSession{
			UserID:            userID,
			EpisodeID:         episodeID,
			Status:            InProgress,
			TotalQuestions:    len(questions),
			AnsweredQuestions: 0,
			CorrectAnswers:    0,
		})
		if err != nil {
			s.logger.Error("Failed to create session", map[string]any{
				"error": err.Error(),
			})
			return QuizSessionDetail{}, &errors.ErrorResponse{
				Message: errors.DatabaseError,
				Details: "Failed to create session",
			}
		}
		return QuizSessionDetail{
			Session:   session,
			Questions: questions,
			Answers:   []UserAnswer{},
		}, nil
	}

	answers, _ := s.repo.GetAnswersBySessionID(session.ID)

	return QuizSessionDetail{
		Session:   session,
		Questions: questions,
		Answers:   answers,
	}, nil
}

// UpdateSessionStatus updates the session status (complete/abandon)
func (s *QuizService) UpdateSessionStatus(sessionID, userID int, status SessionStatus) (UserQuizSession, *errors.ErrorResponse) {
	// Get session and verify ownership
	session, errResp := s.repo.GetSessionByID(sessionID)
	if errResp != nil {
		return UserQuizSession{}, &errors.ErrorResponse{
			Message: errors.DatabaseNotFound,
			Details: "Session not found",
		}
	}

	// Verify session ownership
	if session.UserID != userID {
		return UserQuizSession{}, &errors.ErrorResponse{
			Message: errors.Unauthorized,
			Details: "User does not own this session",
		}
	}

	// Update status
	session.Status = status
	if status == Completed || status == Abandoned {
		now := time.Now()
		session.CompletedAt = &now
	}

	if err := s.repo.UpdateSession(session); err != nil {
		s.logger.Error("Failed to update session status", map[string]any{
			"session_id": sessionID,
			"error":      err.Error(),
		})
		return UserQuizSession{}, &errors.ErrorResponse{
			Message: errors.DatabaseError,
			Details: "Failed to update session status",
		}
	}

	s.logger.Info("Session status updated", map[string]any{
		"session_id": sessionID,
		"status":     status,
	})

	return session, nil
}

// SubmitAnswer submits an answer for a question
func (s *QuizService) SubmitAnswer(sessionID, userID, questionID int, request SubmitAnswerRequest) (UserAnswer, *errors.ErrorResponse) {
	s.logger.Info("Submitting answer", map[string]any{
		"session_id":  sessionID,
		"question_id": questionID,
	})

	// Get session and verify ownership
	session, err := s.repo.GetSessionByID(sessionID)
	if err != nil {
		return UserAnswer{}, &errors.ErrorResponse{
			Message: errors.DatabaseNotFound,
			Details: "Session not found",
		}
	}

	// Verify session ownership
	if session.UserID != userID {
		return UserAnswer{}, &errors.ErrorResponse{
			Message: errors.Unauthorized,
			Details: "User does not own this session",
		}
	}

	// Check if session is still active
	if session.Status != InProgress {
		return UserAnswer{}, &errors.ErrorResponse{
			Message: errors.ValidationError,
			Details: "Cannot submit answers to a completed or abandoned session",
		}
	}

	// Check if already answered
	_, err = s.repo.GetAnswerBySessionAndQuestion(sessionID, questionID)
	if err == nil {
		return UserAnswer{}, &errors.ErrorResponse{
			Message: errors.ValidationError,
			Details: "Question already answered in this session",
		}
	}

	// Get question
	question, err := s.repo.GetQuestionByID(questionID)
	if err != nil {
		return UserAnswer{}, &errors.ErrorResponse{
			Message: errors.DatabaseNotFound,
			Details: "Question not found",
		}
	}

	// Verify question belongs to episode
	if question.EpisodeID != session.EpisodeID {
		return UserAnswer{}, &errors.ErrorResponse{
			Message: errors.ValidationError,
			Details: "Question does not belong to this session's episode",
		}
	}

	// Evaluate answer
	isCorrect, feedback, err := s.evaluateAnswer(question, request)
	if err != nil {
		return UserAnswer{}, &errors.ErrorResponse{
			Message: errors.InternalServerError,
			Details: fmt.Sprintf("Failed to evaluate answer: %v", err),
		}
	}

	// Save answer
	answer := UserAnswer{
		SessionID:        sessionID,
		QuestionID:       questionID,
		UserID:           userID,
		SelectedOptionID: request.SelectedOptionID,
		TextAnswer:       request.TextAnswer,
		IsCorrect:        isCorrect,
		Feedback:         feedback,
	}

	savedAnswer, err := s.repo.CreateAnswer(answer)
	if err != nil {
		return UserAnswer{}, &errors.ErrorResponse{
			Message: errors.DatabaseError,
			Details: "Failed to save answer",
		}
	}

	// Update session stats
	session.AnsweredQuestions++
	if isCorrect {
		session.CorrectAnswers++
	}

	// Auto-complete session if all questions answered
	if session.AnsweredQuestions >= session.TotalQuestions {
		session.Status = Completed
		now := time.Now()
		session.CompletedAt = &now
	}

	if err := s.repo.UpdateSession(session); err != nil {
		s.logger.Error("Failed to update session", map[string]any{
			"session_id": sessionID,
			"error":      err.Error(),
		})
		return UserAnswer{}, &errors.ErrorResponse{
			Message: errors.DatabaseError,
			Details: "Failed to update session stats",
		}
	}

	s.logger.Info("Answer submitted successfully", map[string]any{
		"answer_id":  savedAnswer.ID,
		"is_correct": isCorrect,
	})

	return savedAnswer, nil
}

// evaluateAnswer evaluates the user's answer and generates feedback
func (s *QuizService) evaluateAnswer(question Question, request SubmitAnswerRequest) (bool, string, error) {
	switch question.Type {
	case MultipleChoice, TrueFalse:
		// Validate that SelectedOptionID is provided
		if request.SelectedOptionID == nil {
			return false, "Selected option ID is required for multiple choice and true/false questions", nil
		}

		// Find the selected option
		var selectedOption *QuestionOption
		for _, opt := range question.Options {
			if opt.ID == *request.SelectedOptionID {
				selectedOption = &opt
				break
			}
		}

		if selectedOption == nil {
			return false, "Invalid option selected", nil
		}

		if selectedOption.IsCorrect {
			feedback := s.generateFeedbackWithLLM(question, selectedOption.OptionText, true)
			return true, feedback, nil
		}

		// Find correct option for feedback
		var correctOption *QuestionOption
		for _, opt := range question.Options {
			if opt.IsCorrect {
				correctOption = &opt
				break
			}
		}

		if correctOption != nil {
			feedback := s.generateFeedbackWithLLM(question, selectedOption.OptionText, false)
			return false, fmt.Sprintf("Incorrect. The correct answer is: %s. %s",
				correctOption.OptionText, feedback), nil
		}

		return false, "Incorrect answer.", nil

	case OpenEnded:
		// Validate that TextAnswer is provided
		if request.TextAnswer == nil {
			return false, "Text answer is required for open-ended questions", nil
		}

		// Use LLM to evaluate open-ended answer
		return s.evaluateOpenEndedAnswer(question, *request.TextAnswer)

	default:
		return false, "Unknown question type", fmt.Errorf("unknown question type: %s", question.Type)
	}
}
