package transcripts

import (
	"net/http"

	"cribeapp.com/cribe-server/internal/clients/llm"
	"cribeapp.com/cribe-server/internal/clients/transcription"
)

func HandleHTTPRequests() func(http.ResponseWriter, *http.Request) {
	transcriptionClient := transcription.NewClient()
	llmClient := llm.NewClient()

	service := NewService(transcriptionClient, llmClient)
	handler := NewTranscriptHandler(service)

	return handler.HandleRequest
}
