package quizzes

import (
	"fmt"
	"strings"
	"testing"

	"cribeapp.com/cribe-server/internal/clients/llm"
	"cribeapp.com/cribe-server/internal/routes/transcripts"
)

func TestQuizService_generateQuestionsWithLLM(t *testing.T) {
	tests := []struct {
		name        string
		mockResp    llm.ChatCompletionResponse
		mockError   error
		wantCount   int
		wantErr     bool
		errContains string
	}{
		{"clean JSON", makeLLMResponse(makeQuizJSON(3)), nil, 3, false, ""},
		{"markdown wrapped", makeLLMResponse("```json\n" + makeQuizJSON(1) + "\n```"), nil, 1, false, ""},
		{"no choices", llm.ChatCompletionResponse{Choices: []llm.Choice{}}, nil, 0, true, "no response from LLM"},
		{"invalid JSON", makeLLMResponse("Not JSON"), nil, 0, true, "failed to parse LLM response"},
		{"API error", llm.ChatCompletionResponse{}, fmt.Errorf("API failed"), 0, true, "API failed"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := setupQuizService(&MockLLMClient{ChatResponse: tt.mockResp, ChatError: tt.mockError}, nil)
			questions, err := svc.generateQuestionsWithLLM("test")

			if tt.wantErr {
				if err == nil || !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("got error %v, want error containing '%s'", err, tt.errContains)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if len(questions) != tt.wantCount {
					t.Errorf("got %d questions, want %d", len(questions), tt.wantCount)
				}
			}
		})
	}
}

func TestQuizService_evaluateOpenEndedAnswer(t *testing.T) {
	tests := []struct {
		name         string
		mockResp     llm.ChatCompletionResponse
		mockError    error
		wantCorrect  bool
		wantFeedback string
		wantErr      bool
	}{
		{"correct", makeLLMResponse(`{"is_correct": true, "feedback": "Great!"}`), nil, true, "Great!", false},
		{"incorrect", makeLLMResponse(`{"is_correct": false, "feedback": "Try again"}`), nil, false, "Try again", false},
		{"markdown wrapped", makeLLMResponse("```json\n" + `{"is_correct": true, "feedback": "Good!"}` + "\n```"), nil, true, "Good!", false},
		{"no choices", llm.ChatCompletionResponse{Choices: []llm.Choice{}}, nil, false, "", true},
		{"invalid JSON", makeLLMResponse("Not JSON"), nil, false, "", true},
		{"network error", llm.ChatCompletionResponse{}, fmt.Errorf("network error"), false, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := setupQuizService(&MockLLMClient{ChatResponse: tt.mockResp, ChatError: tt.mockError}, nil)
			isCorrect, feedback, err := svc.evaluateOpenEndedAnswer(Question{QuestionText: "Test"}, "Answer")

			if tt.wantErr {
				if err == nil {
					t.Error("got nil error, want error")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if isCorrect != tt.wantCorrect || feedback != tt.wantFeedback {
					t.Errorf("got (%v, %q), want (%v, %q)", isCorrect, feedback, tt.wantCorrect, tt.wantFeedback)
				}
			}
		})
	}
}

func TestQuizService_generateFeedbackWithLLM(t *testing.T) {
	tests := []struct {
		name     string
		llmResp  llm.ChatCompletionResponse
		llmErr   error
		correct  bool
		wantFeed string
	}{
		{"API error correct", llm.ChatCompletionResponse{}, fmt.Errorf("API failed"), true, "Correct answer!"},
		{"API error incorrect", llm.ChatCompletionResponse{}, fmt.Errorf("API failed"), false, "Incorrect answer!"},
		{"empty choices correct", llm.ChatCompletionResponse{Choices: []llm.Choice{}}, nil, true, "Correct answer!"},
		{"empty choices incorrect", llm.ChatCompletionResponse{Choices: []llm.Choice{}}, nil, false, "Incorrect answer!"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockTranscript := &MockTranscriptRepository{
				GetTranscriptByEpisodeIDFunc: func(int) (transcripts.Transcript, error) { return transcripts.Transcript{ID: 1}, nil },
				GetChunksByTranscriptIDFunc: func(int) ([]transcripts.TranscriptChunk, error) {
					return []transcripts.TranscriptChunk{{Text: "Test"}}, nil
				},
			}
			svc := setupQuizService(&MockLLMClient{ChatResponse: tt.llmResp, ChatError: tt.llmErr}, mockTranscript)

			feedback := svc.generateFeedbackWithLLM(Question{ID: 1, EpisodeID: 1, QuestionText: "Test?", Type: MultipleChoice}, "Answer", tt.correct)

			if feedback != tt.wantFeed {
				t.Errorf("got %q, want %q", feedback, tt.wantFeed)
			}
		})
	}
}

func TestQuizService_generateFeedbackWithLLM_Success(t *testing.T) {
	mockTranscript := &MockTranscriptRepository{
		GetTranscriptByEpisodeIDFunc: func(int) (transcripts.Transcript, error) { return transcripts.Transcript{ID: 1}, nil },
		GetChunksByTranscriptIDFunc: func(int) ([]transcripts.TranscriptChunk, error) {
			return []transcripts.TranscriptChunk{{Text: "Test transcript"}}, nil
		},
	}
	svc := setupQuizService(&MockLLMClient{ChatResponse: makeLLMResponse("Great explanation!")}, mockTranscript)

	feedback := svc.generateFeedbackWithLLM(Question{ID: 1, EpisodeID: 1, QuestionText: "Test?", Type: MultipleChoice}, "Answer", true)

	if !strings.Contains(feedback, "Great") {
		t.Errorf("got %q, want feedback containing 'Great'", feedback)
	}
}

func TestQuizService_generateFeedbackWithLLM_NilClient(t *testing.T) {
	repo := NewMockQuizRepository()
	svc := NewQuizService(*repo, nil, nil)

	feedback := svc.generateFeedbackWithLLM(Question{ID: 1, QuestionText: "Test?"}, "Answer", true)
	if feedback != "Correct answer!" {
		t.Errorf("got %q, want 'Correct answer!'", feedback)
	}

	feedback = svc.generateFeedbackWithLLM(Question{ID: 1, QuestionText: "Test?"}, "Answer", false)
	if feedback != "Incorrect answer!" {
		t.Errorf("got %q, want 'Incorrect answer!'", feedback)
	}
}
