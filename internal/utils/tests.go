package utils

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"time"

	"cribeapp.com/cribe-server/internal/core/logger"
)

var testLog = logger.NewUtilLogger("TestUtils")

func MockGetCurrentTime() time.Time {
	return time.Date(2025, time.January, 1, 1, 0, 0, 0, time.UTC)
}

func MockGetCurrentTimeISO() string {
	return MockGetCurrentTime().Format(time.RFC3339)
}

func CleanDatabase() error {
	return Exec("DROP SCHEMA public CASCADE; CREATE SCHEMA public;")
}

func CleanDatabaseAndRunMigrations(handlerFunc http.HandlerFunc) {
	testLog.Info("Cleaning database and running migrations", nil)
	err := CleanDatabase()
	if err != nil {
		testLog.Error("Failed to clean database", map[string]interface{}{
			"error": err.Error(),
		})
		panic(err) // Use panic instead of log.Fatalf for test utilities
	}
	result := MustSendTestRequest[any](TestRequest{
		Method:      http.MethodPost,
		URL:         "/migrations",
		HandlerFunc: handlerFunc,
	})
	if result.StatusCode != http.StatusCreated {
		testLog.Error("Failed to run migrations", map[string]interface{}{
			"status_code": result.StatusCode,
			"response":    result.Body,
		})
		panic("Failed to run migrations") // Use panic instead of log.Fatalf for test utilities
	}
}

// TestRequest represents a test HTTP request with its configuration
type TestRequest struct {
	Method      string
	URL         string
	Body        interface{}
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
	testLog.Debug("Starting test request", map[string]interface{}{
		"method": req.Method,
		"url":    req.URL,
	})

	// Create request body if provided
	var bodyReader *bytes.Buffer
	if req.Body != nil {
		testLog.Debug("Marshaling request body", map[string]interface{}{
			"body": req.Body,
		})
		body, err := json.Marshal(req.Body)
		if err != nil {
			testLog.Error("Failed to marshal request body", map[string]interface{}{
				"error": err.Error(),
			})
			return nil, err
		}
		bodyReader = bytes.NewBuffer(body)
		testLog.Debug("Request body marshaled successfully", nil)
	} else {
		bodyReader = bytes.NewBuffer(nil)
	}

	// Create HTTP request
	testLog.Debug("Creating new HTTP request", nil)
	httpReq, err := http.NewRequest(req.Method, req.URL, bodyReader)
	if err != nil {
		testLog.Error("Failed to create HTTP request", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, err
	}
	testLog.Debug("HTTP request created successfully", nil)

	// Set default headers
	testLog.Debug("Setting default Content-Type header", nil)
	httpReq.Header.Set("Content-Type", "application/json")

	// Set additional headers if provided
	if len(req.Headers) > 0 {
		testLog.Debug("Setting additional headers", map[string]interface{}{
			"headers": req.Headers,
		})
		for key, value := range req.Headers {
			httpReq.Header.Set(key, value)
		}
	}

	// Create response recorder
	testLog.Debug("Creating response recorder", nil)
	rec := httptest.NewRecorder()

	// Call handler
	testLog.Debug("Calling handler function", nil)
	req.HandlerFunc(rec, httpReq)
	testLog.Debug("Handler function completed", map[string]interface{}{
		"status_code": rec.Code,
	})

	// Decode response based on type
	testLog.Debug("Decoding response body", nil)
	var responseBody T
	if err := json.NewDecoder(rec.Body).Decode(&responseBody); err != nil {
		testLog.Warn("Failed to decode response body", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, err
	}
	testLog.Debug("Response body decoded successfully", nil)

	// Return response
	testLog.Debug("Returning test response", nil)
	return &TestResponse[T]{
		StatusCode: rec.Code,
		Body:       responseBody,
		Recorder:   rec,
	}, nil
}

// MustSendTestRequest is a helper that panics if SendTestRequest fails
func MustSendTestRequest[T any](req TestRequest) *TestResponse[T] {
	testLog.Debug("Starting MustSendTestRequest", nil)
	resp, err := SendTestRequest[T](req)
	if err != nil {
		testLog.Error("Failed to send test request", map[string]interface{}{
			"error": err.Error(),
		})
		panic(err) // Use panic instead of log.Fatalf for test utilities
	}
	testLog.Debug("Test request completed successfully", nil)
	return resp
}
