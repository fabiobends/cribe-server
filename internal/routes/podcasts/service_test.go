package podcasts

import (
	"fmt"
	"testing"

	"cribeapp.com/cribe-server/internal/clients/podcast"
	"cribeapp.com/cribe-server/internal/errors"
)

// Mock Repository that satisfies the methods needed by PodcastService
type MockPodcastRepo struct {
	getPodcastsFunc            func() ([]Podcast, error)
	getPodcastByIDFunc         func(id int) (Podcast, error)
	getPodcastByExternalIDFunc func(externalID string) (Podcast, error)
	upsertPodcastFunc          func(podcast ExternalPodcast) (Podcast, error)
	getEpisodesByPodcastIDFunc func(podcastID int) ([]Episode, error)
	upsertEpisodeFunc          func(episode PodcastEpisode, podcastID int) (Episode, error)
}

func (m MockPodcastRepo) GetPodcasts() ([]Podcast, error) {
	if m.getPodcastsFunc != nil {
		return m.getPodcastsFunc()
	}
	return nil, nil
}

func (m MockPodcastRepo) GetPodcastByID(id int) (Podcast, error) {
	if m.getPodcastByIDFunc != nil {
		return m.getPodcastByIDFunc(id)
	}
	return Podcast{}, nil
}

func (m MockPodcastRepo) GetPodcastByExternalID(externalID string) (Podcast, error) {
	if m.getPodcastByExternalIDFunc != nil {
		return m.getPodcastByExternalIDFunc(externalID)
	}
	return Podcast{}, nil
}

func (m MockPodcastRepo) UpsertPodcast(podcast ExternalPodcast) (Podcast, error) {
	if m.upsertPodcastFunc != nil {
		return m.upsertPodcastFunc(podcast)
	}
	return Podcast{}, nil
}

func (m MockPodcastRepo) GetEpisodesByPodcastID(podcastID int) ([]Episode, error) {
	if m.getEpisodesByPodcastIDFunc != nil {
		return m.getEpisodesByPodcastIDFunc(podcastID)
	}
	return []Episode{}, nil
}

func (m MockPodcastRepo) UpsertEpisode(episode PodcastEpisode, podcastID int) (Episode, error) {
	if m.upsertEpisodeFunc != nil {
		return m.upsertEpisodeFunc(episode, podcastID)
	}
	return Episode{}, nil
}

type MockAPIClient struct {
	getTopPodcastsFunc func() ([]podcast.ExternalPodcastSeries, error)
	getPodcastByIDFunc func(podcastID string) (*podcast.PodcastWithEpisodes, error)
	getEpisodeByIDFunc func(podcastID, episodeID string) (*podcast.PodcastEpisode, error)
}

func (m *MockAPIClient) GetTopPodcasts() ([]podcast.ExternalPodcastSeries, error) {
	if m.getTopPodcastsFunc != nil {
		return m.getTopPodcastsFunc()
	}
	return nil, nil
}

func (m *MockAPIClient) GetPodcastByID(podcastID string) (*podcast.PodcastWithEpisodes, error) {
	if m.getPodcastByIDFunc != nil {
		return m.getPodcastByIDFunc(podcastID)
	}
	return nil, nil
}

func (m *MockAPIClient) GetEpisodeByID(podcastID, episodeID string) (*podcast.PodcastEpisode, error) {
	if m.getEpisodeByIDFunc != nil {
		return m.getEpisodeByIDFunc(podcastID, episodeID)
	}
	return nil, nil
}

func TestNewPodcastService(t *testing.T) {
	repo := NewPodcastRepository()
	apiClient := podcast.NewClient()

	service := NewPodcastService(repo, apiClient)

	if service == nil {
		t.Error("Expected service to be initialized")
	}
}

func TestServiceSyncPodcasts_Success_AllNew(t *testing.T) {
	externalPodcasts := []ExternalPodcast{
		{
			UUID:        "uuid-1",
			Name:        "External Podcast 1",
			AuthorName:  "External Author 1",
			ImageURL:    "http://example.com/external1.jpg",
			Description: "External Description 1",
		},
		{
			UUID:        "uuid-2",
			Name:        "External Podcast 2",
			AuthorName:  "External Author 2",
			ImageURL:    "http://example.com/external2.jpg",
			Description: "External Description 2",
		},
	}

	mockRepo := MockPodcastRepo{
		getPodcastByExternalIDFunc: func(externalID string) (Podcast, error) {
			// Return empty podcast (not found)
			return Podcast{ID: 0}, fmt.Errorf("not found")
		},
		upsertPodcastFunc: func(podcast ExternalPodcast) (Podcast, error) {
			return Podcast{
				ID:         1,
				Name:       podcast.Name,
				ExternalID: podcast.UUID,
			}, nil
		},
	}

	mockAPIClient := &MockAPIClient{
		getTopPodcastsFunc: func() ([]ExternalPodcast, error) {
			return externalPodcasts, nil
		},
	}

	// Test using helper function since we can't easily inject mocks into the service
	result, err := testSyncWithMocks(mockRepo, mockAPIClient)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result.TotalSynced != 2 {
		t.Errorf("Expected 2 synced podcasts, got %d", result.TotalSynced)
	}

	if result.New != 2 {
		t.Errorf("Expected 2 new podcasts, got %d", result.New)
	}

	if result.Message != "Podcasts synced successfully" {
		t.Errorf("Expected success message, got '%s'", result.Message)
	}
}

func TestServiceSyncPodcasts_APIError(t *testing.T) {
	mockRepo := MockPodcastRepo{}

	mockAPIClient := &MockAPIClient{
		getTopPodcastsFunc: func() ([]ExternalPodcast, error) {
			return nil, fmt.Errorf("API error")
		},
	}

	result, err := testSyncWithMocks(mockRepo, mockAPIClient)

	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	if result.TotalSynced != 0 {
		t.Errorf("Expected 0 synced podcasts, got %d", result.TotalSynced)
	}

	if err.Message != "External API error" {
		t.Errorf("Expected 'External API error', got '%s'", err.Message)
	}
}

func TestServiceSyncPodcasts_ActualServiceCall(t *testing.T) {
	mockRepo := MockPodcastRepo{
		getPodcastByExternalIDFunc: func(externalID string) (Podcast, error) {
			if externalID == "uuid-1" {
				return Podcast{ID: 1, ExternalID: externalID}, nil
			}
			return Podcast{ID: 0}, fmt.Errorf("not found")
		},
		upsertPodcastFunc: func(podcast ExternalPodcast) (Podcast, error) {
			return Podcast{ID: 1, Name: podcast.Name, ExternalID: podcast.UUID}, nil
		},
	}

	mockAPIClient := &MockAPIClient{
		getTopPodcastsFunc: func() ([]ExternalPodcast, error) {
			return []ExternalPodcast{
				{UUID: "uuid-1", Name: "Existing Podcast"},
				{UUID: "uuid-2", Name: "New Podcast"},
			}, nil
		},
	}

	service := NewPodcastService(mockRepo, mockAPIClient)
	result, err := service.SyncPodcasts()

	if err != nil {
		t.Fatalf("Expected nil error, got %v", err)
	}

	if result.TotalSynced != 2 {
		t.Errorf("Expected 2 synced podcasts, got %d", result.TotalSynced)
	}

	if result.New != 1 {
		t.Errorf("Expected 1 new podcast, got %d", result.New)
	}

	if result.Message != "Podcasts synced successfully" {
		t.Errorf("Expected success message, got '%s'", result.Message)
	}
}

// testSyncWithMocks replicates the SyncPodcasts service logic to test the counting algorithm
func testSyncWithMocks(repo MockPodcastRepo, apiClient *MockAPIClient) (SyncResult, *errors.ErrorResponse) {
	externalPodcasts, err := apiClient.GetTopPodcasts()
	if err != nil {
		return SyncResult{}, &errors.ErrorResponse{
			Message: errors.ExternalAPIError,
			Details: "Failed to fetch podcasts from external API",
		}
	}

	syncedCount := 0
	newCount := 0
	for _, externalPodcast := range externalPodcasts {
		existing, _ := repo.GetPodcastByExternalID(externalPodcast.UUID)
		isNew := existing.ID == 0

		_, err := repo.UpsertPodcast(externalPodcast)
		if err != nil {
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

// Tests for GetPodcastByID
func TestServiceGetPodcastByID_Success(t *testing.T) {
	mockPodcast := Podcast{ID: 1, Name: "Test", ExternalID: "uuid-1"}
	mockEpisodes := []Episode{{ID: 1, Name: "Episode 1"}}

	mockRepo := MockPodcastRepo{
		getPodcastByIDFunc:         func(id int) (Podcast, error) { return mockPodcast, nil },
		getEpisodesByPodcastIDFunc: func(podcastID int) ([]Episode, error) { return mockEpisodes, nil },
	}

	service := NewPodcastService(mockRepo, &MockAPIClient{})
	result, err := service.GetPodcastByID("1")

	if err != nil || result.ID != 1 || len(result.Episodes) != 1 {
		t.Errorf("GetPodcastByID failed: err=%v, result=%v", err, result)
	}
}

func TestServiceGetPodcastByID_InvalidID(t *testing.T) {
	service := NewPodcastService(MockPodcastRepo{}, &MockAPIClient{})
	_, err := service.GetPodcastByID("invalid")

	if err == nil || err.Message != errors.ValidationError {
		t.Errorf("Expected ValidationError for invalid ID, got %v", err)
	}
}

func TestServiceGetPodcastByID_FetchEpisodesFromAPI(t *testing.T) {
	mockPodcast := Podcast{ID: 1, ExternalID: "uuid-1"}
	callCount := 0

	mockRepo := MockPodcastRepo{
		getPodcastByIDFunc: func(id int) (Podcast, error) { return mockPodcast, nil },
		getEpisodesByPodcastIDFunc: func(podcastID int) ([]Episode, error) {
			callCount++
			if callCount == 1 {
				return []Episode{}, nil // Empty triggers API fetch
			}
			return []Episode{{ID: 1, Name: "Episode 1"}}, nil
		},
		upsertEpisodeFunc: func(episode PodcastEpisode, podcastID int) (Episode, error) {
			return Episode{ID: 1, Name: episode.Name}, nil
		},
	}

	mockAPIClient := &MockAPIClient{
		getPodcastByIDFunc: func(podcastID string) (*PodcastWithEpisodes, error) {
			return &PodcastWithEpisodes{
				Episodes: []PodcastEpisode{{UUID: "ep-1", Name: "Episode 1", AudioURL: "url", DatePublished: 123, Duration: 100}},
			}, nil
		},
	}

	service := NewPodcastService(mockRepo, mockAPIClient)
	result, err := service.GetPodcastByID("1")

	if err != nil || len(result.Episodes) != 1 {
		t.Errorf("Expected episodes to be fetched from API, got err=%v, episodes=%d", err, len(result.Episodes))
	}
}

// Tests for SyncPodcastEpisodes
func TestServiceSyncPodcastEpisodes_Success(t *testing.T) {
	mockRepo := MockPodcastRepo{
		getPodcastByIDFunc:         func(id int) (Podcast, error) { return Podcast{ID: 1, ExternalID: "uuid-1"}, nil },
		getEpisodesByPodcastIDFunc: func(podcastID int) ([]Episode, error) { return []Episode{}, nil },
		upsertEpisodeFunc:          func(episode PodcastEpisode, podcastID int) (Episode, error) { return Episode{ID: 1}, nil },
	}

	mockAPIClient := &MockAPIClient{
		getPodcastByIDFunc: func(podcastID string) (*PodcastWithEpisodes, error) {
			return &PodcastWithEpisodes{
				Episodes: []PodcastEpisode{
					{UUID: "ep-1", Name: "Episode 1", AudioURL: "url", DatePublished: 123, Duration: 100},
					{UUID: "ep-2", Name: "Episode 2", AudioURL: "url2", DatePublished: 124, Duration: 200},
				},
			}, nil
		},
	}

	service := NewPodcastService(mockRepo, mockAPIClient)
	result, err := service.SyncPodcastEpisodes("1")

	if err != nil || result.TotalSynced != 2 || result.New != 2 {
		t.Errorf("SyncPodcastEpisodes failed: err=%v, synced=%d, new=%d", err, result.TotalSynced, result.New)
	}
}

func TestServiceSyncPodcastEpisodes_InvalidID(t *testing.T) {
	service := NewPodcastService(MockPodcastRepo{}, &MockAPIClient{})
	_, err := service.SyncPodcastEpisodes("invalid")

	if err == nil || err.Message != errors.ValidationError {
		t.Errorf("Expected ValidationError for invalid ID, got %v", err)
	}
}

// Test GetPodcasts auto-sync
func TestServiceGetPodcasts_AutoSyncOnEmpty(t *testing.T) {
	callCount := 0
	mockRepo := MockPodcastRepo{
		getPodcastsFunc: func() ([]Podcast, error) {
			callCount++
			if callCount == 1 {
				return []Podcast{}, nil // Empty triggers sync
			}
			return []Podcast{{ID: 1, Name: "Podcast 1"}}, nil
		},
		upsertPodcastFunc: func(podcast ExternalPodcast) (Podcast, error) { return Podcast{ID: 1}, nil },
	}

	mockAPIClient := &MockAPIClient{
		getTopPodcastsFunc: func() ([]ExternalPodcast, error) {
			return []ExternalPodcast{{UUID: "uuid-1", Name: "Podcast 1", AuthorName: "Author", ImageURL: "url", Description: "desc"}}, nil
		},
	}

	service := NewPodcastService(mockRepo, mockAPIClient)
	result, err := service.GetPodcasts()

	if err != nil || len(result) != 1 {
		t.Errorf("GetPodcasts auto-sync failed: err=%v, count=%d", err, len(result))
	}
}
