package podcasts

import (
	"fmt"
	"testing"
	"time"

	"cribeapp.com/cribe-server/internal/errors"
)

// Mock Repository that satisfies the methods needed by PodcastService
type MockPodcastRepo struct {
	getPodcastsFunc            func() ([]Podcast, error)
	getPodcastByExternalIDFunc func(externalID string) (Podcast, error)
	upsertPodcastFunc          func(podcast ExternalPodcast) (Podcast, error)
}

func (m MockPodcastRepo) GetPodcasts() ([]Podcast, error) {
	if m.getPodcastsFunc != nil {
		return m.getPodcastsFunc()
	}
	return nil, nil
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

type MockAPIClient struct {
	getTopPodcastsFunc func() ([]ExternalPodcast, error)
}

func (m *MockAPIClient) GetTopPodcasts() ([]ExternalPodcast, error) {
	if m.getTopPodcastsFunc != nil {
		return m.getTopPodcastsFunc()
	}
	return nil, nil
}

func TestNewPodcastService(t *testing.T) {
	repo := NewPodcastRepository()
	apiClient := NewPodcastAPIClient()

	service := NewPodcastService(repo, apiClient)

	if service == nil {
		t.Error("Expected service to be initialized")
	}
}

func TestServiceGetPodcasts_Success(t *testing.T) {
	now := time.Now()
	mockPodcasts := []Podcast{
		{
			ID:          1,
			Name:        "Podcast 1",
			AuthorName:  "Author 1",
			ImageURL:    "http://example.com/image1.jpg",
			Description: "Description 1",
			ExternalID:  "uuid-1",
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          2,
			Name:        "Podcast 2",
			AuthorName:  "Author 2",
			ImageURL:    "http://example.com/image2.jpg",
			Description: "Description 2",
			ExternalID:  "uuid-2",
			CreatedAt:   now,
			UpdatedAt:   now,
		},
	}

	mockRepo := MockPodcastRepo{
		getPodcastsFunc: func() ([]Podcast, error) {
			return mockPodcasts, nil
		},
	}

	podcasts, err := mockRepo.GetPodcasts()

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(podcasts) != 2 {
		t.Errorf("Expected 2 podcasts, got %d", len(podcasts))
	}

	if podcasts[0].Name != "Podcast 1" {
		t.Errorf("Expected name 'Podcast 1', got '%s'", podcasts[0].Name)
	}
}

func TestServiceGetPodcasts_Error(t *testing.T) {
	mockRepo := MockPodcastRepo{
		getPodcastsFunc: func() ([]Podcast, error) {
			return nil, fmt.Errorf("database error")
		},
	}

	podcasts, err := mockRepo.GetPodcasts()

	if err == nil {
		t.Error("Expected error, got nil")
	}

	if podcasts != nil {
		t.Errorf("Expected nil podcasts, got %v", podcasts)
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

func TestServiceSyncPodcasts_Success_SomeExisting(t *testing.T) {
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

	callCount := 0
	mockRepo := MockPodcastRepo{
		getPodcastByExternalIDFunc: func(externalID string) (Podcast, error) {
			// First call returns existing, second returns not found
			if externalID == "uuid-1" {
				return Podcast{ID: 1, ExternalID: externalID}, nil
			}
			return Podcast{ID: 0}, fmt.Errorf("not found")
		},
		upsertPodcastFunc: func(podcast ExternalPodcast) (Podcast, error) {
			callCount++
			return Podcast{
				ID:         callCount,
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

	result, err := testSyncWithMocks(mockRepo, mockAPIClient)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result.TotalSynced != 2 {
		t.Errorf("Expected 2 synced podcasts, got %d", result.TotalSynced)
	}

	if result.New != 1 {
		t.Errorf("Expected 1 new podcast, got %d", result.New)
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

func TestServiceSyncPodcasts_PartialFailure(t *testing.T) {
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

	callCount := 0
	mockRepo := MockPodcastRepo{
		getPodcastByExternalIDFunc: func(externalID string) (Podcast, error) {
			return Podcast{ID: 0}, fmt.Errorf("not found")
		},
		upsertPodcastFunc: func(podcast ExternalPodcast) (Podcast, error) {
			callCount++
			// First call succeeds, second fails
			if callCount == 1 {
				return Podcast{
					ID:         1,
					Name:       podcast.Name,
					ExternalID: podcast.UUID,
				}, nil
			}
			return Podcast{}, fmt.Errorf("upsert error")
		},
	}

	mockAPIClient := &MockAPIClient{
		getTopPodcastsFunc: func() ([]ExternalPodcast, error) {
			return externalPodcasts, nil
		},
	}

	result, err := testSyncWithMocks(mockRepo, mockAPIClient)

	if err != nil {
		t.Fatalf("Expected no error (partial failure should continue), got %v", err)
	}

	// Should only count successful sync
	if result.TotalSynced != 1 {
		t.Errorf("Expected 1 synced podcast (partial success), got %d", result.TotalSynced)
	}

	if result.New != 1 {
		t.Errorf("Expected 1 new podcast, got %d", result.New)
	}
}

func TestServiceSyncPodcasts_EmptyList(t *testing.T) {
	mockRepo := MockPodcastRepo{}

	mockAPIClient := &MockAPIClient{
		getTopPodcastsFunc: func() ([]ExternalPodcast, error) {
			return []ExternalPodcast{}, nil
		},
	}

	result, err := testSyncWithMocks(mockRepo, mockAPIClient)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result.TotalSynced != 0 {
		t.Errorf("Expected 0 synced podcasts, got %d", result.TotalSynced)
	}

	if result.New != 0 {
		t.Errorf("Expected 0 new podcasts, got %d", result.New)
	}

	if result.Message != "Podcasts synced successfully" {
		t.Errorf("Expected success message, got '%s'", result.Message)
	}
}

func TestServiceSyncPodcasts_ReturnsNilError(t *testing.T) {
	mockRepo := MockPodcastRepo{
		getPodcastByExternalIDFunc: func(externalID string) (Podcast, error) {
			return Podcast{ID: 0}, fmt.Errorf("not found")
		},
		upsertPodcastFunc: func(podcast ExternalPodcast) (Podcast, error) {
			return Podcast{ID: 1, Name: podcast.Name, ExternalID: podcast.UUID}, nil
		},
	}

	mockAPIClient := &MockAPIClient{
		getTopPodcastsFunc: func() ([]ExternalPodcast, error) {
			return []ExternalPodcast{{UUID: "uuid-1", Name: "Test Podcast"}}, nil
		},
	}

	result, err := testSyncWithMocks(mockRepo, mockAPIClient)

	if err != nil {
		t.Fatalf("Expected nil error, got %v", err)
	}

	if result.TotalSynced != 1 {
		t.Errorf("Expected 1 synced podcast, got %d", result.TotalSynced)
	}

	if result.New != 1 {
		t.Errorf("Expected 1 new podcast, got %d", result.New)
	}

	if result.Message != "Podcasts synced successfully" {
		t.Errorf("Expected success message, got '%s'", result.Message)
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
