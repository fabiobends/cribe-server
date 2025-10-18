package podcasts

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"cribeapp.com/cribe-server/internal/core/logger"
	"cribeapp.com/cribe-server/internal/utils"
)

// HTTPClient interface for making HTTP requests (allows mocking)
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// PodcastAPIClient handles communication with external podcast API provider
type PodcastAPIClient struct {
	userID     string
	apiKey     string
	logger     *logger.ContextualLogger
	httpClient HTTPClient
	baseURL    string
}

func NewPodcastAPIClient() *PodcastAPIClient {
	return &PodcastAPIClient{
		userID:     os.Getenv("TADDY_USER_ID"),
		apiKey:     os.Getenv("TADDY_API_KEY"),
		logger:     logger.NewServiceLogger("PodcastAPIClient"),
		httpClient: &http.Client{},
		baseURL:    os.Getenv("PODCAST_API_URL"),
	}
}

type GraphQLRequest struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables,omitempty"`
}

type GraphQLResponse struct {
	Data   GraphQLData              `json:"data"`
	Errors []map[string]interface{} `json:"errors,omitempty"`
}

type GraphQLData struct {
	GetPopularContent PopularContent `json:"getPopularContent"`
}

type PopularContent struct {
	PodcastSeries []ExternalPodcastSeries `json:"podcastSeries"`
}

type ExternalPodcastSeries struct {
	UUID        string `json:"uuid"`
	Name        string `json:"name"`
	AuthorName  string `json:"authorName"`
	ImageURL    string `json:"imageUrl"`
	Description string `json:"description"`
}

type PodcastEpisode struct {
	UUID          string `json:"uuid"`
	Name          string `json:"name"`
	Description   string `json:"description"`
	AudioURL      string `json:"audioUrl"`
	DatePublished int64  `json:"datePublished"`
	Duration      int    `json:"duration"`
	ImageURL      string `json:"imageUrl"`
}

type PodcastWithEpisodes struct {
	UUID        string           `json:"uuid"`
	Name        string           `json:"name"`
	AuthorName  string           `json:"authorName"`
	ImageURL    string           `json:"imageUrl"`
	Description string           `json:"description"`
	Episodes    []PodcastEpisode `json:"episodes"`
}

type GetPodcastByIDData struct {
	GetPodcastSeries PodcastWithEpisodes `json:"getPodcastSeries"`
}

type GetEpisodeByIDData struct {
	GetPodcastEpisode PodcastEpisode `json:"getPodcastEpisode"`
}

// GetTopPodcasts fetches the top podcasts from the external API provider
func (c *PodcastAPIClient) GetTopPodcasts() ([]ExternalPodcast, error) {
	c.logger.Debug("Fetching top 10 podcasts from external API", nil)

	query := `
		query {
			getPopularContent {
				popularityRankId
				podcastSeries {
					uuid
					name
					authorName
					imageUrl
					description
				}
			}
		}
	`

	reqBody := GraphQLRequest{
		Query: query,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		c.logger.Error("Failed to marshal GraphQL request", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", c.baseURL, bytes.NewBuffer(jsonData))
	if err != nil {
		c.logger.Error("Failed to create HTTP request", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-USER-ID", c.userID)
	req.Header.Set("X-API-KEY", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Error("Failed to execute HTTP request", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			c.logger.Error("Failed to close response body", map[string]interface{}{
				"error": closeErr.Error(),
			})
		}
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.logger.Error("Failed to read response body", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		c.logger.Error("External API returned non-200 status", map[string]interface{}{
			"statusCode": resp.StatusCode,
			"response":   string(body),
		})
		return nil, fmt.Errorf("external API error: status %d", resp.StatusCode)
	}

	graphQLResp, err := utils.DecodeResponse[GraphQLResponse](string(body))
	if err != nil {
		c.logger.Error("Failed to unmarshal GraphQL response", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(graphQLResp.Errors) > 0 {
		c.logger.Error("GraphQL returned errors", map[string]interface{}{
			"errors": graphQLResp.Errors,
		})
		return nil, fmt.Errorf("graphQL errors: %v", graphQLResp.Errors)
	}

	// Convert external podcast series to standard format
	podcasts := make([]ExternalPodcast, len(graphQLResp.Data.GetPopularContent.PodcastSeries))
	for i, series := range graphQLResp.Data.GetPopularContent.PodcastSeries {
		podcasts[i] = ExternalPodcast(series)
	}

	c.logger.Info("Successfully fetched popular podcasts from external API", map[string]interface{}{
		"count": len(podcasts),
	})

	return podcasts, nil
}

func (c *PodcastAPIClient) GetPodcastByID(podcastID string) (*PodcastWithEpisodes, error) {
	c.logger.Debug("Fetching podcast by ID from external API", map[string]interface{}{
		"podcastID": podcastID,
	})

	query := `
		query {
			getPodcastSeries(uuid: "` + podcastID + `") {
				uuid
				name
				authorName
				imageUrl
				description
				episodes {
					uuid
					name
					audioUrl
					datePublished
					duration
					imageUrl
					description
				}
			}
		}
	`

	reqBody := GraphQLRequest{
		Query: query,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		c.logger.Error("Failed to marshal GraphQL request", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", c.baseURL, bytes.NewBuffer(jsonData))
	if err != nil {
		c.logger.Error("Failed to create HTTP request", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-USER-ID", c.userID)
	req.Header.Set("X-API-KEY", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Error("Failed to execute HTTP request", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			c.logger.Error("Failed to close response body", map[string]interface{}{
				"error": closeErr.Error(),
			})
		}
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.logger.Error("Failed to read response body", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		c.logger.Error("External API returned non-200 status", map[string]interface{}{
			"statusCode": resp.StatusCode,
			"response":   string(body),
		})
		return nil, fmt.Errorf("external API error: status %d", resp.StatusCode)
	}

	type GetPodcastResponse struct {
		Data   GetPodcastByIDData       `json:"data"`
		Errors []map[string]interface{} `json:"errors,omitempty"`
	}

	graphQLResp, err := utils.DecodeResponse[GetPodcastResponse](string(body))
	if err != nil {
		c.logger.Error("Failed to unmarshal GraphQL response", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(graphQLResp.Errors) > 0 {
		c.logger.Error("GraphQL returned errors", map[string]interface{}{
			"errors": graphQLResp.Errors,
		})
		return nil, fmt.Errorf("graphQL errors: %v", graphQLResp.Errors)
	}

	podcast := graphQLResp.Data.GetPodcastSeries

	c.logger.Info("Successfully fetched podcast with episodes from external API", map[string]interface{}{
		"podcastID":     podcast.UUID,
		"episodesCount": len(podcast.Episodes),
	})

	return &podcast, nil
}

func (c *PodcastAPIClient) GetEpisodeByID(podcastID, episodeID string) (*PodcastEpisode, error) {
	c.logger.Debug("Fetching episode by ID from external API", map[string]interface{}{
		"podcastID": podcastID,
		"episodeID": episodeID,
	})

	query := `
		query {
			getPodcastEpisode(uuid: "` + episodeID + `") {
				uuid
				name
				description
				audioUrl
				datePublished
				duration
			}
		}
	`

	reqBody := GraphQLRequest{
		Query: query,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		c.logger.Error("Failed to marshal GraphQL request", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", c.baseURL, bytes.NewBuffer(jsonData))
	if err != nil {
		c.logger.Error("Failed to create HTTP request", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-USER-ID", c.userID)
	req.Header.Set("X-API-KEY", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Error("Failed to execute HTTP request", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			c.logger.Error("Failed to close response body", map[string]interface{}{
				"error": closeErr.Error(),
			})
		}
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.logger.Error("Failed to read response body", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		c.logger.Error("External API returned non-200 status", map[string]interface{}{
			"statusCode": resp.StatusCode,
			"response":   string(body),
		})
		return nil, fmt.Errorf("external API error: status %d", resp.StatusCode)
	}

	type GetEpisodeResponse struct {
		Data   GetEpisodeByIDData       `json:"data"`
		Errors []map[string]interface{} `json:"errors,omitempty"`
	}

	graphQLResp, err := utils.DecodeResponse[GetEpisodeResponse](string(body))
	if err != nil {
		c.logger.Error("Failed to unmarshal GraphQL response", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(graphQLResp.Errors) > 0 {
		c.logger.Error("GraphQL returned errors", map[string]interface{}{
			"errors": graphQLResp.Errors,
		})
		return nil, fmt.Errorf("graphQL errors: %v", graphQLResp.Errors)
	}

	episode := graphQLResp.Data.GetPodcastEpisode

	c.logger.Info("Successfully fetched episode from external API", map[string]interface{}{
		"podcastID": podcastID,
		"episodeID": episode.UUID,
	})

	return &episode, nil
}
