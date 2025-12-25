package quizzes

import (
	"net/http"

	"cribeapp.com/cribe-server/internal/clients/llm"
	"cribeapp.com/cribe-server/internal/routes/transcripts"
)

func HandleHTTPRequests() func(http.ResponseWriter, *http.Request) {
	repo := NewQuizRepository()
	transcriptRepo := transcripts.NewTranscriptRepository()
	llmClient := llm.NewClient()

	service := NewQuizService(*repo, transcriptRepo, llmClient)
	handler := NewQuizHandler(service)

	return handler.HandleRequest
}
