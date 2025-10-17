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
