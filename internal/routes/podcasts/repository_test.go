package podcasts

import (
	"context"
	"fmt"
	"testing"
	"time"

	"cribeapp.com/cribe-server/internal/utils"
	"github.com/pashagolub/pgxmock/v4"
)

func TestNewPodcastRepository(t *testing.T) {
	repo := NewPodcastRepository()

	if repo == nil {
		t.Fatal("Expected repository to be initialized")
	}

	if repo.logger == nil {
		t.Error("Expected logger to be initialized")
	}
}

func TestGetPodcasts_Success(t *testing.T) {
	conn, _ := pgxmock.NewConn()
	defer func() { _ = conn.Close(context.Background()) }()

	db := utils.NewDatabase[Podcast](conn)
	executor := utils.QueryExecutor[Podcast]{
		QueryItem: db.QueryItem,
		QueryList: db.QueryList,
		Exec:      db.Exec,
	}
	repo := NewPodcastRepository(utils.WithQueryExecutor(executor))

	now := time.Now()
	rows := pgxmock.NewRows([]string{"id", "author_name", "name", "image_url", "description", "external_id", "created_at", "updated_at"}).
		AddRow(1, "Author 1", "Podcast 1", "http://example.com/image1.jpg", "Description 1", "uuid-1", now, now).
		AddRow(2, "Author 2", "Podcast 2", "http://example.com/image2.jpg", "Description 2", "uuid-2", now, now)

	conn.ExpectQuery("SELECT \\* FROM podcasts ORDER BY created_at DESC").WillReturnRows(rows)

	podcasts, err := repo.GetPodcasts()

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(podcasts) != 2 {
		t.Errorf("Expected 2 podcasts, got %d", len(podcasts))
	}

	if podcasts[0].Name != "Podcast 1" {
		t.Errorf("Expected name 'Podcast 1', got '%s'", podcasts[0].Name)
	}

	if err := conn.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet expectations: %v", err)
	}
}

func TestGetPodcasts_Error(t *testing.T) {
	conn, _ := pgxmock.NewConn()
	defer func() { _ = conn.Close(context.Background()) }()

	db := utils.NewDatabase[Podcast](conn)
	executor := utils.QueryExecutor[Podcast]{
		QueryItem: db.QueryItem,
		QueryList: db.QueryList,
		Exec:      db.Exec,
	}
	repo := NewPodcastRepository(utils.WithQueryExecutor(executor))

	conn.ExpectQuery("SELECT \\* FROM podcasts ORDER BY created_at DESC").
		WillReturnError(fmt.Errorf("database error"))

	podcasts, err := repo.GetPodcasts()

	if err == nil {
		t.Error("Expected error, got nil")
	}

	if len(podcasts) != 0 {
		t.Errorf("Expected 0 podcasts on error, got %d", len(podcasts))
	}

	if err := conn.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet expectations: %v", err)
	}
}

func TestGetPodcastByExternalID_Success(t *testing.T) {
	conn, _ := pgxmock.NewConn()
	defer func() { _ = conn.Close(context.Background()) }()

	db := utils.NewDatabase[Podcast](conn)
	executor := utils.QueryExecutor[Podcast]{
		QueryItem: db.QueryItem,
		QueryList: db.QueryList,
		Exec:      db.Exec,
	}
	repo := NewPodcastRepository(utils.WithQueryExecutor(executor))

	now := time.Now()
	rows := pgxmock.NewRows([]string{"id", "author_name", "name", "image_url", "description", "external_id", "created_at", "updated_at"}).
		AddRow(1, "Author 1", "Podcast 1", "http://example.com/image1.jpg", "Description 1", "uuid-1", now, now)

	conn.ExpectQuery("SELECT \\* FROM podcasts WHERE external_id = \\$1").
		WithArgs("uuid-1").
		WillReturnRows(rows)

	podcast, err := repo.GetPodcastByExternalID("uuid-1")

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if podcast.ExternalID != "uuid-1" {
		t.Errorf("Expected external_id 'uuid-1', got '%s'", podcast.ExternalID)
	}

	if podcast.Name != "Podcast 1" {
		t.Errorf("Expected name 'Podcast 1', got '%s'", podcast.Name)
	}

	if err := conn.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet expectations: %v", err)
	}
}

func TestGetPodcastByExternalID_NotFound(t *testing.T) {
	conn, _ := pgxmock.NewConn()
	defer func() { _ = conn.Close(context.Background()) }()

	db := utils.NewDatabase[Podcast](conn)
	executor := utils.QueryExecutor[Podcast]{
		QueryItem: db.QueryItem,
		QueryList: db.QueryList,
		Exec:      db.Exec,
	}
	repo := NewPodcastRepository(utils.WithQueryExecutor(executor))

	conn.ExpectQuery("SELECT \\* FROM podcasts WHERE external_id = \\$1").
		WithArgs("nonexistent-uuid").
		WillReturnError(fmt.Errorf("no rows in result set"))

	podcast, err := repo.GetPodcastByExternalID("nonexistent-uuid")

	if err == nil {
		t.Error("Expected error, got nil")
	}

	if podcast.ID != 0 {
		t.Errorf("Expected podcast ID to be 0, got %d", podcast.ID)
	}

	if err := conn.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet expectations: %v", err)
	}
}

func TestUpsertPodcast_Insert(t *testing.T) {
	conn, _ := pgxmock.NewConn()
	defer func() { _ = conn.Close(context.Background()) }()

	db := utils.NewDatabase[Podcast](conn)
	executor := utils.QueryExecutor[Podcast]{
		QueryItem: db.QueryItem,
		QueryList: db.QueryList,
		Exec:      db.Exec,
	}
	repo := NewPodcastRepository(utils.WithQueryExecutor(executor))

	externalPodcast := ExternalPodcast{
		UUID:        "uuid-1",
		Name:        "New Podcast",
		AuthorName:  "New Author",
		ImageURL:    "http://example.com/new.jpg",
		Description: "New Description",
	}

	now := time.Now()
	rows := pgxmock.NewRows([]string{"id", "author_name", "name", "image_url", "description", "external_id", "created_at", "updated_at"}).
		AddRow(1, externalPodcast.AuthorName, externalPodcast.Name, externalPodcast.ImageURL, externalPodcast.Description, externalPodcast.UUID, now, now)

	conn.ExpectQuery("INSERT INTO podcasts").
		WithArgs(externalPodcast.AuthorName, externalPodcast.Name, externalPodcast.ImageURL, externalPodcast.Description, externalPodcast.UUID).
		WillReturnRows(rows)

	podcast, err := repo.UpsertPodcast(externalPodcast)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if podcast.ID != 1 {
		t.Errorf("Expected ID 1, got %d", podcast.ID)
	}

	if podcast.Name != externalPodcast.Name {
		t.Errorf("Expected name '%s', got '%s'", externalPodcast.Name, podcast.Name)
	}

	if err := conn.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet expectations: %v", err)
	}
}

func TestUpsertPodcast_Update(t *testing.T) {
	conn, _ := pgxmock.NewConn()
	defer func() { _ = conn.Close(context.Background()) }()

	db := utils.NewDatabase[Podcast](conn)
	executor := utils.QueryExecutor[Podcast]{
		QueryItem: db.QueryItem,
		QueryList: db.QueryList,
		Exec:      db.Exec,
	}
	repo := NewPodcastRepository(utils.WithQueryExecutor(executor))

	externalPodcast := ExternalPodcast{
		UUID:        "uuid-1",
		Name:        "Updated Podcast",
		AuthorName:  "Updated Author",
		ImageURL:    "http://example.com/updated.jpg",
		Description: "Updated Description",
	}

	now := time.Now()
	rows := pgxmock.NewRows([]string{"id", "author_name", "name", "image_url", "description", "external_id", "created_at", "updated_at"}).
		AddRow(1, externalPodcast.AuthorName, externalPodcast.Name, externalPodcast.ImageURL, externalPodcast.Description, externalPodcast.UUID, now, now)

	conn.ExpectQuery("INSERT INTO podcasts").
		WithArgs(externalPodcast.AuthorName, externalPodcast.Name, externalPodcast.ImageURL, externalPodcast.Description, externalPodcast.UUID).
		WillReturnRows(rows)

	podcast, err := repo.UpsertPodcast(externalPodcast)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if podcast.Name != "Updated Podcast" {
		t.Errorf("Expected name 'Updated Podcast', got '%s'", podcast.Name)
	}

	if err := conn.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet expectations: %v", err)
	}
}

func TestUpsertPodcast_Error(t *testing.T) {
	conn, _ := pgxmock.NewConn()
	defer func() { _ = conn.Close(context.Background()) }()

	db := utils.NewDatabase[Podcast](conn)
	executor := utils.QueryExecutor[Podcast]{
		QueryItem: db.QueryItem,
		QueryList: db.QueryList,
		Exec:      db.Exec,
	}
	repo := NewPodcastRepository(utils.WithQueryExecutor(executor))

	externalPodcast := ExternalPodcast{
		UUID:        "uuid-1",
		Name:        "Podcast",
		AuthorName:  "Author",
		ImageURL:    "http://example.com/image.jpg",
		Description: "Description",
	}

	conn.ExpectQuery("INSERT INTO podcasts").
		WithArgs(externalPodcast.AuthorName, externalPodcast.Name, externalPodcast.ImageURL, externalPodcast.Description, externalPodcast.UUID).
		WillReturnError(fmt.Errorf("database error"))

	podcast, err := repo.UpsertPodcast(externalPodcast)

	if err == nil {
		t.Error("Expected error, got nil")
	}

	if podcast.ID != 0 {
		t.Errorf("Expected podcast ID to be 0 on error, got %d", podcast.ID)
	}

	if err := conn.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet expectations: %v", err)
	}
}

func TestRepositoryGetPodcastByID_Success(t *testing.T) {
	conn, _ := pgxmock.NewConn()
	defer func() { _ = conn.Close(context.Background()) }()

	db := utils.NewDatabase[Podcast](conn)
	executor := utils.QueryExecutor[Podcast]{
		QueryItem: db.QueryItem,
		QueryList: db.QueryList,
		Exec:      db.Exec,
	}
	repo := NewPodcastRepository(utils.WithQueryExecutor(executor))

	now := time.Now()
	rows := pgxmock.NewRows([]string{"id", "author_name", "name", "image_url", "description", "external_id", "created_at", "updated_at"}).
		AddRow(1, "Author", "Podcast", "url", "desc", "uuid-1", now, now)

	conn.ExpectQuery("SELECT \\* FROM podcasts WHERE id = \\$1").
		WithArgs(1).
		WillReturnRows(rows)

	podcast, err := repo.GetPodcastByID(1)

	if err != nil || podcast.ID != 1 || podcast.Name != "Podcast" {
		t.Errorf("GetPodcastByID failed: err=%v, podcast=%v", err, podcast)
	}

	if err := conn.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet expectations: %v", err)
	}
}

func TestGetEpisodesByPodcastID_Success(t *testing.T) {
	conn, _ := pgxmock.NewConn()
	defer func() { _ = conn.Close(context.Background()) }()

	db := utils.NewDatabase[Podcast](conn)
	podcastExecutor := utils.QueryExecutor[Podcast]{
		QueryItem: db.QueryItem,
		QueryList: db.QueryList,
		Exec:      db.Exec,
	}

	episodeDB := utils.NewDatabase[Episode](conn)
	episodeExecutor := utils.QueryExecutor[Episode]{
		QueryItem: episodeDB.QueryItem,
		QueryList: episodeDB.QueryList,
		Exec:      episodeDB.Exec,
	}

	repo := NewPodcastRepository(utils.WithQueryExecutor(podcastExecutor)).
		WithOptions(WithEpisodeExecutor(episodeExecutor))

	now := time.Now()
	datePublished := "2009-02-13T23:31:30Z"
	rows := pgxmock.NewRows([]string{"id", "external_id", "podcast_id", "name", "description", "audio_url", "image_url", "date_published", "duration", "created_at", "updated_at"}).
		AddRow(1, "episode-uuid-1", 1, "Episode 1", "Description 1", "http://example.com/audio1.mp3", "http://example.com/image1.jpg", datePublished, 3600, now, now).
		AddRow(2, "episode-uuid-2", 1, "Episode 2", "Description 2", "http://example.com/audio2.mp3", "http://example.com/image2.jpg", datePublished, 2400, now, now)

	conn.ExpectQuery("SELECT \\* FROM episodes WHERE podcast_id = \\$1 ORDER BY date_published DESC").
		WithArgs(1).
		WillReturnRows(rows)

	episodes, err := repo.GetEpisodesByPodcastID(1)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(episodes) != 2 {
		t.Errorf("Expected 2 episodes, got %d", len(episodes))
	}

	if episodes[0].Name != "Episode 1" {
		t.Errorf("Expected name 'Episode 1', got '%s'", episodes[0].Name)
	}

	if episodes[0].PodcastID != 1 {
		t.Errorf("Expected podcast_id 1, got %d", episodes[0].PodcastID)
	}

	if err := conn.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet expectations: %v", err)
	}
}

func TestGetEpisodesByPodcastID_Error(t *testing.T) {
	conn, _ := pgxmock.NewConn()
	defer func() { _ = conn.Close(context.Background()) }()

	db := utils.NewDatabase[Podcast](conn)
	podcastExecutor := utils.QueryExecutor[Podcast]{
		QueryItem: db.QueryItem,
		QueryList: db.QueryList,
		Exec:      db.Exec,
	}

	episodeDB := utils.NewDatabase[Episode](conn)
	episodeExecutor := utils.QueryExecutor[Episode]{
		QueryItem: episodeDB.QueryItem,
		QueryList: episodeDB.QueryList,
		Exec:      episodeDB.Exec,
	}

	repo := NewPodcastRepository(utils.WithQueryExecutor(podcastExecutor)).
		WithOptions(WithEpisodeExecutor(episodeExecutor))

	conn.ExpectQuery("SELECT \\* FROM episodes WHERE podcast_id = \\$1 ORDER BY date_published DESC").
		WithArgs(1).
		WillReturnError(fmt.Errorf("database error"))

	episodes, err := repo.GetEpisodesByPodcastID(1)

	if err == nil {
		t.Error("Expected error, got nil")
	}

	if len(episodes) != 0 {
		t.Errorf("Expected 0 episodes on error, got %d", len(episodes))
	}

	if err := conn.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet expectations: %v", err)
	}
}

func TestUpsertEpisode_Insert(t *testing.T) {
	conn, _ := pgxmock.NewConn()
	defer func() { _ = conn.Close(context.Background()) }()

	db := utils.NewDatabase[Podcast](conn)
	podcastExecutor := utils.QueryExecutor[Podcast]{
		QueryItem: db.QueryItem,
		QueryList: db.QueryList,
		Exec:      db.Exec,
	}

	episodeDB := utils.NewDatabase[Episode](conn)
	episodeExecutor := utils.QueryExecutor[Episode]{
		QueryItem: episodeDB.QueryItem,
		QueryList: episodeDB.QueryList,
		Exec:      episodeDB.Exec,
	}

	repo := NewPodcastRepository(utils.WithQueryExecutor(podcastExecutor)).
		WithOptions(WithEpisodeExecutor(episodeExecutor))

	externalEpisode := PodcastEpisode{
		UUID:          "episode-uuid-1",
		Name:          "New Episode",
		Description:   "New Description",
		AudioURL:      "http://example.com/audio.mp3",
		ImageURL:      "http://example.com/image.jpg",
		DatePublished: int64(1234567890),
		Duration:      3600,
	}

	now := time.Now()
	datePublished := utils.UnixToISO(externalEpisode.DatePublished)
	rows := pgxmock.NewRows([]string{"id", "external_id", "podcast_id", "name", "description", "audio_url", "image_url", "date_published", "duration", "created_at", "updated_at"}).
		AddRow(1, externalEpisode.UUID, 1, externalEpisode.Name, externalEpisode.Description, externalEpisode.AudioURL, externalEpisode.ImageURL, datePublished, externalEpisode.Duration, now, now)

	conn.ExpectQuery("INSERT INTO episodes").
		WithArgs(externalEpisode.UUID, 1, externalEpisode.Name, externalEpisode.Description, externalEpisode.AudioURL, externalEpisode.ImageURL, datePublished, externalEpisode.Duration).
		WillReturnRows(rows)

	episode, err := repo.UpsertEpisode(externalEpisode, 1)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if episode.ID != 1 {
		t.Errorf("Expected ID 1, got %d", episode.ID)
	}

	if episode.Name != externalEpisode.Name {
		t.Errorf("Expected name '%s', got '%s'", externalEpisode.Name, episode.Name)
	}

	if episode.PodcastID != 1 {
		t.Errorf("Expected podcast_id 1, got %d", episode.PodcastID)
	}

	if err := conn.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet expectations: %v", err)
	}
}

func TestUpsertEpisode_Update(t *testing.T) {
	conn, _ := pgxmock.NewConn()
	defer func() { _ = conn.Close(context.Background()) }()

	db := utils.NewDatabase[Podcast](conn)
	podcastExecutor := utils.QueryExecutor[Podcast]{
		QueryItem: db.QueryItem,
		QueryList: db.QueryList,
		Exec:      db.Exec,
	}

	episodeDB := utils.NewDatabase[Episode](conn)
	episodeExecutor := utils.QueryExecutor[Episode]{
		QueryItem: episodeDB.QueryItem,
		QueryList: episodeDB.QueryList,
		Exec:      episodeDB.Exec,
	}

	repo := NewPodcastRepository(utils.WithQueryExecutor(podcastExecutor)).
		WithOptions(WithEpisodeExecutor(episodeExecutor))

	externalEpisode := PodcastEpisode{
		UUID:          "episode-uuid-1",
		Name:          "Updated Episode",
		Description:   "Updated Description",
		AudioURL:      "http://example.com/updated-audio.mp3",
		ImageURL:      "http://example.com/updated-image.jpg",
		DatePublished: int64(1234567890),
		Duration:      4200,
	}

	now := time.Now()
	datePublished := utils.UnixToISO(externalEpisode.DatePublished)
	rows := pgxmock.NewRows([]string{"id", "external_id", "podcast_id", "name", "description", "audio_url", "image_url", "date_published", "duration", "created_at", "updated_at"}).
		AddRow(1, externalEpisode.UUID, 1, externalEpisode.Name, externalEpisode.Description, externalEpisode.AudioURL, externalEpisode.ImageURL, datePublished, externalEpisode.Duration, now, now)

	conn.ExpectQuery("INSERT INTO episodes").
		WithArgs(externalEpisode.UUID, 1, externalEpisode.Name, externalEpisode.Description, externalEpisode.AudioURL, externalEpisode.ImageURL, datePublished, externalEpisode.Duration).
		WillReturnRows(rows)

	episode, err := repo.UpsertEpisode(externalEpisode, 1)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if episode.Name != "Updated Episode" {
		t.Errorf("Expected name 'Updated Episode', got '%s'", episode.Name)
	}

	if episode.Duration != 4200 {
		t.Errorf("Expected duration 4200, got %d", episode.Duration)
	}

	if err := conn.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet expectations: %v", err)
	}
}

func TestUpsertEpisode_Error(t *testing.T) {
	conn, _ := pgxmock.NewConn()
	defer func() { _ = conn.Close(context.Background()) }()

	db := utils.NewDatabase[Podcast](conn)
	podcastExecutor := utils.QueryExecutor[Podcast]{
		QueryItem: db.QueryItem,
		QueryList: db.QueryList,
		Exec:      db.Exec,
	}

	episodeDB := utils.NewDatabase[Episode](conn)
	episodeExecutor := utils.QueryExecutor[Episode]{
		QueryItem: episodeDB.QueryItem,
		QueryList: episodeDB.QueryList,
		Exec:      episodeDB.Exec,
	}

	repo := NewPodcastRepository(utils.WithQueryExecutor(podcastExecutor)).
		WithOptions(WithEpisodeExecutor(episodeExecutor))

	externalEpisode := PodcastEpisode{
		UUID:          "episode-uuid-1",
		Name:          "Episode",
		Description:   "Description",
		AudioURL:      "http://example.com/audio.mp3",
		ImageURL:      "http://example.com/image.jpg",
		DatePublished: int64(1234567890),
		Duration:      3600,
	}

	datePublished := utils.UnixToISO(externalEpisode.DatePublished)
	conn.ExpectQuery("INSERT INTO episodes").
		WithArgs(externalEpisode.UUID, 1, externalEpisode.Name, externalEpisode.Description, externalEpisode.AudioURL, externalEpisode.ImageURL, datePublished, externalEpisode.Duration).
		WillReturnError(fmt.Errorf("database error"))

	episode, err := repo.UpsertEpisode(externalEpisode, 1)

	if err == nil {
		t.Error("Expected error, got nil")
	}

	if episode.ID != 0 {
		t.Errorf("Expected episode ID to be 0 on error, got %d", episode.ID)
	}

	if err := conn.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet expectations: %v", err)
	}
}
