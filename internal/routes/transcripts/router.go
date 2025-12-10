package transcripts

import (
	"net/http"

	"cribeapp.com/cribe-server/internal/clients/llm"
	"cribeapp.com/cribe-server/internal/clients/transcription"
)

func HandleHTTPRequests(w http.ResponseWriter, r *http.Request) {
	transcriptionClient := transcription.NewClient()
	llmClient := llm.NewClient()

	service := NewService(transcriptionClient, llmClient)
	handler := NewTranscriptHandler(service)

	handler.HandleRequest(w, r)
}
