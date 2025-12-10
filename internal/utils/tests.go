package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"time"

	"cribeapp.com/cribe-server/internal/core/logger"
)

var log = logger.NewUtilLogger("Utils")

const (
	TestUserID       = 1
	TestPodcastID    = 1
	TestEpisodeID    = 1
	TestTranscriptID = 1
)

func MockGetCurrentTime() time.Time {
	return time.Date(2025, time.January, 1, 1, 0, 0, 0, time.UTC)
}

func MockGetCurrentTimeISO() string {
	return MockGetCurrentTime().Format(time.RFC3339)
}

func CleanDatabase() error {
	db := NewDatabase[any](nil)
	return db.Exec("DROP SCHEMA public CASCADE; CREATE SCHEMA public;")
}

func CleanDatabaseAndRunMigrations(handlerFunc http.HandlerFunc) error {
	log.Info("Cleaning database and running migrations", nil)
	err := CleanDatabase()
	if err != nil {
		log.Error("Failed to clean database", map[string]any{
			"error": err.Error(),
		})
		return err
	}
	result := MustSendTestRequest[any](TestRequest{
		Method:      http.MethodPost,
		URL:         "/migrations",
		HandlerFunc: handlerFunc,
	})
	if result.StatusCode != http.StatusCreated {
		log.Error("Failed to run migrations", map[string]any{
			"status_code": result.StatusCode,
			"response":    result.Body,
		})
		return fmt.Errorf("failed to run migrations: status code %d", result.StatusCode)
	}
	return nil
}

// TestRequest represents a test HTTP request with its configuration
type TestRequest struct {
	Method      string
	URL         string
	Body        any
	Headers     map[string]string
	HandlerFunc http.HandlerFunc
}

// TestResponse represents the response from a test HTTP request
type TestResponse[T any] struct {
	StatusCode int
	Body       T
	Recorder   *httptest.ResponseRecorder
}

// SendTestRequest sends an HTTP request and returns the response
func SendTestRequest[T any](req TestRequest) (*TestResponse[T], error) {
	log.Debug("Starting test request", map[string]any{
		"method": req.Method,
		"url":    req.URL,
	})

	// Create request body if provided
	var bodyReader *bytes.Buffer
	if req.Body != nil {
		log.Debug("Marshaling request body", map[string]any{
			"body": req.Body,
		})
		body, err := json.Marshal(req.Body)
		if err != nil {
			log.Error("Failed to marshal request body", map[string]any{
				"error": err.Error(),
			})
			return nil, err
		}
		bodyReader = bytes.NewBuffer(body)
		log.Debug("Request body marshaled successfully", nil)
	} else {
		bodyReader = bytes.NewBuffer(nil)
	}

	// Create HTTP request
	log.Debug("Creating new HTTP request", nil)
	httpReq, err := http.NewRequest(req.Method, req.URL, bodyReader)
	if err != nil {
		log.Error("Failed to create HTTP request", map[string]any{
			"error": err.Error(),
		})
		return nil, err
	}
	log.Debug("HTTP request created successfully", nil)

	// Set default headers
	log.Debug("Setting default Content-Type header", nil)
	httpReq.Header.Set("Content-Type", "application/json")

	// Set additional headers if provided
	if len(req.Headers) > 0 {
		log.Debug("Setting additional headers", map[string]any{
			"headers": req.Headers,
		})
		for key, value := range req.Headers {
			httpReq.Header.Set(key, value)
		}
	}

	// Create response recorder
	log.Debug("Creating response recorder", nil)
	rec := httptest.NewRecorder()

	// Call handler
	log.Debug("Calling handler function", nil)
	req.HandlerFunc(rec, httpReq)
	log.Debug("Handler function completed", map[string]any{
		"status_code": rec.Code,
	})

	// Decode response based on type
	log.Debug("Decoding response body", nil)
	var responseBody T
	if err := json.NewDecoder(rec.Body).Decode(&responseBody); err != nil {
		log.Warn("Failed to decode response body", map[string]any{
			"error": err.Error(),
		})
		return nil, err
	}
	log.Debug("Response body decoded successfully", nil)

	// Return response
	log.Debug("Returning test response", nil)
	return &TestResponse[T]{
		StatusCode: rec.Code,
		Body:       responseBody,
		Recorder:   rec,
	}, nil
}

// MustSendTestRequest is a helper that panics if SendTestRequest fails
func MustSendTestRequest[T any](req TestRequest) *TestResponse[T] {
	log.Debug("Starting MustSendTestRequest", nil)
	resp, err := SendTestRequest[T](req)
	if err != nil {
		log.Error("Failed to send test request", map[string]any{
			"error": err.Error(),
		})
		panic(err) // Panic is acceptable in test utility "Must" functions
	}
	log.Debug("Test request completed successfully", nil)
	return resp
}

// CreateTestUser creates a fake user in the database for testing
func CreateTestUser(userID int) error {
	db := NewDatabase[struct{ ID int }](nil)
	_, err := db.QueryItem(
		`INSERT INTO users (id, first_name, last_name, email, password, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
		 RETURNING id`,
		userID,
		"Test",
		"User",
		fmt.Sprintf("test%d@example.com", userID),
		"$2a$10$hashedpassword",
	)
	return err
}

// CreateTestPodcast creates a fake podcast in the database for testing
func CreateTestPodcast() (int, error) {
	db := NewDatabase[struct{ ID int }](nil)
	podcast, err := db.QueryItem(
		`INSERT INTO podcasts (author_name, name, image_url, description, external_id, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
		 RETURNING id`,
		"Test Author",
		"Test Podcast",
		"https://example.com/podcast.jpg",
		"A test podcast for integration tests",
		fmt.Sprintf("test-podcast-%d", time.Now().UnixNano()),
	)
	if err != nil {
		return 0, err
	}
	return podcast.ID, nil
}

// CreateTestEpisode creates a fake episode in the database for testing
func CreateTestEpisode(podcastID int) (int, error) {
	db := NewDatabase[struct{ ID int }](nil)
	episode, err := db.QueryItem(
		`INSERT INTO episodes (external_id, podcast_id, name, description, audio_url, image_url, date_published, duration, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW())
		 RETURNING id`,
		fmt.Sprintf("test-episode-%d", time.Now().UnixNano()),
		podcastID,
		"Test Episode",
		"A test episode for integration tests",
		"https://example.com/episode.mp3",
		"https://example.com/episode.jpg",
		"2025-01-01T00:00:00Z",
		3600,
	)
	if err != nil {
		return 0, err
	}
	return episode.ID, nil
}

// CreateTestTranscript creates a fake transcript in the database for testing
func CreateTestTranscript(episodeID int) (int, error) {
	db := NewDatabase[struct{ ID int }](nil)
	transcript, err := db.QueryItem(
		`INSERT INTO transcripts (episode_id, status, created_at, completed_at)
		 VALUES ($1, $2, NOW(), NOW())
		 RETURNING id`,
		episodeID,
		"complete",
	)
	if err != nil {
		return 0, err
	}
	return transcript.ID, nil
}

// CreateTestTranscriptChunks creates fake transcript chunks in the database for testing
func CreateTestTranscriptChunks(transcriptID int) error {
	db := NewDatabase[struct{}](nil)
	chunks := []struct {
		position     int
		speakerIndex int
		startTime    float64
		endTime      float64
		text         string
	}{
		{0, 0, 0.0, 0.5, "Welcome"},
		{1, 0, 0.5, 0.7, "to"},
		{2, 0, 0.7, 0.9, "our"},
		{3, 0, 0.9, 1.2, "test"},
		{4, 0, 1.2, 1.8, "podcast"},
		{5, 0, 1.8, 2.1, "I'm"},
		{6, 0, 2.1, 2.6, "excited"},
		{7, 0, 2.6, 2.8, "to"},
		{8, 0, 2.8, 3.0, "be"},
		{9, 0, 3.0, 3.3, "here"},
		{10, 0, 3.3, 3.7, "today"},
		{11, 0, 3.7, 3.9, "to"},
		{12, 0, 3.9, 4.2, "talk"},
		{13, 0, 4.2, 4.6, "about"},
		{14, 0, 4.6, 5.1, "software"},
		{15, 0, 5.1, 5.5, "testing"},
		{16, 1, 5.5, 6.0, "Today"},
		{17, 1, 6.0, 6.3, "we'll"},
		{18, 1, 6.3, 6.8, "discuss"},
		{19, 1, 6.8, 7.3, "testing"},
		{20, 1, 7.3, 8.0, "strategies"},
		{21, 1, 8.0, 8.5, "including"},
		{22, 1, 8.5, 8.9, "unit"},
		{23, 1, 8.9, 9.3, "tests"},
		{24, 1, 9.3, 10.0, "integration"},
		{25, 1, 10.0, 10.4, "tests"},
		{26, 1, 10.4, 10.7, "and"},
		{27, 1, 10.7, 11.3, "end-to-end"},
		{28, 1, 11.3, 11.7, "testing"},
		{29, 1, 11.7, 12.3, "approaches"},
		{30, 0, 12.3, 12.6, "That"},
		{31, 0, 12.6, 13.0, "sounds"},
		{32, 0, 13.0, 13.4, "great"},
		{33, 0, 13.4, 13.8, "Let's"},
		{34, 0, 13.8, 14.1, "dive"},
		{35, 0, 14.1, 14.3, "in"},
		{36, 0, 14.3, 14.5, "and"},
		{37, 0, 14.5, 15.0, "explore"},
		{38, 0, 15.0, 15.2, "how"},
		{39, 0, 15.2, 15.5, "these"},
		{40, 0, 15.5, 16.0, "different"},
		{41, 0, 16.0, 16.5, "testing"},
		{42, 0, 16.5, 17.3, "methodologies"},
		{43, 0, 17.3, 17.5, "can"},
		{44, 0, 17.5, 18.0, "improve"},
		{45, 0, 18.0, 18.3, "code"},
		{46, 0, 18.3, 18.8, "quality"},
		{47, 0, 18.8, 19.0, "and"},
		{48, 0, 19.0, 20.1, "reliability"},
	}

	for _, chunk := range chunks {
		err := db.Exec(
			`INSERT INTO transcript_chunks (transcript_id, position, speaker_index, start_time, end_time, text)
			 VALUES ($1, $2, $3, $4, $5, $6)`,
			transcriptID,
			chunk.position,
			chunk.speakerIndex,
			chunk.startTime,
			chunk.endTime,
			chunk.text,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

// CreateTestQuestions creates fake quiz questions with options for testing
func CreateTestQuestions(episodeID int) error {
	db := NewDatabase[struct{ ID int }](nil)

	// Create first question
	question1, err := db.QueryItem(
		`INSERT INTO questions (episode_id, question_text, type, position, created_at)
		 VALUES ($1, $2, $3, $4, NOW())
		 RETURNING id`,
		episodeID,
		"What is the topic of this episode?",
		"multiple_choice",
		0,
	)
	if err != nil {
		return fmt.Errorf("failed to create question 1: %w", err)
	}

	// Create options for question 1
	options1 := []struct {
		text      string
		position  int
		isCorrect bool
	}{
		{"Testing strategies", 0, true},
		{"Cooking recipes", 1, false},
	}

	for _, opt := range options1 {
		err = db.Exec(
			`INSERT INTO question_options (question_id, option_text, position, is_correct)
			 VALUES ($1, $2, $3, $4)`,
			question1.ID, opt.text, opt.position, opt.isCorrect,
		)
		if err != nil {
			return fmt.Errorf("failed to create option for question 1: %w", err)
		}
	}

	// Create second question
	question2, err := db.QueryItem(
		`INSERT INTO questions (episode_id, question_text, type, position, created_at)
		 VALUES ($1, $2, $3, $4, NOW())
		 RETURNING id`,
		episodeID,
		"How many speakers are in this episode?",
		"multiple_choice",
		1,
	)
	if err != nil {
		return fmt.Errorf("failed to create question 2: %w", err)
	}

	// Create options for question 2
	options2 := []struct {
		text      string
		position  int
		isCorrect bool
	}{
		{"One", 0, false},
		{"Two", 1, true},
	}

	for _, opt := range options2 {
		err = db.Exec(
			`INSERT INTO question_options (question_id, option_text, position, is_correct)
			 VALUES ($1, $2, $3, $4)`,
			question2.ID, opt.text, opt.position, opt.isCorrect,
		)
		if err != nil {
			return fmt.Errorf("failed to create option for question 2: %w", err)
		}
	}

	return nil
}
