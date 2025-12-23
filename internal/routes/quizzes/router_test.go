package quizzes

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"cribeapp.com/cribe-server/internal/core/logger"
	"cribeapp.com/cribe-server/internal/errors"
	"cribeapp.com/cribe-server/internal/middlewares"
	"cribeapp.com/cribe-server/internal/routes/migrations"
	"cribeapp.com/cribe-server/internal/utils"
)

var log = logger.NewCoreLogger("QuizzesRouterTest")

// handlerWithAuth injects userID into context for authenticated routes
func handlerWithAuth(w http.ResponseWriter, r *http.Request) {
	ctx := context.WithValue(r.Context(), middlewares.UserIDContextKey, utils.TestUserID)
	HandleHTTPRequests(w, r.WithContext(ctx))
}

func TestQuizzes_IntegrationTests(t *testing.T) {
	log.Info("Setting up test environment", nil)

	// Clean database
	if err := utils.CleanDatabase(); err != nil {
		t.Fatalf("Failed to clean database: %v", err)
	}

	// Run migrations
	resp := utils.MustSendTestRequest[[]migrations.Migration](utils.TestRequest{
		Method:      http.MethodPost,
		URL:         "/migrations",
		HandlerFunc: migrations.HandleHTTPRequests,
	})
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("Failed to run migrations: status %d", resp.StatusCode)
	}

	// Setup test data
	_ = utils.CreateTestUser(utils.TestUserID)
	_, _ = utils.CreateTestPodcast()
	_, _ = utils.CreateTestEpisode(utils.TestPodcastID)
	_, _ = utils.CreateTestTranscript(utils.TestUserID)
	_ = utils.CreateTestTranscriptChunks(utils.TestTranscriptID)

	log.Info("Test data created", map[string]any{
		"userID":       utils.TestUserID,
		"podcastID":    utils.TestPodcastID,
		"episodeID":    utils.TestEpisodeID,
		"transcriptID": utils.TestTranscriptID,
	})

	// Create quiz questions
	if err := utils.CreateTestQuestions(utils.TestEpisodeID); err != nil {
		t.Fatalf("Failed to create test questions: %v", err)
	}

	log.Info("Test data created successfully", map[string]any{
		"userID":       utils.TestUserID,
		"podcastID":    utils.TestPodcastID,
		"episodeID":    utils.TestEpisodeID,
		"transcriptID": utils.TestTranscriptID,
	})

	t.Run("should return not found for invalid route", func(t *testing.T) {
		resp := utils.MustSendTestRequest[errors.ErrorResponse](utils.TestRequest{
			Method:      http.MethodGet,
			URL:         "/quizzes/invalid",
			HandlerFunc: handlerWithAuth,
		})

		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("Expected status code %d, got %d", http.StatusNotFound, resp.StatusCode)
		}
	})

	t.Run("should start a quiz session with authentication", func(t *testing.T) {
		req := StartSessionRequest{
			EpisodeID: utils.TestEpisodeID,
		}

		resp := utils.MustSendTestRequest[QuizSessionDetail](utils.TestRequest{
			Method:      http.MethodPost,
			URL:         "/quizzes",
			Body:        req,
			HandlerFunc: handlerWithAuth,
		})

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
		}

		if resp.Body.Session.ID == 0 {
			t.Error("Expected session ID to be non-zero")
		}

		if resp.Body.Session.EpisodeID != utils.TestEpisodeID {
			t.Errorf("Expected episode ID %d, got %d", utils.TestEpisodeID, resp.Body.Session.EpisodeID)
		}

		if resp.Body.Session.Status != InProgress {
			t.Errorf("Expected status 'in_progress', got %s", resp.Body.Session.Status)
		}

		log.Info("Quiz session started successfully", map[string]any{
			"sessionID": resp.Body.Session.ID,
			"status":    resp.Body.Session.Status,
		})
	})

	t.Run("should return bad request when starting session with invalid episode", func(t *testing.T) {
		req := StartSessionRequest{
			EpisodeID: 0, // Invalid
		}

		resp := utils.MustSendTestRequest[errors.ErrorResponse](utils.TestRequest{
			Method:      http.MethodPost,
			URL:         "/quizzes",
			Body:        req,
			HandlerFunc: handlerWithAuth,
		})

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, resp.StatusCode)
		}
	})

	// Test answers endpoints
	var sessionID int

	t.Run("should create session for answer tests", func(t *testing.T) {
		req := StartSessionRequest{
			EpisodeID: utils.TestEpisodeID,
		}

		resp := utils.MustSendTestRequest[QuizSessionDetail](utils.TestRequest{
			Method:      http.MethodPost,
			URL:         "/quizzes",
			Body:        req,
			HandlerFunc: handlerWithAuth,
		})

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Failed to create session: status %d", resp.StatusCode)
		}

		sessionID = resp.Body.Session.ID
	})

	t.Run("should submit an answer with selected option", func(t *testing.T) {
		// Get questions using repository directly
		repo := NewQuizRepository()
		questions, err := repo.GetQuestionsByEpisodeID(utils.TestEpisodeID)
		if err != nil || len(questions) == 0 || len(questions[0].Options) == 0 {
			t.Fatal("No questions or options available")
		}

		question := questions[0]
		optionID := question.Options[0].ID

		req := SubmitAnswerRequest{
			QuestionID:       question.ID,
			SelectedOptionID: &optionID,
		}

		resp := utils.MustSendTestRequest[UserAnswer](utils.TestRequest{
			Method:      http.MethodPost,
			URL:         fmt.Sprintf("/quizzes/%d/answers", sessionID),
			Body:        req,
			HandlerFunc: handlerWithAuth,
		})

		if resp.StatusCode != http.StatusCreated {
			t.Errorf("Expected status code %d, got %d", http.StatusCreated, resp.StatusCode)
		}

		if resp.Body.QuestionID != question.ID {
			t.Errorf("Expected question ID %d, got %d", question.ID, resp.Body.QuestionID)
		}

		if resp.Body.SelectedOptionID == nil || *resp.Body.SelectedOptionID != optionID {
			t.Errorf("Expected option ID %d, got %v", optionID, resp.Body.SelectedOptionID)
		}

		log.Info("Answer submitted successfully", map[string]any{
			"answerID":   resp.Body.ID,
			"questionID": resp.Body.QuestionID,
			"optionID":   *resp.Body.SelectedOptionID,
		})
	})

	t.Run("should return bad request when submitting answer without option or text", func(t *testing.T) {
		req := SubmitAnswerRequest{
			QuestionID: 1,
			// No SelectedOptionID or TextAnswer
		}

		resp := utils.MustSendTestRequest[errors.ErrorResponse](utils.TestRequest{
			Method:      http.MethodPost,
			URL:         fmt.Sprintf("/quizzes/%d/answers", sessionID),
			Body:        req,
			HandlerFunc: handlerWithAuth,
		})

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, resp.StatusCode)
		}
	})

	t.Run("should return bad request when submitting answer with both option and text", func(t *testing.T) {
		optionID := 1
		textAnswer := "Some text"

		req := SubmitAnswerRequest{
			QuestionID:       1,
			SelectedOptionID: &optionID,
			TextAnswer:       &textAnswer,
		}

		resp := utils.MustSendTestRequest[errors.ErrorResponse](utils.TestRequest{
			Method:      http.MethodPost,
			URL:         fmt.Sprintf("/quizzes/%d/answers", sessionID),
			Body:        req,
			HandlerFunc: handlerWithAuth,
		})

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, resp.StatusCode)
		}
	})
}
