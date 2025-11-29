package podcast

import (
	"net/http"

	"cribeapp.com/cribe-server/internal/core/logger"
)

// Client handles communication with external podcast API provider
type Client struct {
	userID     string
	apiKey     string
	logger     *logger.ContextualLogger
	httpClient HTTPClient
	baseURL    string
}

// HTTPClient interface for making HTTP requests (allows mocking)
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// GraphQLRequest represents a GraphQL request
type GraphQLRequest struct {
	Query     string         `json:"query"`
	Variables map[string]any `json:"variables,omitempty"`
}

// GraphQLResponse represents a GraphQL response
type GraphQLResponse struct {
	Data   GraphQLData      `json:"data"`
	Errors []map[string]any `json:"errors,omitempty"`
}

// GraphQLData contains the data from a GraphQL response
type GraphQLData struct {
	GetPopularContent PopularContent `json:"getPopularContent"`
}

// PopularContent contains popular podcast series
type PopularContent struct {
	PodcastSeries []ExternalPodcastSeries `json:"podcastSeries"`
}

// ExternalPodcastSeries represents a podcast series from the external API
type ExternalPodcastSeries struct {
	UUID        string `json:"uuid"`
	Name        string `json:"name"`
	AuthorName  string `json:"authorName"`
	ImageURL    string `json:"imageUrl"`
	Description string `json:"description"`
}

// PodcastEpisode represents a podcast episode
type PodcastEpisode struct {
	UUID          string `json:"uuid"`
	Name          string `json:"name"`
	Description   string `json:"description"`
	AudioURL      string `json:"audioUrl"`
	DatePublished int64  `json:"datePublished"`
	Duration      int    `json:"duration"`
	ImageURL      string `json:"imageUrl"`
}

// PodcastWithEpisodes represents a podcast with its episodes
type PodcastWithEpisodes struct {
	UUID        string           `json:"uuid"`
	Name        string           `json:"name"`
	AuthorName  string           `json:"authorName"`
	ImageURL    string           `json:"imageUrl"`
	Description string           `json:"description"`
	Episodes    []PodcastEpisode `json:"episodes"`
}

// GetPodcastByIDData represents the data structure for GetPodcastByID query
type GetPodcastByIDData struct {
	GetPodcastSeries PodcastWithEpisodes `json:"getPodcastSeries"`
}

// GetEpisodeByIDData represents the data structure for GetEpisodeByID query
type GetEpisodeByIDData struct {
	GetPodcastEpisode PodcastEpisode `json:"getPodcastEpisode"`
}
