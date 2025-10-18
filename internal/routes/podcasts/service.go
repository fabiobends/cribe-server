package podcasts

import (
	"strconv"

	"cribeapp.com/cribe-server/internal/errors"
)

type PodcastRepo interface {
	GetPodcasts() ([]Podcast, error)
	GetPodcastByID(id int) (Podcast, error)
	GetPodcastByExternalID(externalID string) (Podcast, error)
	UpsertPodcast(podcast ExternalPodcast) (Podcast, error)
	GetEpisodesByPodcastID(podcastID int) ([]Episode, error)
	UpsertEpisode(episode PodcastEpisode, podcastID int) (Episode, error)
}

type PodcastAPIClientInterface interface {
	GetTopPodcasts() ([]ExternalPodcast, error)
	GetPodcastByID(podcastID string) (*PodcastWithEpisodes, error)
	GetEpisodeByID(podcastID, episodeID string) (*PodcastEpisode, error)
}

type PodcastService struct {
	repo      PodcastRepo
	apiClient PodcastAPIClientInterface
}

func NewPodcastService(repo PodcastRepo, apiClient PodcastAPIClientInterface) *PodcastService {
	return &PodcastService{
		repo:      repo,
		apiClient: apiClient,
	}
}

func (s *PodcastService) GetPodcasts() ([]Podcast, *errors.ErrorResponse) {
	// Check if podcasts already exist in database
	result, err := s.repo.GetPodcasts()
	if err != nil {
		return nil, &errors.ErrorResponse{
			Message: errors.DatabaseError,
			Details: "Failed to retrieve podcasts",
		}
	}

	// If no podcasts in database, try to fetch from external API
	// If external API fetch fails, just return empty array (valid state)
	if len(result) == 0 {
		externalPodcasts, err := s.apiClient.GetTopPodcasts()
		if err != nil {
			// Return empty array instead of error - having no podcasts is valid
			return []Podcast{}, nil
		}

		// Store podcasts in database
		for _, externalPodcast := range externalPodcasts {
			_, err := s.repo.UpsertPodcast(externalPodcast)
			if err != nil {
				// Log error but continue with other podcasts
				continue
			}
		}

		// Fetch podcasts again from database to get complete data with IDs
		result, err = s.repo.GetPodcasts()
		if err != nil {
			return nil, &errors.ErrorResponse{
				Message: errors.DatabaseError,
				Details: "Failed to retrieve podcasts after sync",
			}
		}
	}

	return result, nil
}

func (s *PodcastService) SyncPodcasts() (SyncResult, *errors.ErrorResponse) {
	externalPodcasts, err := s.apiClient.GetTopPodcasts()
	if err != nil {
		return SyncResult{}, &errors.ErrorResponse{
			Message: errors.ExternalAPIError,
			Details: "Failed to fetch podcasts from external API",
		}
	}

	// Upsert each podcast into the database
	syncedCount := 0
	newCount := 0
	for _, externalPodcast := range externalPodcasts {
		// Check if podcast exists
		existing, _ := s.repo.GetPodcastByExternalID(externalPodcast.UUID)
		isNew := existing.ID == 0

		_, err := s.repo.UpsertPodcast(externalPodcast)
		if err != nil {
			// Log error but continue with other podcasts
			continue
		}

		syncedCount++
		if isNew {
			newCount++
		}
	}

	return SyncResult{
		TotalSynced: syncedCount,
		New:         newCount,
		Message:     "Podcasts synced successfully",
	}, nil
}

func (s *PodcastService) GetPodcastByID(podcastID string) (*Podcast, *errors.ErrorResponse) {
	// Convert string ID to int
	id, err := strconv.Atoi(podcastID)
	if err != nil {
		return nil, &errors.ErrorResponse{
			Message: errors.ValidationError,
			Details: "Invalid podcast ID",
		}
	}

	// Get podcast from database
	podcast, err := s.repo.GetPodcastByID(id)
	if err != nil {
		return nil, &errors.ErrorResponse{
			Message: errors.DatabaseError,
			Details: "Failed to retrieve podcast",
		}
	}

	// Check if episodes already exist in database
	episodes, err := s.repo.GetEpisodesByPodcastID(id)
	if err != nil {
		return nil, &errors.ErrorResponse{
			Message: errors.DatabaseError,
			Details: "Failed to retrieve episodes",
		}
	}

	// If no episodes in database, fetch from external API using external_id
	if len(episodes) == 0 {
		externalPodcast, err := s.apiClient.GetPodcastByID(podcast.ExternalID)
		if err != nil {
			return nil, &errors.ErrorResponse{
				Message: errors.ExternalAPIError,
				Details: "Failed to fetch podcast episodes from external API",
			}
		}

		// Store episodes in database
		for _, episode := range externalPodcast.Episodes {
			_, err := s.repo.UpsertEpisode(episode, id)
			if err != nil {
				// Log error but continue with other episodes
				continue
			}
		}

		// Fetch episodes again from database to get complete data with IDs
		episodes, err = s.repo.GetEpisodesByPodcastID(id)
		if err != nil {
			return nil, &errors.ErrorResponse{
				Message: errors.DatabaseError,
				Details: "Failed to retrieve episodes after sync",
			}
		}
	}

	podcast.Episodes = episodes

	return &podcast, nil
}

func (s *PodcastService) SyncPodcastEpisodes(podcastID string) (SyncResult, *errors.ErrorResponse) {
	// Convert string ID to int
	id, err := strconv.Atoi(podcastID)
	if err != nil {
		return SyncResult{}, &errors.ErrorResponse{
			Message: errors.ValidationError,
			Details: "Invalid podcast ID",
		}
	}

	// Get podcast from database to get external_id
	podcast, err := s.repo.GetPodcastByID(id)
	if err != nil {
		return SyncResult{}, &errors.ErrorResponse{
			Message: errors.DatabaseError,
			Details: "Failed to retrieve podcast",
		}
	}

	// Fetch episodes from external API
	externalPodcast, err := s.apiClient.GetPodcastByID(podcast.ExternalID)
	if err != nil {
		return SyncResult{}, &errors.ErrorResponse{
			Message: errors.ExternalAPIError,
			Details: "Failed to fetch podcast episodes from external API",
		}
	}

	// Upsert each episode into the database
	syncedCount := 0
	newCount := 0
	for _, episode := range externalPodcast.Episodes {
		// Check if episode exists
		existing, _ := s.repo.GetEpisodesByPodcastID(id)
		isNew := true
		for _, e := range existing {
			if e.ExternalID == episode.UUID {
				isNew = false
				break
			}
		}

		_, err := s.repo.UpsertEpisode(episode, id)
		if err != nil {
			// Log error but continue with other episodes
			continue
		}

		syncedCount++
		if isNew {
			newCount++
		}
	}

	return SyncResult{
		TotalSynced: syncedCount,
		New:         newCount,
		Message:     "Episodes synced successfully",
	}, nil
}
