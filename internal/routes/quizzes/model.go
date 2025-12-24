package quizzes

import (
	"time"

	"cribeapp.com/cribe-server/internal/errors"
	"cribeapp.com/cribe-server/internal/utils"
)

// QuestionType represents the type of question
type QuestionType string

const (
	MultipleChoice QuestionType = "multiple_choice"
	TrueFalse      QuestionType = "true_false"
	OpenEnded      QuestionType = "open_ended"
)

// SessionStatus represents the status of a quiz session
type SessionStatus string

const (
	InProgress SessionStatus = "in_progress"
	Completed  SessionStatus = "completed"
	Abandoned  SessionStatus = "abandoned"
)

// Question represents a quiz question for an episode
type Question struct {
	ID           int              `json:"id"`
	EpisodeID    int              `json:"episode_id"`
	QuestionText string           `json:"question_text"`
	Type         QuestionType     `json:"type"`
	Position     int              `json:"position"`
	CreatedAt    time.Time        `json:"created_at"`
	UpdatedAt    time.Time        `json:"updated_at"`
	Options      []QuestionOption `json:"options,omitempty"`
}

// QuestionOption represents an option for a multiple choice or true/false question
type QuestionOption struct {
	ID         int       `json:"id"`
	QuestionID int       `json:"question_id"`
	OptionText string    `json:"option_text"`
	Position   int       `json:"position"`
	IsCorrect  bool      `json:"is_correct"`
	CreatedAt  time.Time `json:"created_at"`
}

// UserQuizSession represents a user's quiz session for an episode
type UserQuizSession struct {
	ID                int           `json:"id"`
	UserID            int           `json:"user_id"`
	EpisodeID         int           `json:"episode_id"`
	EpisodeName       string        `json:"episode_name"`
	PodcastName       string        `json:"podcast_name"`
	Status            SessionStatus `json:"status"`
	TotalQuestions    int           `json:"total_questions"`
	AnsweredQuestions int           `json:"answered_questions"`
	CorrectAnswers    int           `json:"correct_answers"`
	StartedAt         time.Time     `json:"started_at"`
	CompletedAt       *time.Time    `json:"completed_at,omitempty"`
	UpdatedAt         time.Time     `json:"updated_at"`
}

// UserAnswer represents a user's answer to a question
type UserAnswer struct {
	ID               int       `json:"id"`
	SessionID        int       `json:"session_id"`
	QuestionID       int       `json:"question_id"`
	UserID           int       `json:"user_id"`
	SelectedOptionID *int      `json:"selected_option_id,omitempty"`
	TextAnswer       *string   `json:"text_answer,omitempty"`
	IsCorrect        bool      `json:"is_correct"`
	Feedback         string    `json:"feedback"`
	AnsweredAt       time.Time `json:"answered_at"`
}

// DTOs for API requests/responses

// GenerateQuestionsRequest is the request to generate questions for an episode
type GenerateQuestionsRequest struct {
	EpisodeID int `json:"episode_id" validate:"required,min=1"`
}

func (dto GenerateQuestionsRequest) Validate() *errors.ErrorResponse {
	return utils.ValidateStruct(dto)
}

// StartSessionRequest is the request to start a quiz session
type StartSessionRequest struct {
	EpisodeID int `json:"episode_id" validate:"required,min=1"`
}

func (dto StartSessionRequest) Validate() *errors.ErrorResponse {
	return utils.ValidateStruct(dto)
}

// GetOrCreateSessionRequest is the request to get or create a quiz session
type GetOrCreateSessionRequest struct {
	EpisodeID int `json:"episode_id" validate:"required,min=1"`
}

func (dto GetOrCreateSessionRequest) Validate() *errors.ErrorResponse {
	return utils.ValidateStruct(dto)
}

// QuizSessionDetail is the complete session response with questions and answers
type QuizSessionDetail struct {
	Session   UserQuizSession `json:"session"`
	Questions []Question      `json:"questions"`
	Answers   []UserAnswer    `json:"answers"`
}

// QuizSessionSummary is a simplified session for list views
type QuizSessionSummary struct {
	ID                int           `json:"id"`
	EpisodeID         int           `json:"episode_id"`
	Status            SessionStatus `json:"status"`
	TotalQuestions    int           `json:"total_questions"`
	AnsweredQuestions int           `json:"answered_questions"`
	CorrectAnswers    int           `json:"correct_answers"`
	StartedAt         time.Time     `json:"started_at"`
	CompletedAt       *time.Time    `json:"completed_at,omitempty"`
	UpdatedAt         time.Time     `json:"updated_at"`
}

// SubmitAnswerRequest is the request to submit an answer
type SubmitAnswerRequest struct {
	QuestionID       int     `json:"question_id" validate:"required,min=1"`
	SelectedOptionID *int    `json:"selected_option_id,omitempty"`
	TextAnswer       *string `json:"text_answer,omitempty"`
}

func (dto SubmitAnswerRequest) Validate() *errors.ErrorResponse {
	// Custom validation: must have either selected_option_id or text_answer, but not both
	if (dto.SelectedOptionID == nil && dto.TextAnswer == nil) ||
		(dto.SelectedOptionID != nil && dto.TextAnswer != nil) {
		return &errors.ErrorResponse{
			Message: errors.ValidationError,
			Details: "Must provide either selected_option_id or text_answer, but not both",
		}
	}

	return utils.ValidateStruct(dto)
}

// UpdateSessionStatusRequest is the request to update session status
type UpdateSessionStatusRequest struct {
	Status SessionStatus `json:"status" validate:"required,oneof=completed abandoned"`
}

func (dto UpdateSessionStatusRequest) Validate() *errors.ErrorResponse {
	return utils.ValidateStruct(dto)
}

// LLM Response structures for question generation

// LLMQuestionOption represents an option generated by LLM
type LLMQuestionOption struct {
	Text      string `json:"text"`
	IsCorrect bool   `json:"is_correct"`
}

// LLMQuestion represents a question generated by LLM
type LLMQuestion struct {
	QuestionText string              `json:"question_text"`
	Type         string              `json:"type"`
	Options      []LLMQuestionOption `json:"options,omitempty"`
}

// LLMQuestionsResponse represents the full response from LLM
type LLMQuestionsResponse struct {
	Questions []LLMQuestion `json:"questions"`
}

type LLMEvaluationResponse struct {
	IsCorrect bool   `json:"is_correct"`
	Feedback  string `json:"feedback"`
}
