package podcasts

import (
	"time"

	"cribeapp.com/cribe-server/internal/clients/podcast"
)

type Podcast struct {
	ID          int       `json:"id"`
	AuthorName  string    `json:"author_name"`
	Name        string    `json:"name"`
	ImageURL    string    `json:"image_url"`
	Description string    `json:"description"`
	ExternalID  string    `json:"external_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Episodes    []Episode `json:"episodes,omitempty"`
}

type Episode struct {
	ID            int       `json:"id"`
	ExternalID    string    `json:"external_id"`
	PodcastID     int       `json:"podcast_id"`
	Name          string    `json:"name"`
	Description   string    `json:"description"`
	AudioURL      string    `json:"audio_url"`
	ImageURL      string    `json:"image_url"`
	DatePublished string    `json:"date_published"`
	Duration      int       `json:"duration"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// ExternalPodcast is an alias for the external podcast series from the client
type ExternalPodcast = podcast.ExternalPodcastSeries

// PodcastEpisode is an alias for the podcast episode from the client
type PodcastEpisode = podcast.PodcastEpisode

// PodcastWithEpisodes is an alias for the podcast with episodes from the client
type PodcastWithEpisodes = podcast.PodcastWithEpisodes

// SyncResult represents the result of a podcast sync operation
type SyncResult struct {
	TotalSynced int    `json:"total_synced"`
	New         int    `json:"new"`
	Message     string `json:"message"`
}
