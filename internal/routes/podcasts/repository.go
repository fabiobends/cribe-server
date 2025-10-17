package podcasts

import (
	"cribeapp.com/cribe-server/internal/core/logger"
	"cribeapp.com/cribe-server/internal/utils"
)

type PodcastRepository struct {
	*utils.Repository[Podcast]
	logger *logger.ContextualLogger
}

func NewPodcastRepository(options ...utils.Option[Podcast]) *PodcastRepository {
	repo := utils.NewRepository(options...)
	return &PodcastRepository{
		Repository: repo,
		logger:     logger.NewRepositoryLogger("PodcastRepository"),
	}
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
