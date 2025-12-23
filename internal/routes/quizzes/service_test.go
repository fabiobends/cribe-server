package quizzes

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"cribeapp.com/cribe-server/internal/routes/transcripts"
)

// Helpers

func makeQuizJSON(questionCount int) string {
	if questionCount == 1 {
		return `{"questions": [{"question_text": "What is Go?", "type": "multiple_choice", "options": [{"text": "A language", "is_correct": true}, {"text": "A tool", "is_correct": false}]}]}`
	}
	return `{"questions": [{"question_text": "What is Go?", "type": "multiple_choice", "options": [{"text": "A programming language", "is_correct": true}, {"text": "A game", "is_correct": false}, {"text": "A car", "is_correct": false}, {"text": "A bird", "is_correct": false}]}, {"question_text": "Go is statically typed", "type": "true_false", "options": [{"text": "True", "is_correct": true}, {"text": "False", "is_correct": false}]}, {"question_text": "Explain goroutines", "type": "open_ended"}]}`
}

func setupQuizService(llmClient *MockLLMClient, transcriptRepo *MockTranscriptRepository) *QuizService {
	repo := NewMockQuizRepository()
	if transcriptRepo == nil {
		transcriptRepo = &MockTranscriptRepository{}
	}
	if llmClient == nil {
		llmClient = &MockLLMClient{}
	}
	return NewQuizService(*repo, transcriptRepo, llmClient)
}

// Tests

func TestQuizService_UpdateSessionStatus(t *testing.T) {
	tests := []struct {
		name          string
		setup         bool
		userID        int
		reqUserID     int
		status        SessionStatus
		wantErr       bool
		checkComplete bool
	}{
		{"complete", true, 1, 1, Completed, false, true},
		{"abandon", true, 1, 1, Abandoned, false, true},
		{"not found", false, 1, 1, Completed, true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewMockQuizRepository()
			if tt.setup {
				_, _ = repo.CreateSession(UserQuizSession{ID: 1, UserID: tt.userID, EpisodeID: 1, Status: InProgress})
			}
			svc := NewQuizService(*repo, nil, nil)

			session, errResp := svc.UpdateSessionStatus(1, tt.reqUserID, tt.status)

			if tt.wantErr {
				if errResp == nil {
					t.Error("got nil error, want error")
				}
				return
			}

			if errResp != nil {
				t.Errorf("unexpected error: %v", errResp.Details)
			}
			if session.Status != tt.status {
				t.Errorf("got status %s, want %s", session.Status, tt.status)
			}
			if tt.checkComplete && session.CompletedAt == nil {
				t.Error("CompletedAt not set")
			}
		})
	}

	t.Run("CompletedAt timing", func(t *testing.T) {
		repo := NewMockQuizRepository()
		_, _ = repo.CreateSession(UserQuizSession{ID: 1, UserID: 1, EpisodeID: 1, Status: InProgress})
		svc := NewQuizService(*repo, nil, nil)

		before := time.Now()
		session, _ := svc.UpdateSessionStatus(1, 1, Completed)
		after := time.Now()

		if session.CompletedAt == nil || session.CompletedAt.Before(before) || session.CompletedAt.After(after) {
			t.Error("CompletedAt not set correctly")
		}
	})

	t.Run("CompletedAt persists", func(t *testing.T) {
		repo := NewMockQuizRepository()
		_, _ = repo.CreateSession(UserQuizSession{ID: 1, UserID: 1, EpisodeID: 1, Status: InProgress})
		svc := NewQuizService(*repo, nil, nil)

		completed, _ := svc.UpdateSessionStatus(1, 1, Completed)
		updated, _ := svc.UpdateSessionStatus(1, 1, InProgress)

		if updated.CompletedAt == nil || !updated.CompletedAt.Equal(*completed.CompletedAt) {
			t.Error("CompletedAt changed unexpectedly")
		}
	})
}

func TestQuizService_GetSessionsWithDetailsByUserID_DatabaseError(t *testing.T) {
	repo := NewMockQuizRepository()

	repo.sessionRepo.Executor.QueryList = func(query string, args ...any) ([]UserQuizSession, error) {
		return nil, fmt.Errorf("database connection failed")
	}

	svc := NewQuizService(*repo, nil, nil)
	sessions, errResp := svc.GetSessionsWithDetailsByUserID(1)

	if errResp == nil {
		t.Fatal("expected error when database fetch fails")
	}
	if sessions != nil {
		t.Errorf("expected nil sessions, got %d", len(sessions))
	}
	if errResp.Message != "Database error" {
		t.Errorf("expected 'Database error', got '%s'", errResp.Message)
	}
	if errResp.Details != "Failed to fetch sessions" {
		t.Errorf("expected 'Failed to fetch sessions', got '%s'", errResp.Details)
	}
}

func TestQuizService_evaluateAnswer(t *testing.T) {
	t.Run("MultipleChoice - correct answer", func(t *testing.T) {
		mockLLM := &MockLLMClient{ChatResponse: makeLLMResponse("Great job!")}
		svc := setupQuizService(mockLLM, nil)
		correctOptionID := 1
		question := Question{ID: 1, Type: MultipleChoice, Options: []QuestionOption{{ID: 1, IsCorrect: true}}}
		request := SubmitAnswerRequest{QuestionID: 1, SelectedOptionID: &correctOptionID}
		isCorrect, _, err := svc.evaluateAnswer(question, request)
		if err != nil || !isCorrect {
			t.Errorf("expected correct answer, got err=%v isCorrect=%v", err, isCorrect)
		}
	})

	t.Run("MultipleChoice - incorrect answer", func(t *testing.T) {
		svc := setupQuizService(nil, nil)
		wrongOptionID := 2
		question := Question{ID: 1, Type: MultipleChoice, Options: []QuestionOption{{ID: 1, OptionText: "4", IsCorrect: true}, {ID: 2, IsCorrect: false}}}
		request := SubmitAnswerRequest{QuestionID: 1, SelectedOptionID: &wrongOptionID}
		isCorrect, feedback, err := svc.evaluateAnswer(question, request)
		if err != nil || isCorrect || !strings.Contains(feedback, "Incorrect") {
			t.Errorf("expected incorrect with feedback, got err=%v isCorrect=%v feedback=%s", err, isCorrect, feedback)
		}
	})

	t.Run("OpenEnded - evaluates with LLM", func(t *testing.T) {
		mockLLM := &MockLLMClient{ChatResponse: makeLLMResponse(`{"is_correct": true, "feedback": "Well explained!"}`)}
		svc := setupQuizService(mockLLM, nil)
		textAnswer := "Answer"
		question := Question{ID: 1, Type: OpenEnded}
		request := SubmitAnswerRequest{QuestionID: 1, TextAnswer: &textAnswer}
		isCorrect, feedback, err := svc.evaluateAnswer(question, request)
		if err != nil || !isCorrect || feedback != "Well explained!" {
			t.Errorf("expected correct with feedback, got err=%v isCorrect=%v feedback=%s", err, isCorrect, feedback)
		}
	})

	t.Run("OpenEnded - nil TextAnswer", func(t *testing.T) {
		svc := setupQuizService(nil, nil)
		question := Question{ID: 1, Type: OpenEnded}
		request := SubmitAnswerRequest{QuestionID: 1, TextAnswer: nil}
		isCorrect, feedback, err := svc.evaluateAnswer(question, request)
		if err != nil || isCorrect || feedback != "Text answer is required for open-ended questions" {
			t.Errorf("expected error feedback for nil TextAnswer, got err=%v isCorrect=%v feedback=%s", err, isCorrect, feedback)
		}
	})

	t.Run("Unknown question type", func(t *testing.T) {
		svc := setupQuizService(nil, nil)

		optionID := 1
		question := Question{
			ID:           1,
			QuestionText: "Test",
			Type:         "UnknownType",
		}

		request := SubmitAnswerRequest{
			QuestionID:       1,
			SelectedOptionID: &optionID,
		}

		isCorrect, feedback, err := svc.evaluateAnswer(question, request)

		if err == nil {
			t.Error("expected error for unknown question type")
		}
		if isCorrect {
			t.Error("expected incorrect answer")
		}
		if feedback != "Unknown question type" {
			t.Errorf("expected 'Unknown question type', got: %s", feedback)
		}
	})
}

func TestQuizService_GenerateAndSaveQuestions(t *testing.T) {
	t.Run("successful generation and saving with multiple question types", func(t *testing.T) {
		// Create LLM response with different question types
		llmResp := LLMQuestionsResponse{Questions: []LLMQuestion{
			{
				QuestionText: "What is Go?",
				Type:         "multiple_choice",
				Options: []LLMQuestionOption{
					{Text: "A programming language", IsCorrect: true},
					{Text: "A game", IsCorrect: false},
					{Text: "A car", IsCorrect: false},
					{Text: "A bird", IsCorrect: false},
				},
			},
			{
				QuestionText: "Go is statically typed",
				Type:         "true_false",
				Options: []LLMQuestionOption{
					{Text: "True", IsCorrect: true},
					{Text: "False", IsCorrect: false},
				},
			},
			{
				QuestionText: "Explain goroutines",
				Type:         "open_ended",
			},
		}}
		jsonBytes, _ := json.Marshal(llmResp)

		mockLLM := &MockLLMClient{ChatResponse: makeLLMResponse(string(jsonBytes))}
		mockTranscript := &MockTranscriptRepository{
			GetTranscriptByEpisodeIDFunc: func(int) (transcripts.Transcript, error) {
				return transcripts.Transcript{ID: 1, Status: "complete"}, nil
			},
			GetChunksByTranscriptIDFunc: func(int) ([]transcripts.TranscriptChunk, error) {
				return []transcripts.TranscriptChunk{{Text: "Test transcript"}}, nil
			},
		}

		repo := NewMockQuizRepository()
		svc := NewQuizService(*repo, mockTranscript, mockLLM)

		// Call GetOrCreateSessionWithDetails which triggers question generation
		detail, errResp := svc.GetOrCreateSessionWithDetails(1, 1)

		if errResp != nil {
			t.Fatalf("unexpected error: %v", errResp.Details)
		}

		// Verify all questions were saved
		if len(detail.Questions) != 3 {
			t.Errorf("expected 3 questions, got %d", len(detail.Questions))
		}

		// Verify multiple choice question and options
		mcQuestion := detail.Questions[0]
		if mcQuestion.QuestionText != "What is Go?" {
			t.Errorf("expected 'What is Go?', got %q", mcQuestion.QuestionText)
		}
		if mcQuestion.Type != MultipleChoice {
			t.Errorf("expected type %q, got %q", MultipleChoice, mcQuestion.Type)
		}
		if len(mcQuestion.Options) != 4 {
			t.Errorf("expected 4 options, got %d", len(mcQuestion.Options))
		}
		// Verify correct option is marked correctly
		foundCorrect := false
		for _, opt := range mcQuestion.Options {
			if opt.OptionText == "A programming language" && opt.IsCorrect {
				foundCorrect = true
			}
		}
		if !foundCorrect {
			t.Error("expected to find correct option marked as IsCorrect")
		}

		// Verify true/false question
		tfQuestion := detail.Questions[1]
		if tfQuestion.Type != TrueFalse {
			t.Errorf("expected type %q, got %q", TrueFalse, tfQuestion.Type)
		}
		if len(tfQuestion.Options) != 2 {
			t.Errorf("expected 2 options, got %d", len(tfQuestion.Options))
		}

		// Verify open-ended question has no options
		oeQuestion := detail.Questions[2]
		if oeQuestion.Type != OpenEnded {
			t.Errorf("expected type %q, got %q", OpenEnded, oeQuestion.Type)
		}
		if len(oeQuestion.Options) != 0 {
			t.Errorf("expected 0 options for open-ended, got %d", len(oeQuestion.Options))
		}
	})

	t.Run("LLM generation fails", func(t *testing.T) {
		mockLLM := &MockLLMClient{ChatError: fmt.Errorf("API error")}
		mockTranscript := &MockTranscriptRepository{
			GetTranscriptByEpisodeIDFunc: func(int) (transcripts.Transcript, error) {
				return transcripts.Transcript{ID: 1, Status: "complete"}, nil
			},
			GetChunksByTranscriptIDFunc: func(int) ([]transcripts.TranscriptChunk, error) {
				return []transcripts.TranscriptChunk{{Text: "Test"}}, nil
			},
		}

		repo := NewMockQuizRepository()
		svc := NewQuizService(*repo, mockTranscript, mockLLM)

		_, errResp := svc.GetOrCreateSessionWithDetails(1, 1)

		if errResp == nil {
			t.Fatal("expected error, got nil")
		}
		if errResp.Details != "Failed to generate questions" {
			t.Errorf("expected 'Failed to generate questions', got %q", errResp.Details)
		}
	})

	t.Run("question save continues on option save failure", func(t *testing.T) {
		llmResp := LLMQuestionsResponse{Questions: []LLMQuestion{
			{
				QuestionText: "Test question",
				Type:         "multiple_choice",
				Options: []LLMQuestionOption{
					{Text: "Option 1", IsCorrect: true},
					{Text: "Option 2", IsCorrect: false},
				},
			},
		}}
		jsonBytes, _ := json.Marshal(llmResp)

		mockLLM := &MockLLMClient{ChatResponse: makeLLMResponse(string(jsonBytes))}
		mockTranscript := &MockTranscriptRepository{
			GetTranscriptByEpisodeIDFunc: func(int) (transcripts.Transcript, error) {
				return transcripts.Transcript{ID: 1, Status: "complete"}, nil
			},
			GetChunksByTranscriptIDFunc: func(int) ([]transcripts.TranscriptChunk, error) {
				return []transcripts.TranscriptChunk{{Text: "Test"}}, nil
			},
		}

		repo := NewMockQuizRepository()

		// Make option creation fail
		callCount := 0
		originalQueryItem := repo.optionRepo.Executor.QueryItem
		repo.optionRepo.Executor.QueryItem = func(query string, args ...any) (QuestionOption, error) {
			callCount++
			// Fail the first option save
			if callCount == 1 {
				return QuestionOption{}, fmt.Errorf("option save failed")
			}
			return originalQueryItem(query, args...)
		}

		svc := NewQuizService(*repo, mockTranscript, mockLLM)

		detail, errResp := svc.GetOrCreateSessionWithDetails(1, 1)

		if errResp != nil {
			t.Fatalf("unexpected error: %v", errResp.Details)
		}

		// Question should still be saved even if option save fails
		if len(detail.Questions) != 1 {
			t.Errorf("expected 1 question, got %d", len(detail.Questions))
		}

		// Only one option should be saved (the second one succeeded)
		if len(detail.Questions[0].Options) != 1 {
			t.Errorf("expected 1 option (one failed), got %d", len(detail.Questions[0].Options))
		}
	})
}
