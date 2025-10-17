package podcasts

import (
	"cribeapp.com/cribe-server/internal/errors"
)

type PodcastRepo interface {
	GetPodcasts() ([]Podcast, error)
	GetPodcastByExternalID(externalID string) (Podcast, error)
	UpsertPodcast(podcast ExternalPodcast) (Podcast, error)
}

type PodcastAPIClientInterface interface {
	GetTopPodcasts() ([]ExternalPodcast, error)
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
	result, err := s.repo.GetPodcasts()
	if err != nil {
		return nil, &errors.ErrorResponse{
			Message: errors.DatabaseError,
			Details: "Failed to retrieve podcasts",
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
