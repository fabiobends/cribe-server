//gp:build !test
package status

import (
	"net/http"
	"time"
)

func HandleHTTPRequests() func(http.ResponseWriter, *http.Request) {
	repo := *NewStatusRepository()
	service := NewStatusService(repo, time.Now)
	handler := NewStatusHandler(service)

	return handler.HandleRequest
}
