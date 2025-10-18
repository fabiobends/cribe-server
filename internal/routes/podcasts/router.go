package podcasts

import "net/http"

func HandleHTTPRequests(w http.ResponseWriter, r *http.Request) {
	repo := NewPodcastRepository()
	apiClient := NewPodcastAPIClient()
	service := NewPodcastService(repo, apiClient)
	handler := NewPodcastHandler(service)

	handler.HandleRequest(w, r)
}
