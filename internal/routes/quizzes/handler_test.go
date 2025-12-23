package quizzes

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"cribeapp.com/cribe-server/internal/middlewares"
)

var handler = NewMockQuizHandlerReady()

func TestQuizHandler_StartSession(t *testing.T) {
	t.Run("should return error when starting session without questions", func(t *testing.T) {
		req := StartSessionRequest{EpisodeID: 999}
		body, _ := json.Marshal(req)

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/quizzes", bytes.NewBuffer(body))
		ctx := context.WithValue(r.Context(), middlewares.UserIDContextKey, 1)
		r = r.WithContext(ctx)

		handler.HandleRequest(w, r)

		if w.Code == http.StatusCreated {
			t.Error("Expected error status when no questions exist")
		}
	})
}

func TestQuizHandler_NotFound(t *testing.T) {
	t.Run("should return 404 for invalid routes", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/quizzes/invalid", nil)

		handler.HandleRequest(w, r)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
		}
	})
}

func TestQuizHandler_NotAllowed(t *testing.T) {
	t.Run("should return 405 for invalid method", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPatch, "/quizzes", nil)

		handler.HandleRequest(w, r)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("Expected status %d, got %d", http.StatusMethodNotAllowed, w.Code)
		}
	})
}

func TestQuizHandler_handleGetSessionsWithDetails(t *testing.T) {
	t.Run("should get all sessions with details for user", func(t *testing.T) {
		// Setup: Create test data
		repo := NewMockQuizRepository()

		// Create questions for episode
		_, _ = repo.CreateQuestion(Question{EpisodeID: 1, QuestionText: "Question 1", Type: MultipleChoice, Position: 1})
		_, _ = repo.CreateQuestion(Question{EpisodeID: 1, QuestionText: "Question 2", Type: MultipleChoice, Position: 2})

		// Create a session for user 1
		_, _ = repo.CreateSession(UserQuizSession{UserID: 1, EpisodeID: 1, Status: InProgress, TotalQuestions: 2, AnsweredQuestions: 0, CorrectAnswers: 0})

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/quizzes", nil)
		ctx := context.WithValue(r.Context(), middlewares.UserIDContextKey, 1)
		r = r.WithContext(ctx)

		service := NewQuizService(*repo, nil, nil)
		handler := NewQuizHandler(service)
		handler.HandleRequest(w, r)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}

		var sessions []QuizSessionDetail
		if err := json.NewDecoder(w.Body).Decode(&sessions); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if len(sessions) != 1 {
			t.Errorf("Expected 1 session, got %d", len(sessions))
		}

		if len(sessions) > 0 {
			if sessions[0].Session.ID != 1 {
				t.Errorf("Expected session ID 1, got %d", sessions[0].Session.ID)
			}
			if sessions[0].Session.UserID != 1 {
				t.Errorf("Expected user ID 1, got %d", sessions[0].Session.UserID)
			}
		}
	})

	t.Run("should return empty array when user has no sessions", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/quizzes", nil)
		ctx := context.WithValue(r.Context(), middlewares.UserIDContextKey, 999)
		r = r.WithContext(ctx)

		handler.HandleRequest(w, r)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}

		var sessions []QuizSessionDetail
		if err := json.NewDecoder(w.Body).Decode(&sessions); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if len(sessions) != 0 {
			t.Errorf("Expected 0 sessions, got %d", len(sessions))
		}
	})
}

func TestQuizHandler_handleSessionByID(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		setupData  bool
		sessionID  int
		userID     int
		wantStatus int
	}{
		{
			name:       "method not allowed",
			method:     http.MethodPost,
			setupData:  true,
			sessionID:  1,
			userID:     1,
			wantStatus: http.StatusMethodNotAllowed,
		},
		{
			name:       "successful get session",
			method:     http.MethodGet,
			setupData:  true,
			sessionID:  1,
			userID:     1,
			wantStatus: http.StatusOK,
		},
		{
			name:       "session not found",
			method:     http.MethodGet,
			setupData:  false,
			sessionID:  999,
			userID:     1,
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewMockQuizRepository()
			service := NewQuizService(*repo, nil, nil)
			testHandler := NewQuizHandler(service)

			if tt.setupData {
				_, _ = repo.CreateSession(UserQuizSession{
					ID:        1,
					UserID:    1,
					EpisodeID: 1,
					Status:    InProgress,
				})
			}

			w := httptest.NewRecorder()
			r := httptest.NewRequest(tt.method, "/quizzes/1", nil)
			ctx := context.WithValue(r.Context(), middlewares.UserIDContextKey, tt.userID)
			r = r.WithContext(ctx)

			testHandler.handleSessionWithDetailsByID(w, r, tt.sessionID)

			if w.Code != tt.wantStatus {
				t.Errorf("Expected status %d, got %d", tt.wantStatus, w.Code)
			}
		})
	}
}

func TestQuizHandler_handleSessionStatus(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		body       any
		setupData  bool
		sessionID  int
		userID     int
		wantStatus int
	}{
		{
			name:       "method not allowed",
			method:     http.MethodPost,
			setupData:  true,
			sessionID:  1,
			userID:     1,
			wantStatus: http.StatusMethodNotAllowed,
		},
		{
			name:       "invalid request body",
			method:     http.MethodPatch,
			body:       "invalid json",
			setupData:  true,
			sessionID:  1,
			userID:     1,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "validation error - invalid status",
			method:     http.MethodPatch,
			body:       UpdateSessionStatusRequest{Status: InProgress},
			setupData:  true,
			sessionID:  1,
			userID:     1,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "successful status update",
			method:     http.MethodPatch,
			body:       UpdateSessionStatusRequest{Status: Completed},
			setupData:  true,
			sessionID:  1,
			userID:     1,
			wantStatus: http.StatusOK,
		},
		{
			name:       "session not found - returns not found",
			method:     http.MethodPatch,
			body:       UpdateSessionStatusRequest{Status: Completed},
			setupData:  false,
			sessionID:  999,
			userID:     1,
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewMockQuizRepository()
			service := NewQuizService(*repo, nil, nil)
			testHandler := NewQuizHandler(service)

			if tt.setupData {
				_, _ = repo.CreateSession(UserQuizSession{
					ID:        1,
					UserID:    1,
					EpisodeID: 1,
					Status:    InProgress,
				})
			}

			var body []byte
			if tt.body != nil {
				if str, ok := tt.body.(string); ok {
					body = []byte(str)
				} else {
					body, _ = json.Marshal(tt.body)
				}
			}

			w := httptest.NewRecorder()
			r := httptest.NewRequest(tt.method, "/quizzes/1/status", bytes.NewBuffer(body))
			ctx := context.WithValue(r.Context(), middlewares.UserIDContextKey, tt.userID)
			r = r.WithContext(ctx)

			testHandler.handleSessionStatus(w, r, tt.sessionID, tt.userID)

			if w.Code != tt.wantStatus {
				t.Errorf("Expected status %d, got %d", tt.wantStatus, w.Code)
			}
		})
	}
}

func TestQuizHandler_Routing(t *testing.T) {
	t.Run("retrieves session details by ID", func(t *testing.T) {
		repo := NewMockQuizRepository()
		service := NewQuizService(*repo, nil, nil)
		testHandler := NewQuizHandler(service)

		// Create test session
		_, _ = repo.CreateSession(UserQuizSession{
			ID:        1,
			UserID:    1,
			EpisodeID: 1,
			Status:    InProgress,
		})

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/quizzes/1", nil)
		ctx := context.WithValue(r.Context(), middlewares.UserIDContextKey, 1)
		r = r.WithContext(ctx)

		testHandler.HandleRequest(w, r)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}

		var session QuizSessionDetail
		if err := json.NewDecoder(w.Body).Decode(&session); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if session.Session.ID != 1 {
			t.Errorf("Expected session ID 1, got %d", session.Session.ID)
		}
	})

	t.Run("submits answer to question", func(t *testing.T) {
		repo := NewMockQuizRepository()
		service := NewQuizService(*repo, nil, nil)
		testHandler := NewQuizHandler(service)

		// Create test data
		session, _ := repo.CreateSession(UserQuizSession{
			ID:        1,
			UserID:    1,
			EpisodeID: 1,
			Status:    InProgress,
		})
		question, _ := repo.CreateQuestion(Question{
			ID:           1,
			EpisodeID:    1,
			QuestionText: "Test question",
			Type:         MultipleChoice,
			Position:     0,
		})
		option, _ := repo.CreateQuestionOption(QuestionOption{
			ID:         1,
			QuestionID: question.ID,
			OptionText: "Correct answer",
			Position:   0,
			IsCorrect:  true,
		})

		reqBody := SubmitAnswerRequest{
			QuestionID:       question.ID,
			SelectedOptionID: &option.ID,
		}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/quizzes/1/answers", bytes.NewBuffer(body))
		ctx := context.WithValue(r.Context(), middlewares.UserIDContextKey, session.UserID)
		r = r.WithContext(ctx)

		testHandler.HandleRequest(w, r)

		if w.Code != http.StatusCreated {
			t.Errorf("Expected status %d, got %d", http.StatusCreated, w.Code)
		}
	})

	t.Run("updates session status when completing quiz", func(t *testing.T) {
		repo := NewMockQuizRepository()
		service := NewQuizService(*repo, nil, nil)
		testHandler := NewQuizHandler(service)

		// Create test session
		_, _ = repo.CreateSession(UserQuizSession{
			ID:        1,
			UserID:    1,
			EpisodeID: 1,
			Status:    InProgress,
		})

		reqBody := UpdateSessionStatusRequest{Status: Completed}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPatch, "/quizzes/1/status", bytes.NewBuffer(body))
		ctx := context.WithValue(r.Context(), middlewares.UserIDContextKey, 1)
		r = r.WithContext(ctx)

		testHandler.HandleRequest(w, r)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}
	})

	t.Run("returns not found for unknown sub-path", func(t *testing.T) {
		repo := NewMockQuizRepository()
		service := NewQuizService(*repo, nil, nil)
		testHandler := NewQuizHandler(service)

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/quizzes/1/unknown", nil)
		ctx := context.WithValue(r.Context(), middlewares.UserIDContextKey, 1)
		r = r.WithContext(ctx)

		testHandler.HandleRequest(w, r)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
		}
	})

	t.Run("returns not found for invalid session ID format", func(t *testing.T) {
		repo := NewMockQuizRepository()
		service := NewQuizService(*repo, nil, nil)
		testHandler := NewQuizHandler(service)

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/quizzes/invalid", nil)
		ctx := context.WithValue(r.Context(), middlewares.UserIDContextKey, 1)
		r = r.WithContext(ctx)

		testHandler.HandleRequest(w, r)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
		}
	})
}
