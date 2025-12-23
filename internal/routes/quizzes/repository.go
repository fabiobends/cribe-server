package quizzes

import (
	"cribeapp.com/cribe-server/internal/core/logger"
	"cribeapp.com/cribe-server/internal/utils"
)

type QuizRepository struct {
	questionRepo *utils.Repository[Question]
	optionRepo   *utils.Repository[QuestionOption]
	sessionRepo  *utils.Repository[UserQuizSession]
	answerRepo   *utils.Repository[UserAnswer]
	logger       *logger.ContextualLogger
}

func NewQuizRepository(options ...utils.Option[Question]) *QuizRepository {
	return &QuizRepository{
		questionRepo: utils.NewRepository(options...),
		optionRepo:   utils.NewRepository[QuestionOption](),
		sessionRepo:  utils.NewRepository[UserQuizSession](),
		answerRepo:   utils.NewRepository[UserAnswer](),
		logger:       logger.NewRepositoryLogger("QuizRepository"),
	}
}

// Question operations

func (r *QuizRepository) CreateQuestion(question Question) (Question, error) {
	r.logger.Debug("Creating question in database", map[string]any{
		"episode_id": question.EpisodeID,
		"type":       question.Type,
	})

	query := `
		INSERT INTO questions (episode_id, question_text, type, position)
		VALUES ($1, $2, $3, $4)
		RETURNING id, episode_id, question_text, type, position, created_at, updated_at
	`

	result, err := r.questionRepo.Executor.QueryItem(query,
		question.EpisodeID,
		question.QuestionText,
		question.Type,
		question.Position,
	)
	if err != nil {
		r.logger.Error("Failed to create question", map[string]any{
			"error": err.Error(),
		})
		return Question{}, err
	}

	return result, nil
}

func (r *QuizRepository) CreateQuestionOption(option QuestionOption) (QuestionOption, error) {
	r.logger.Debug("Creating question option in database", map[string]any{
		"question_id": option.QuestionID,
	})

	query := `
		INSERT INTO question_options (question_id, option_text, position, is_correct)
		VALUES ($1, $2, $3, $4)
		RETURNING id, question_id, option_text, position, is_correct, created_at
	`

	result, err := r.optionRepo.Executor.QueryItem(query,
		option.QuestionID,
		option.OptionText,
		option.Position,
		option.IsCorrect,
	)
	if err != nil {
		r.logger.Error("Failed to create question option", map[string]any{
			"error": err.Error(),
		})
		return QuestionOption{}, err
	}

	return result, nil
}

func (r *QuizRepository) GetQuestionsByEpisodeID(episodeID int) ([]Question, error) {
	r.logger.Debug("Fetching questions by episode ID", map[string]any{
		"episode_id": episodeID,
	})

	query := `
		SELECT
			q.id, q.episode_id, q.question_text, q.type, q.position, q.created_at, q.updated_at,
			COALESCE(json_agg(
				json_build_object(
					'id', qo.id,
					'question_id', qo.question_id,
					'option_text', qo.option_text,
					'position', qo.position,
					'is_correct', qo.is_correct,
					'created_at', qo.created_at
				) ORDER BY qo.position
			) FILTER (WHERE qo.id IS NOT NULL), '[]') as options
		FROM questions q
		LEFT JOIN question_options qo ON q.id = qo.question_id
		WHERE q.episode_id = $1
		GROUP BY q.id
		ORDER BY q.position
	`

	result, err := r.questionRepo.Executor.QueryList(query, episodeID)
	if err != nil {
		r.logger.Error("Failed to fetch questions", map[string]any{
			"episode_id": episodeID,
			"error":      err.Error(),
		})
		return nil, err
	}

	return result, nil
}

func (r *QuizRepository) GetQuestionByID(questionID int) (Question, error) {
	r.logger.Debug("Fetching question by ID", map[string]any{
		"question_id": questionID,
	})

	query := `
		SELECT
			q.id, q.episode_id, q.question_text, q.type, q.position, q.created_at, q.updated_at,
			COALESCE(json_agg(
				json_build_object(
					'id', qo.id,
					'question_id', qo.question_id,
					'option_text', qo.option_text,
					'position', qo.position,
					'is_correct', qo.is_correct,
					'created_at', qo.created_at
				) ORDER BY qo.position
			) FILTER (WHERE qo.id IS NOT NULL), '[]') as options
		FROM questions q
		LEFT JOIN question_options qo ON q.id = qo.question_id
		WHERE q.id = $1
		GROUP BY q.id
	`

	result, err := r.questionRepo.Executor.QueryItem(query, questionID)
	if err != nil {
		r.logger.Error("Failed to fetch question", map[string]any{
			"question_id": questionID,
			"error":       err.Error(),
		})
		return Question{}, err
	}

	return result, nil
}

func (r *QuizRepository) DeleteQuestionsByEpisodeID(episodeID int) error {
	r.logger.Debug("Deleting questions for episode", map[string]any{
		"episode_id": episodeID,
	})

	query := `DELETE FROM questions WHERE episode_id = $1`
	err := r.questionRepo.Executor.Exec(query, episodeID)
	if err != nil {
		r.logger.Error("Failed to delete questions", map[string]any{
			"episode_id": episodeID,
			"error":      err.Error(),
		})
		return err
	}

	return nil
}

// Session operations

func (r *QuizRepository) CreateSession(session UserQuizSession) (UserQuizSession, error) {
	r.logger.Debug("Creating quiz session", map[string]any{
		"user_id":    session.UserID,
		"episode_id": session.EpisodeID,
	})

	query := `
		INSERT INTO user_quiz_sessions (user_id, episode_id, status, total_questions, answered_questions, correct_answers)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, user_id, episode_id, status, total_questions, answered_questions, correct_answers, started_at, completed_at, updated_at
	`

	result, err := r.sessionRepo.Executor.QueryItem(query,
		session.UserID,
		session.EpisodeID,
		session.Status,
		session.TotalQuestions,
		session.AnsweredQuestions,
		session.CorrectAnswers,
	)
	if err != nil {
		r.logger.Error("Failed to create session", map[string]any{
			"error": err.Error(),
		})
		return UserQuizSession{}, err
	}

	return result, nil
}

func (r *QuizRepository) GetSessionByID(sessionID int) (UserQuizSession, error) {
	r.logger.Debug("Fetching session by ID", map[string]any{
		"session_id": sessionID,
	})

	query := `
		SELECT id, user_id, episode_id, status, total_questions, answered_questions, correct_answers, started_at, completed_at, updated_at
		FROM user_quiz_sessions
		WHERE id = $1
	`

	result, err := r.sessionRepo.Executor.QueryItem(query, sessionID)
	if err != nil {
		r.logger.Error("Failed to fetch session", map[string]any{
			"session_id": sessionID,
			"error":      err.Error(),
		})
		return UserQuizSession{}, err
	}

	return result, nil
}

func (r *QuizRepository) GetActiveSessionByUserAndEpisode(userID, episodeID int) (UserQuizSession, error) {
	r.logger.Debug("Fetching session for user and episode", map[string]any{
		"user_id":    userID,
		"episode_id": episodeID,
	})

	// Get the most recent session (completed or in_progress)
	query := `
		SELECT id, user_id, episode_id, status, total_questions, answered_questions, correct_answers, started_at, completed_at, updated_at
		FROM user_quiz_sessions
		WHERE user_id = $1 AND episode_id = $2
		ORDER BY started_at DESC
		LIMIT 1
	`

	result, err := r.sessionRepo.Executor.QueryItem(query, userID, episodeID)
	if err != nil {
		r.logger.Error("Failed to fetch active session", map[string]any{
			"user_id":    userID,
			"episode_id": episodeID,
			"error":      err.Error(),
		})
		return UserQuizSession{}, err
	}

	return result, nil
}

func (r *QuizRepository) GetSessionsByUserID(userID int) ([]UserQuizSession, error) {
	r.logger.Debug("Fetching all sessions for user", map[string]any{
		"user_id": userID,
	})

	query := `
		SELECT id, user_id, episode_id, status, total_questions, answered_questions, correct_answers, started_at, completed_at, updated_at
		FROM user_quiz_sessions
		WHERE user_id = $1
		ORDER BY updated_at DESC
	`

	result, err := r.sessionRepo.Executor.QueryList(query, userID)
	if err != nil {
		r.logger.Error("Failed to fetch user sessions", map[string]any{
			"user_id": userID,
			"error":   err.Error(),
		})
		return nil, err
	}

	return result, nil
}

func (r *QuizRepository) UpdateSession(session UserQuizSession) error {
	r.logger.Debug("Updating session", map[string]any{
		"session_id": session.ID,
	})

	query := `
		UPDATE user_quiz_sessions
		SET status = $1, answered_questions = $2, correct_answers = $3, completed_at = $4, updated_at = NOW()
		WHERE id = $5
	`

	err := r.sessionRepo.Executor.Exec(query,
		session.Status,
		session.AnsweredQuestions,
		session.CorrectAnswers,
		session.CompletedAt,
		session.ID,
	)

	if err != nil {
		r.logger.Error("Failed to update session", map[string]any{
			"session_id": session.ID,
			"error":      err.Error(),
		})
		return err
	}

	return nil
}

func (r *QuizRepository) DeleteSession(sessionID int) error {
	r.logger.Debug("Deleting session", map[string]any{
		"session_id": sessionID,
	})

	// Delete the session and related answers
	deleteSessionQuery := `DELETE FROM user_quiz_sessions WHERE id = $1`
	if err := r.sessionRepo.Executor.Exec(deleteSessionQuery, sessionID); err != nil {
		r.logger.Error("Failed to delete session", map[string]any{
			"session_id": sessionID,
			"error":      err.Error(),
		})
		return err
	}

	r.logger.Info("Session deleted successfully", map[string]any{
		"session_id": sessionID,
	})

	return nil
}

// Answer operations

func (r *QuizRepository) CreateAnswer(answer UserAnswer) (UserAnswer, error) {
	r.logger.Debug("Creating user answer", map[string]any{
		"session_id":  answer.SessionID,
		"question_id": answer.QuestionID,
	})

	query := `
		INSERT INTO user_answers (session_id, question_id, user_id, selected_option_id, text_answer, is_correct, feedback)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, session_id, question_id, user_id, selected_option_id, text_answer, is_correct, feedback, answered_at
	`

	result, err := r.answerRepo.Executor.QueryItem(query,
		answer.SessionID,
		answer.QuestionID,
		answer.UserID,
		answer.SelectedOptionID,
		answer.TextAnswer,
		answer.IsCorrect,
		answer.Feedback,
	)
	if err != nil {
		r.logger.Error("Failed to create answer", map[string]any{
			"error": err.Error(),
		})
		return UserAnswer{}, err
	}

	return result, nil
}

func (r *QuizRepository) GetAnswersBySessionID(sessionID int) ([]UserAnswer, error) {
	r.logger.Debug("Fetching answers by session ID", map[string]any{
		"session_id": sessionID,
	})

	query := `
		SELECT id, session_id, question_id, user_id, selected_option_id, text_answer, is_correct, feedback, answered_at
		FROM user_answers
		WHERE session_id = $1
		ORDER BY answered_at
	`

	result, err := r.answerRepo.Executor.QueryList(query, sessionID)
	if err != nil {
		r.logger.Error("Failed to fetch answers", map[string]any{
			"session_id": sessionID,
			"error":      err.Error(),
		})
		return nil, err
	}

	return result, nil
}

func (r *QuizRepository) GetAnswerBySessionAndQuestion(sessionID, questionID int) (UserAnswer, error) {
	r.logger.Debug("Fetching answer by session ID and question ID", map[string]any{
		"session_id":  sessionID,
		"question_id": questionID,
	})

	query := `
		SELECT id, session_id, question_id, user_id, selected_option_id, text_answer, is_correct, feedback, answered_at
		FROM user_answers
		WHERE session_id = $1 AND question_id = $2
	`

	result, err := r.answerRepo.Executor.QueryItem(query, sessionID, questionID)
	if err != nil {
		r.logger.Error("Failed to fetch answer", map[string]any{
			"session_id":  sessionID,
			"question_id": questionID,
			"error":       err.Error(),
		})
		return UserAnswer{}, err
	}

	return result, nil
}
