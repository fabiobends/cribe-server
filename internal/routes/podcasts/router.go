package podcasts

import (
	"net/http"

	"cribeapp.com/cribe-server/internal/clients/podcast"
)

func HandleHTTPRequests(w http.ResponseWriter, r *http.Request) {
	repo := NewPodcastRepository()
	apiClient := podcast.NewClient()
	service := NewPodcastService(repo, apiClient)
	handler := NewPodcastHandler(service)

	handler.HandleRequest(w, r)
}
