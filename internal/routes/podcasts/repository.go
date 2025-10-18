package podcasts

import (
	"cribeapp.com/cribe-server/internal/core/logger"
	"cribeapp.com/cribe-server/internal/utils"
)

type PodcastRepository struct {
	*utils.Repository[Podcast]
	episodeExecutor utils.QueryExecutor[Episode]
	logger          *logger.ContextualLogger
}

type PodcastRepositoryOption func(*PodcastRepository)

// WithEpisodeExecutor creates an option to set a custom episode executor
func WithEpisodeExecutor(executor utils.QueryExecutor[Episode]) PodcastRepositoryOption {
	return func(r *PodcastRepository) {
		r.episodeExecutor = executor
	}
}

func defaultEpisodeExecutor() utils.QueryExecutor[Episode] {
	episodeRepo := utils.NewRepository[Episode]()
	return episodeRepo.Executor
}

func NewPodcastRepository(options ...utils.Option[Podcast]) *PodcastRepository {
	podcastRepo := utils.NewRepository(options...)

	pr := &PodcastRepository{
		Repository:      podcastRepo,
		episodeExecutor: defaultEpisodeExecutor(),
		logger:          logger.NewRepositoryLogger("PodcastRepository"),
	}

	return pr
}

// WithOptions applies podcast-specific options to the repository
func (r *PodcastRepository) WithOptions(options ...PodcastRepositoryOption) *PodcastRepository {
	for _, option := range options {
		option(r)
	}
	return r
}

func (r *PodcastRepository) GetPodcasts() ([]Podcast, error) {
	r.logger.Debug("Fetching all podcasts from database", nil)

	query := "SELECT * FROM podcasts ORDER BY created_at DESC"

	result, err := r.Executor.QueryList(query)
	if err != nil {
		r.logger.Error("Failed to fetch podcasts", map[string]interface{}{
			"error": err.Error(),
		})
		return result, err
	}

	r.logger.Debug("Podcasts fetched successfully", map[string]interface{}{
		"count": len(result),
	})

	return result, nil
}

func (r *PodcastRepository) GetPodcastByExternalID(externalID string) (Podcast, error) {
	r.logger.Debug("Fetching podcast by external ID", map[string]interface{}{
		"externalID": externalID,
	})

	query := "SELECT * FROM podcasts WHERE external_id = $1"

	result, err := r.Executor.QueryItem(query, externalID)
	if err != nil {
		r.logger.Error("Failed to fetch podcast by external ID", map[string]interface{}{
			"externalID": externalID,
			"error":      err.Error(),
		})
		return result, err
	}

	r.logger.Debug("Podcast found by external ID", map[string]interface{}{
		"externalID": externalID,
		"podcastID":  result.ID,
	})

	return result, nil
}

func (r *PodcastRepository) GetPodcastByID(id int) (Podcast, error) {
	r.logger.Debug("Fetching podcast by ID", map[string]interface{}{
		"id": id,
	})

	query := "SELECT * FROM podcasts WHERE id = $1"

	result, err := r.Executor.QueryItem(query, id)
	if err != nil {
		r.logger.Error("Failed to fetch podcast by ID", map[string]interface{}{
			"id":    id,
			"error": err.Error(),
		})
		return result, err
	}

	r.logger.Debug("Podcast found by ID", map[string]interface{}{
		"id": id,
	})

	return result, nil
}

func (r *PodcastRepository) UpsertPodcast(podcast ExternalPodcast) (Podcast, error) {
	r.logger.Debug("Upserting podcast", map[string]interface{}{
		"externalID": podcast.UUID,
		"name":       podcast.Name,
	})

	query := `
		INSERT INTO podcasts (author_name, name, image_url, description, external_id)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (external_id)
		DO UPDATE SET
			author_name = EXCLUDED.author_name,
			name = EXCLUDED.name,
			image_url = EXCLUDED.image_url,
			description = EXCLUDED.description,
			updated_at = NOW()
		RETURNING *
	`

	result, err := r.Executor.QueryItem(
		query,
		podcast.AuthorName,
		podcast.Name,
		podcast.ImageURL,
		podcast.Description,
		podcast.UUID,
	)

	if err != nil {
		r.logger.Error("Failed to upsert podcast", map[string]interface{}{
			"externalID": podcast.UUID,
			"error":      err.Error(),
		})
		return result, err
	}

	r.logger.Info("Podcast upserted successfully", map[string]interface{}{
		"podcastID":  result.ID,
		"externalID": podcast.UUID,
	})

	return result, nil
}

func (r *PodcastRepository) GetEpisodesByPodcastID(podcastID int) ([]Episode, error) {
	r.logger.Debug("Fetching episodes by podcast ID", map[string]interface{}{
		"podcastID": podcastID,
	})

	query := "SELECT * FROM episodes WHERE podcast_id = $1 ORDER BY date_published DESC"

	result, err := r.episodeExecutor.QueryList(query, podcastID)
	if err != nil {
		r.logger.Error("Failed to fetch episodes", map[string]interface{}{
			"podcastID": podcastID,
			"error":     err.Error(),
		})
		return result, err
	}

	r.logger.Debug("Episodes fetched successfully", map[string]interface{}{
		"podcastID": podcastID,
		"count":     len(result),
	})

	return result, nil
}

func (r *PodcastRepository) UpsertEpisode(episode PodcastEpisode, podcastID int) (Episode, error) {
	r.logger.Debug("Upserting episode", map[string]interface{}{
		"externalID": episode.UUID,
		"podcastID":  podcastID,
		"name":       episode.Name,
	})

	query := `
		INSERT INTO episodes (external_id, podcast_id, name, description, audio_url, image_url, date_published, duration)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (external_id)
		DO UPDATE SET
			name = EXCLUDED.name,
			description = EXCLUDED.description,
			audio_url = EXCLUDED.audio_url,
			image_url = EXCLUDED.image_url,
			date_published = EXCLUDED.date_published,
			duration = EXCLUDED.duration,
			updated_at = NOW()
		RETURNING *
	`

	result, err := r.episodeExecutor.QueryItem(
		query,
		episode.UUID,
		podcastID,
		episode.Name,
		episode.Description,
		episode.AudioURL,
		episode.ImageURL,
		utils.UnixToISO(episode.DatePublished),
		episode.Duration,
	)

	if err != nil {
		r.logger.Error("Failed to upsert episode", map[string]interface{}{
			"externalID": episode.UUID,
			"podcastID":  podcastID,
			"error":      err.Error(),
		})
		return result, err
	}

	r.logger.Info("Episode upserted successfully", map[string]interface{}{
		"episodeID":  result.ID,
		"externalID": episode.UUID,
		"podcastID":  podcastID,
	})

	return result, nil
}
