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
