package podcasts

import "time"

type Podcast struct {
	ID          int       `json:"id"`
	AuthorName  string    `json:"author_name"`
	Name        string    `json:"name"`
	ImageURL    string    `json:"image_url"`
	Description string    `json:"description"`
	ExternalID  string    `json:"external_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ExternalPodcast represents the structure from external podcast API provider
type ExternalPodcast struct {
	UUID        string `json:"uuid"`
	Name        string `json:"name"`
	AuthorName  string `json:"authorName"`
	ImageURL    string `json:"imageUrl"`
	Description string `json:"description"`
}

// SyncResult represents the result of a podcast sync operation
type SyncResult struct {
	TotalSynced int    `json:"total_synced"`
	New         int    `json:"new"`
	Message     string `json:"message"`
}
