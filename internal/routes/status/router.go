package status

import (
	"net/http"
	"time"
)

func HandleHTTPRequests(w http.ResponseWriter, r *http.Request) {
	repo := *NewStatusRepository()
	service := NewStatusService(repo, time.Now)
	handler := NewStatusHandler(service)

	handler.HandleRequest(w, r)
}
