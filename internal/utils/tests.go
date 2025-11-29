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
