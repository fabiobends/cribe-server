package quizzes

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"cribeapp.com/cribe-server/internal/clients/llm"
	"cribeapp.com/cribe-server/internal/core/logger"
	"cribeapp.com/cribe-server/internal/routes/transcripts"
	"cribeapp.com/cribe-server/internal/utils"
)

// Mock LLM Client
type MockLLMClient struct {
	ChatResponse llm.ChatCompletionResponse
	ChatError    error
}

func (m *MockLLMClient) Chat(ctx context.Context, req llm.ChatRequest) (llm.ChatCompletionResponse, error) {
	if m.ChatError != nil {
		return llm.ChatCompletionResponse{}, m.ChatError
	}
	return m.ChatResponse, nil
}

// Mock Transcript Repository
type MockTranscriptRepository struct {
	GetTranscriptByEpisodeIDFunc func(episodeID int) (transcripts.Transcript, error)
	GetChunksByTranscriptIDFunc  func(transcriptID int) ([]transcripts.TranscriptChunk, error)
}

func (m *MockTranscriptRepository) GetTranscriptByEpisodeID(episodeID int) (transcripts.Transcript, error) {
	if m.GetTranscriptByEpisodeIDFunc != nil {
		return m.GetTranscriptByEpisodeIDFunc(episodeID)
	}
	return transcripts.Transcript{}, fmt.Errorf("not implemented")
}

func (m *MockTranscriptRepository) GetChunksByTranscriptID(transcriptID int) ([]transcripts.TranscriptChunk, error) {
	if m.GetChunksByTranscriptIDFunc != nil {
		return m.GetChunksByTranscriptIDFunc(transcriptID)
	}
	return nil, fmt.Errorf("not implemented")
}

// Test Helpers
func makeLLMResponse(content string) llm.ChatCompletionResponse {
	return llm.ChatCompletionResponse{
		Choices: []llm.Choice{{Message: llm.Message{Content: content}}},
	}
}

// Mock Quiz Repository
func NewMockQuizRepository() *QuizRepository {
	var questions []Question
	var options []QuestionOption
	var sessions []UserQuizSession
	var answers []UserAnswer

	repo := &QuizRepository{
		logger: logger.NewRepositoryLogger("QuizRepository"),
		questionRepo: utils.NewRepository[Question](utils.WithQueryExecutor[Question](utils.QueryExecutor[Question]{
			QueryItem: func(query string, args ...any) (Question, error) {
				// GetQuestionByID
				if len(args) > 0 {
					qid, ok := args[0].(int)
					if ok {
						for _, q := range questions {
							if q.ID == qid {
								// Populate options for this question
								var questionOptions []QuestionOption
								for _, opt := range options {
									if opt.QuestionID == qid {
										questionOptions = append(questionOptions, opt)
									}
								}
								q.Options = questionOptions
								return q, nil
							}
						}
					}
				}
				return Question{}, fmt.Errorf("question not found")
			},
			QueryList: func(query string, args ...any) ([]Question, error) {
				// GetQuestionsByEpisodeID
				if len(args) > 0 {
					episodeID, ok := args[0].(int)
					if ok {
						var result []Question
						for _, q := range questions {
							if q.EpisodeID == episodeID {
								result = append(result, q)
							}
						}
						return result, nil
					}
				}
				return []Question{}, nil
			},
			Exec: func(query string, args ...any) error {
				// DeleteQuestionsByEpisodeID
				if len(args) > 0 {
					episodeID, ok := args[0].(int)
					if ok {
						// Remove questions with matching episodeID
						var filtered []Question
						for _, q := range questions {
							if q.EpisodeID != episodeID {
								filtered = append(filtered, q)
							}
						}
						questions = filtered
					}
				}
				return nil
			},
		})),
		optionRepo: utils.NewRepository[QuestionOption](utils.WithQueryExecutor[QuestionOption](utils.QueryExecutor[QuestionOption]{
			QueryItem: func(query string, args ...any) (QuestionOption, error) {
				if len(args) >= 4 {
					opt := QuestionOption{
						ID:         len(options) + 1,
						QuestionID: args[0].(int),
						OptionText: args[1].(string),
						Position:   args[2].(int),
						IsCorrect:  args[3].(bool),
					}
					options = append(options, opt)
					return opt, nil
				}
				return QuestionOption{}, nil
			},
		})),
		sessionRepo: utils.NewRepository[UserQuizSession](utils.WithQueryExecutor[UserQuizSession](utils.QueryExecutor[UserQuizSession]{
			QueryItem: func(query string, args ...any) (UserQuizSession, error) {
				// GetSessionByID or GetActiveSessionByUserAndEpisode
				if len(args) == 1 {
					// GetSessionByID
					sid := args[0].(int)
					for _, s := range sessions {
						if s.ID == sid {
							return s, nil
						}
					}
				} else if len(args) == 2 {
					// GetActiveSessionByUserAndEpisode
					userID := args[0].(int)
					episodeID := args[1].(int)
					for _, s := range sessions {
						if s.UserID == userID && s.EpisodeID == episodeID && s.Status == InProgress {
							return s, nil
						}
					}
					return UserQuizSession{}, sql.ErrNoRows
				}
				return UserQuizSession{}, sql.ErrNoRows
			},
			QueryList: func(query string, args ...any) ([]UserQuizSession, error) {
				// GetSessionsByUserID
				if len(args) > 0 {
					userID := args[0].(int)
					var result []UserQuizSession
					for _, s := range sessions {
						if s.UserID == userID {
							result = append(result, s)
						}
					}
					return result, nil
				}
				return []UserQuizSession{}, nil
			},
			Exec: func(query string, args ...any) error {
				// UpdateSession or DeleteSession
				if len(args) >= 5 {
					// UpdateSession
					// args: status, answered_questions, correct_answers, completed_at, session_id
					sessionID, ok := args[4].(int)
					if ok {
						for i := range sessions {
							if sessions[i].ID == sessionID {
								if status, ok := args[0].(SessionStatus); ok {
									sessions[i].Status = status
								}
								if answered, ok := args[1].(int); ok {
									sessions[i].AnsweredQuestions = answered
								}
								if correct, ok := args[2].(int); ok {
									sessions[i].CorrectAnswers = correct
								}
								// Handle CompletedAt - can be *time.Time or nil
								if completedAt, ok := args[3].(*time.Time); ok {
									sessions[i].CompletedAt = completedAt
								}
								break
							}
						}
					}
				} else if len(args) == 1 {
					// DeleteSession - simulate CASCADE by removing session and its answers
					sessionID, ok := args[0].(int)
					if ok {
						var filtered []UserQuizSession
						for _, s := range sessions {
							if s.ID != sessionID {
								filtered = append(filtered, s)
							}
						}
						sessions = filtered
						// CASCADE: Remove associated answers
						var filteredAnswers []UserAnswer
						for _, a := range answers {
							if a.SessionID != sessionID {
								filteredAnswers = append(filteredAnswers, a)
							}
						}
						answers = filteredAnswers
					}
				}
				return nil
			},
		})),
		answerRepo: utils.NewRepository[UserAnswer](utils.WithQueryExecutor[UserAnswer](utils.QueryExecutor[UserAnswer]{
			QueryItem: func(query string, args ...any) (UserAnswer, error) {
				// CreateAnswer
				if len(args) >= 6 {
					ans := UserAnswer{
						ID:         len(answers) + 1,
						SessionID:  args[0].(int),
						QuestionID: args[1].(int),
						UserID:     args[2].(int),
						IsCorrect:  args[5].(bool),
					}
					answers = append(answers, ans)
					return ans, nil
				}
				// GetAnswerBySessionAndQuestion
				if len(args) == 2 {
					sessionID := args[0].(int)
					questionID := args[1].(int)
					for _, a := range answers {
						if a.SessionID == sessionID && a.QuestionID == questionID {
							return a, nil
						}
					}
					return UserAnswer{}, fmt.Errorf("no rows in result set")
				}
				return UserAnswer{}, nil
			},
			QueryList: func(query string, args ...any) ([]UserAnswer, error) {
				// GetAnswersBySessionID
				if len(args) > 0 {
					sessionID := args[0].(int)
					var result []UserAnswer
					for _, a := range answers {
						if a.SessionID == sessionID {
							result = append(result, a)
						}
					}
					return result, nil
				}
				return []UserAnswer{}, nil
			},
			Exec: func(query string, args ...any) error {
				// DeleteAnswersBySessionID
				if len(args) == 1 {
					sessionID, ok := args[0].(int)
					if ok {
						var filtered []UserAnswer
						for _, a := range answers {
							if a.SessionID != sessionID {
								filtered = append(filtered, a)
							}
						}
						answers = filtered
					}
				}
				return nil
			},
		})),
	}

	// Add mock CreateQuestion
	originalQueryItem := repo.questionRepo.Executor.QueryItem
	repo.questionRepo.Executor.QueryItem = func(query string, args ...any) (Question, error) {
		if len(args) >= 4 {
			// Type assertion for QuestionType
			questionType, _ := args[2].(QuestionType)

			q := Question{
				ID:           len(questions) + 1,
				EpisodeID:    args[0].(int),
				QuestionText: args[1].(string),
				Type:         questionType,
				Position:     args[3].(int),
			}
			questions = append(questions, q)
			return q, nil
		}
		return originalQueryItem(query, args...)
	}

	// Add mock CreateSession
	originalSessionQueryItem := repo.sessionRepo.Executor.QueryItem
	repo.sessionRepo.Executor.QueryItem = func(query string, args ...any) (UserQuizSession, error) {
		if len(args) >= 6 {
			// Type assertion for SessionStatus
			status, _ := args[2].(SessionStatus)

			s := UserQuizSession{
				ID:                len(sessions) + 1,
				UserID:            args[0].(int),
				EpisodeID:         args[1].(int),
				Status:            status,
				TotalQuestions:    args[3].(int),
				AnsweredQuestions: args[4].(int),
				CorrectAnswers:    args[5].(int),
			}
			sessions = append(sessions, s)
			return s, nil
		}
		return originalSessionQueryItem(query, args...)
	}

	return repo
}

func NewMockQuizServiceReady() *QuizService {
	mockLLMClient := &MockLLMClient{
		ChatResponse: makeLLMResponse(`{"is_correct": true, "feedback": "Great job! This demonstrates a solid understanding of the concept."}`),
	}
	mockTranscriptRepo := &MockTranscriptRepository{}

	return &QuizService{
		repo:           *NewMockQuizRepository(),
		llmClient:      mockLLMClient,
		transcriptRepo: mockTranscriptRepo,
		logger:         logger.NewServiceLogger("QuizService"),
	}
}

func NewMockQuizHandlerReady() *QuizHandler {
	service := NewMockQuizServiceReady()
	return NewQuizHandler(service)
}
