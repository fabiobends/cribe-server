package podcasts

import (
	"net/http"
	"strings"

	"cribeapp.com/cribe-server/internal/utils"
)

type PodcastHandler struct {
	service PodcastService
}

func NewPodcastHandler(service *PodcastService) *PodcastHandler {
	return &PodcastHandler{service: *service}
}

func (h *PodcastHandler) HandleRequest(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.handleGet(w, r)
	case http.MethodPost:
		h.handlePost(w, r)
	default:
		utils.NotAllowed(w)
	}
}

func (h *PodcastHandler) handleGet(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/podcasts")
	path = strings.TrimPrefix(path, "/")

	if path != "" {
		h.handleGetByID(w, r, path)
		return
	}

	response, errResp := h.service.GetPodcasts()
	if errResp != nil {
		utils.EncodeResponse(w, http.StatusInternalServerError, errResp)
		return
	}
	utils.EncodeResponse(w, http.StatusOK, response)
}

func (h *PodcastHandler) handleGetByID(w http.ResponseWriter, r *http.Request, podcastID string) {
	response, errResp := h.service.GetPodcastByID(podcastID)
	if errResp != nil {
		utils.EncodeResponse(w, http.StatusInternalServerError, errResp)
		return
	}
	utils.EncodeResponse(w, http.StatusOK, response)
}

func (h *PodcastHandler) handlePost(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/podcasts")
	path = strings.TrimPrefix(path, "/")

	if path == "sync" {
		h.handleSync(w, r)
		return
	}

	// Check if path ends with /sync (for /podcasts/:id/sync)
	if strings.HasSuffix(path, "/sync") {
		podcastID := strings.TrimSuffix(path, "/sync")
		h.handleSyncEpisodes(w, r, podcastID)
		return
	}

	utils.NotFound(w, r)
}

func (h *PodcastHandler) handleSync(w http.ResponseWriter, r *http.Request) {
	response, errResp := h.service.SyncPodcasts()
	if errResp != nil {
		utils.EncodeResponse(w, http.StatusInternalServerError, errResp)
		return
	}
	utils.EncodeResponse(w, http.StatusOK, response)
}

func (h *PodcastHandler) handleSyncEpisodes(w http.ResponseWriter, r *http.Request, podcastID string) {
	response, errResp := h.service.SyncPodcastEpisodes(podcastID)
	if errResp != nil {
		utils.EncodeResponse(w, http.StatusInternalServerError, errResp)
		return
	}
	utils.EncodeResponse(w, http.StatusOK, response)
}
