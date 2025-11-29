package transcripts

import "net/http"

func HandleHTTPRequests(w http.ResponseWriter, r *http.Request) {
	service := NewServiceReady()
	handler := NewTranscriptHandler(service)

	handler.HandleRequest(w, r)
}
