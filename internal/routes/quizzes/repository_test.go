package quizzes

import (
	"fmt"
	"testing"

	"cribeapp.com/cribe-server/internal/core/logger"
	"cribeapp.com/cribe-server/internal/utils"
)

var repo = NewMockQuizRepository()

func TestQuizRepository_CreateQuestion(t *testing.T) {
	t.Run("should create a question with valid input", func(t *testing.T) {
		question := Question{
			EpisodeID:    1,
			QuestionText: "What is the main topic?",
			Type:         MultipleChoice,
			Position:     0,
		}

		result, err := repo.CreateQuestion(question)

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if result.ID == 0 {
			t.Errorf("Expected question ID, got 0")
		}

		if result.QuestionText != question.QuestionText {
			t.Errorf("Expected question text to be %v, got %v", question.QuestionText, result.QuestionText)
		}

		if result.Type != question.Type {
			t.Errorf("Expected question type to be %v, got %v", question.Type, result.Type)
		}
	})

	t.Run("should return error when creation fails", func(t *testing.T) {
		// Create a repository with a failing executor
		failingRepo := &QuizRepository{
			questionRepo: utils.NewRepository[Question](utils.WithQueryExecutor[Question](utils.QueryExecutor[Question]{
				QueryItem: func(query string, args ...any) (Question, error) {
					return Question{}, fmt.Errorf("database connection failed")
				},
			})),
			logger: logger.NewRepositoryLogger("QuizRepository"),
		}

		question := Question{
			EpisodeID:    1,
			QuestionText: "What is the main topic?",
			Type:         MultipleChoice,
			Position:     0,
		}

		_, err := failingRepo.CreateQuestion(question)

		if err == nil {
			t.Errorf("Expected error, got nil")
		}

		if err.Error() != "database connection failed" {
			t.Errorf("Expected 'database connection failed' error, got %v", err.Error())
		}
	})
}

func TestQuizRepository_CreateQuestionOption(t *testing.T) {
	t.Run("should create a question option", func(t *testing.T) {
		option := QuestionOption{
			QuestionID: 1,
			OptionText: "Option A",
			Position:   0,
			IsCorrect:  true,
		}

		result, err := repo.CreateQuestionOption(option)

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if result.ID == 0 {
			t.Errorf("Expected option ID, got 0")
		}

		if result.OptionText != option.OptionText {
			t.Errorf("Expected option text to be %v, got %v", option.OptionText, result.OptionText)
		}

		if result.IsCorrect != option.IsCorrect {
			t.Errorf("Expected is_correct to be %v, got %v", option.IsCorrect, result.IsCorrect)
		}
	})
}

func TestQuizRepository_GetQuestionsByEpisodeID(t *testing.T) {
	t.Run("should get questions by episode ID", func(t *testing.T) {
		questions, err := repo.GetQuestionsByEpisodeID(1)

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if len(questions) == 0 {
			t.Errorf("Expected at least 1 question, got %d", len(questions))
		}

		if questions[0].EpisodeID != 1 {
			t.Errorf("Expected episode ID to be 1, got %v", questions[0].EpisodeID)
		}
	})

	t.Run("should return empty list for non-existent episode", func(t *testing.T) {
		questions, err := repo.GetQuestionsByEpisodeID(999)

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if len(questions) != 0 {
			t.Errorf("Expected 0 questions, got %d", len(questions))
		}
	})
}

func TestQuizRepository_GetQuestionByID(t *testing.T) {
	t.Run("should get a question by ID", func(t *testing.T) {
		question, err := repo.GetQuestionByID(1)

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if question.ID != 1 {
			t.Errorf("Expected question ID to be 1, got %v", question.ID)
		}

		if len(question.Options) == 0 {
			t.Errorf("Expected question to have options")
		}
	})

	t.Run("should return error for non-existent question", func(t *testing.T) {
		_, err := repo.GetQuestionByID(999)

		if err == nil {
			t.Errorf("Expected error, got nil")
		}
	})
}

func TestQuizRepository_DeleteQuestionsByEpisodeID(t *testing.T) {
	t.Run("should delete questions for an episode", func(t *testing.T) {
		repo := NewMockQuizRepository()

		// Create a question first
		question := Question{
			EpisodeID:    1,
			QuestionText: "Test question",
			Type:         MultipleChoice,
			Position:     0,
		}
		_, _ = repo.CreateQuestion(question)

		// Delete questions
		err := repo.DeleteQuestionsByEpisodeID(1)

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		// Verify deletion
		questions, _ := repo.GetQuestionsByEpisodeID(1)
		if len(questions) != 0 {
			t.Errorf("Expected 0 questions after deletion, got %d", len(questions))
		}
	})
}

func TestQuizRepository_CreateSession(t *testing.T) {
	t.Run("should create a quiz session", func(t *testing.T) {
		session := UserQuizSession{
			UserID:            1,
			EpisodeID:         1,
			Status:            InProgress,
			TotalQuestions:    5,
			AnsweredQuestions: 0,
			CorrectAnswers:    0,
		}

		result, err := repo.CreateSession(session)

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if result.ID == 0 {
			t.Errorf("Expected session ID, got 0")
		}

		if result.UserID != session.UserID {
			t.Errorf("Expected user ID to be %v, got %v", session.UserID, result.UserID)
		}

		if result.Status != session.Status {
			t.Errorf("Expected status to be %v, got %v", session.Status, result.Status)
		}
	})
}

func TestQuizRepository_GetSessionByID(t *testing.T) {
	t.Run("should get a session by ID", func(t *testing.T) {
		session, err := repo.GetSessionByID(1)

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if session.ID != 1 {
			t.Errorf("Expected session ID to be 1, got %v", session.ID)
		}
	})

	t.Run("should return error for non-existent session", func(t *testing.T) {
		_, err := repo.GetSessionByID(999)

		if err == nil {
			t.Errorf("Expected error, got nil")
		}
	})
}

func TestQuizRepository_GetActiveSessionByUserAndEpisode(t *testing.T) {
	t.Run("should get active session for user and episode", func(t *testing.T) {
		session, err := repo.GetActiveSessionByUserAndEpisode(1, 1)

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if session.UserID != 1 {
			t.Errorf("Expected user ID to be 1, got %v", session.UserID)
		}

		if session.Status != InProgress {
			t.Errorf("Expected status to be in_progress, got %v", session.Status)
		}
	})

	t.Run("should return error when no active session exists", func(t *testing.T) {
		_, err := repo.GetActiveSessionByUserAndEpisode(999, 999)

		if err == nil {
			t.Errorf("Expected error, got nil")
		}
	})
}

func TestQuizRepository_UpdateSession(t *testing.T) {
	t.Run("should update a session", func(t *testing.T) {
		session, _ := repo.GetSessionByID(1)
		session.AnsweredQuestions = 3
		session.CorrectAnswers = 2
		session.Status = Completed

		err := repo.UpdateSession(session)

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		// Verify update
		updated, _ := repo.GetSessionByID(1)
		if updated.AnsweredQuestions != 3 {
			t.Errorf("Expected answered_questions to be 3, got %v", updated.AnsweredQuestions)
		}
	})
}

func TestQuizRepository_CreateAnswer(t *testing.T) {
	t.Run("should create an answer", func(t *testing.T) {
		optionID := 1
		answer := UserAnswer{
			SessionID:        1,
			QuestionID:       1,
			UserID:           1,
			SelectedOptionID: &optionID,
			IsCorrect:        true,
			Feedback:         "Correct!",
		}

		result, err := repo.CreateAnswer(answer)

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if result.ID == 0 {
			t.Errorf("Expected answer ID, got 0")
		}

		if result.SessionID != answer.SessionID {
			t.Errorf("Expected session ID to be %v, got %v", answer.SessionID, result.SessionID)
		}
	})
}

func TestQuizRepository_GetAnswersBySessionID(t *testing.T) {
	t.Run("should get answers by session ID", func(t *testing.T) {
		answers, err := repo.GetAnswersBySessionID(1)

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if len(answers) == 0 {
			t.Errorf("Expected at least 1 answer, got %d", len(answers))
		}
	})

	t.Run("should return empty list for session with no answers", func(t *testing.T) {
		answers, err := repo.GetAnswersBySessionID(999)

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if len(answers) != 0 {
			t.Errorf("Expected 0 answers, got %d", len(answers))
		}
	})
}

func TestQuizRepository_GetAnswerBySessionAndQuestion(t *testing.T) {
	t.Run("should get answer by session and question", func(t *testing.T) {
		answer, err := repo.GetAnswerBySessionAndQuestion(1, 1)

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if answer.SessionID != 1 {
			t.Errorf("Expected session ID to be 1, got %v", answer.SessionID)
		}

		if answer.QuestionID != 1 {
			t.Errorf("Expected question ID to be 1, got %v", answer.QuestionID)
		}
	})

	t.Run("should return error for non-existent answer", func(t *testing.T) {
		_, err := repo.GetAnswerBySessionAndQuestion(999, 999)

		if err == nil {
			t.Errorf("Expected error, got nil")
		}
	})
}

func TestQuizRepository_GetSessionsByUserID(t *testing.T) {
	t.Run("should get all sessions for a user", func(t *testing.T) {
		// Create multiple sessions for the same user
		session1 := UserQuizSession{
			UserID:            1,
			EpisodeID:         1,
			Status:            InProgress,
			TotalQuestions:    5,
			AnsweredQuestions: 0,
			CorrectAnswers:    0,
		}
		session2 := UserQuizSession{
			UserID:            1,
			EpisodeID:         2,
			Status:            Completed,
			TotalQuestions:    5,
			AnsweredQuestions: 5,
			CorrectAnswers:    4,
		}

		_, _ = repo.CreateSession(session1)
		_, _ = repo.CreateSession(session2)

		sessions, err := repo.GetSessionsByUserID(1)

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if len(sessions) < 2 {
			t.Errorf("Expected at least 2 sessions, got %d", len(sessions))
		}

		// Verify all sessions belong to user 1
		for _, session := range sessions {
			if session.UserID != 1 {
				t.Errorf("Expected user ID to be 1, got %v", session.UserID)
			}
		}
	})

	t.Run("should return empty list for user with no sessions", func(t *testing.T) {
		sessions, err := repo.GetSessionsByUserID(999)

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if len(sessions) != 0 {
			t.Errorf("Expected 0 sessions, got %d", len(sessions))
		}
	})
}

func TestQuizRepository_DeleteSession(t *testing.T) {
	t.Run("should delete a session and its answers", func(t *testing.T) {
		// Create a session
		session := UserQuizSession{
			UserID:            1,
			EpisodeID:         1,
			Status:            InProgress,
			TotalQuestions:    5,
			AnsweredQuestions: 0,
			CorrectAnswers:    0,
		}
		createdSession, _ := repo.CreateSession(session)

		// Create an answer for the session
		optionID := 1
		answer := UserAnswer{
			SessionID:        createdSession.ID,
			QuestionID:       1,
			UserID:           1,
			SelectedOptionID: &optionID,
			IsCorrect:        true,
			Feedback:         "Correct!",
		}
		_, _ = repo.CreateAnswer(answer)

		// Delete the session
		err := repo.DeleteSession(createdSession.ID)

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		// Verify session is deleted
		_, err = repo.GetSessionByID(createdSession.ID)
		if err == nil {
			t.Errorf("Expected error when fetching deleted session, got nil")
		}

		// Verify CASCADE deletion removed associated answers
		answers, _ := repo.GetAnswersBySessionID(createdSession.ID)
		if len(answers) != 0 {
			t.Errorf("CASCADE deletion failed: expected 0 answers after session deletion, got %d", len(answers))
		}
	})

	t.Run("should not error when deleting non-existent session", func(t *testing.T) {
		err := repo.DeleteSession(999)

		// This should not error - deleting non-existent records is idempotent
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
	})
}
