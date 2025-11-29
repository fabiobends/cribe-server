package podcast

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

// NewClient creates a new podcast API client
func NewClient() *Client {
	return &Client{
		userID:     os.Getenv("TADDY_USER_ID"),
		apiKey:     os.Getenv("TADDY_API_KEY"),
		logger:     logger.NewServiceLogger("PodcastClient"),
		httpClient: &http.Client{},
		baseURL:    os.Getenv("PODCAST_API_URL"),
	}
}

// GetTopPodcasts fetches the top podcasts from the external API provider
func (c *Client) GetTopPodcasts() ([]ExternalPodcastSeries, error) {
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
		c.logger.Error("Failed to marshal GraphQL request", map[string]any{
			"error": err.Error(),
		})
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", c.baseURL, bytes.NewBuffer(jsonData))
	if err != nil {
		c.logger.Error("Failed to create HTTP request", map[string]any{
			"error": err.Error(),
		})
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-USER-ID", c.userID)
	req.Header.Set("X-API-KEY", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Error("Failed to execute HTTP request", map[string]any{
			"error": err.Error(),
		})
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			c.logger.Error("Failed to close response body", map[string]any{
				"error": closeErr.Error(),
			})
		}
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.logger.Error("Failed to read response body", map[string]any{
			"error": err.Error(),
		})
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		c.logger.Error("External API returned non-200 status", map[string]any{
			"statusCode": resp.StatusCode,
			"response":   string(body),
		})
		return nil, fmt.Errorf("external API error: status %d", resp.StatusCode)
	}

	graphQLResp, err := utils.DecodeResponse[GraphQLResponse](string(body))
	if err != nil {
		c.logger.Error("Failed to unmarshal GraphQL response", map[string]any{
			"error": err.Error(),
		})
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(graphQLResp.Errors) > 0 {
		c.logger.Error("GraphQL returned errors", map[string]any{
			"errors": graphQLResp.Errors,
		})
		return nil, fmt.Errorf("graphQL errors: %v", graphQLResp.Errors)
	}

	podcasts := graphQLResp.Data.GetPopularContent.PodcastSeries

	c.logger.Info("Successfully fetched popular podcasts from external API", map[string]any{
		"count": len(podcasts),
	})

	return podcasts, nil
}

// GetPodcastByID fetches a podcast with its episodes by ID from the external API
func (c *Client) GetPodcastByID(podcastID string) (*PodcastWithEpisodes, error) {
	c.logger.Debug("Fetching podcast by ID from external API", map[string]any{
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
		c.logger.Error("Failed to marshal GraphQL request", map[string]any{
			"error": err.Error(),
		})
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", c.baseURL, bytes.NewBuffer(jsonData))
	if err != nil {
		c.logger.Error("Failed to create HTTP request", map[string]any{
			"error": err.Error(),
		})
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-USER-ID", c.userID)
	req.Header.Set("X-API-KEY", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Error("Failed to execute HTTP request", map[string]any{
			"error": err.Error(),
		})
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			c.logger.Error("Failed to close response body", map[string]any{
				"error": closeErr.Error(),
			})
		}
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.logger.Error("Failed to read response body", map[string]any{
			"error": err.Error(),
		})
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		c.logger.Error("External API returned non-200 status", map[string]any{
			"statusCode": resp.StatusCode,
			"response":   string(body),
		})
		return nil, fmt.Errorf("external API error: status %d", resp.StatusCode)
	}

	type GetPodcastResponse struct {
		Data   GetPodcastByIDData `json:"data"`
		Errors []map[string]any   `json:"errors,omitempty"`
	}

	graphQLResp, err := utils.DecodeResponse[GetPodcastResponse](string(body))
	if err != nil {
		c.logger.Error("Failed to unmarshal GraphQL response", map[string]any{
			"error": err.Error(),
		})
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(graphQLResp.Errors) > 0 {
		c.logger.Error("GraphQL returned errors", map[string]any{
			"errors": graphQLResp.Errors,
		})
		return nil, fmt.Errorf("graphQL errors: %v", graphQLResp.Errors)
	}

	podcast := graphQLResp.Data.GetPodcastSeries

	c.logger.Info("Successfully fetched podcast with episodes from external API", map[string]any{
		"podcastID":     podcast.UUID,
		"episodesCount": len(podcast.Episodes),
	})

	return &podcast, nil
}

// GetEpisodeByID fetches a specific episode by ID from the external API
func (c *Client) GetEpisodeByID(podcastID, episodeID string) (*PodcastEpisode, error) {
	c.logger.Debug("Fetching episode by ID from external API", map[string]any{
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
		c.logger.Error("Failed to marshal GraphQL request", map[string]any{
			"error": err.Error(),
		})
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", c.baseURL, bytes.NewBuffer(jsonData))
	if err != nil {
		c.logger.Error("Failed to create HTTP request", map[string]any{
			"error": err.Error(),
		})
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-USER-ID", c.userID)
	req.Header.Set("X-API-KEY", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Error("Failed to execute HTTP request", map[string]any{
			"error": err.Error(),
		})
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			c.logger.Error("Failed to close response body", map[string]any{
				"error": closeErr.Error(),
			})
		}
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.logger.Error("Failed to read response body", map[string]any{
			"error": err.Error(),
		})
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		c.logger.Error("External API returned non-200 status", map[string]any{
			"statusCode": resp.StatusCode,
			"response":   string(body),
		})
		return nil, fmt.Errorf("external API error: status %d", resp.StatusCode)
	}

	type GetEpisodeResponse struct {
		Data   GetEpisodeByIDData `json:"data"`
		Errors []map[string]any   `json:"errors,omitempty"`
	}

	graphQLResp, err := utils.DecodeResponse[GetEpisodeResponse](string(body))
	if err != nil {
		c.logger.Error("Failed to unmarshal GraphQL response", map[string]any{
			"error": err.Error(),
		})
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(graphQLResp.Errors) > 0 {
		c.logger.Error("GraphQL returned errors", map[string]any{
			"errors": graphQLResp.Errors,
		})
		return nil, fmt.Errorf("graphQL errors: %v", graphQLResp.Errors)
	}

	episode := graphQLResp.Data.GetPodcastEpisode

	c.logger.Info("Successfully fetched episode from external API", map[string]any{
		"podcastID": podcastID,
		"episodeID": episode.UUID,
	})

	return &episode, nil
}
