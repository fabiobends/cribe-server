package quizzes

import (
	"net/http"

	"cribeapp.com/cribe-server/internal/clients/llm"
	"cribeapp.com/cribe-server/internal/routes/transcripts"
)

// HandleHTTPRequests is the entry point for all quiz-related HTTP requests
func HandleHTTPRequests(w http.ResponseWriter, r *http.Request) {
	// Initialize dependencies
	repo := NewQuizRepository()
	transcriptRepo := transcripts.NewTranscriptRepository()
	llmClient := llm.NewClient()

	service := NewQuizService(*repo, transcriptRepo, llmClient)
	handler := NewQuizHandler(service)

	handler.HandleRequest(w, r)
}
