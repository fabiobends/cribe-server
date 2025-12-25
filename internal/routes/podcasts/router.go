package podcasts

import (
	"net/http"

	"cribeapp.com/cribe-server/internal/clients/podcast"
)

func HandleHTTPRequests() func(http.ResponseWriter, *http.Request) {
	repo := NewPodcastRepository()
	apiClient := podcast.NewClient()
	service := NewPodcastService(repo, apiClient)
	handler := NewPodcastHandler(service)

	return handler.HandleRequest
}
