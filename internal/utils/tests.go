package utils

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"time"
)

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
	log.Printf("Cleaning database and running migrations")
	err := CleanDatabase()
	if err != nil {
		log.Fatalf("Failed to clean database: %v", err)
	}
	result := MustSendTestRequest[any](TestRequest{
		Method:      http.MethodPost,
		URL:         "/migrations",
		HandlerFunc: handlerFunc,
	})
	if result.StatusCode != http.StatusCreated {
		log.Fatalf("Failed to run migrations: %v", result.Body)
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
	log.Printf("Starting test request: Method=%s, URL=%s", req.Method, req.URL)

	// Create request body if provided
	var bodyReader *bytes.Buffer
	if req.Body != nil {
		log.Printf("Marshaling request body: %+v", req.Body)
		body, err := json.Marshal(req.Body)
		if err != nil {
			log.Fatalf("Failed to marshal request body: %v", err)
			return nil, err
		}
		bodyReader = bytes.NewBuffer(body)
		log.Printf("Request body marshaled successfully")
	} else {
		bodyReader = bytes.NewBuffer(nil)
	}

	// Create HTTP request
	log.Printf("Creating new HTTP request")
	httpReq, err := http.NewRequest(req.Method, req.URL, bodyReader)
	if err != nil {
		log.Fatalf("Failed to create HTTP request: %v", err)
		return nil, err
	}
	log.Printf("HTTP request created successfully")

	// Set default headers
	log.Printf("Setting default Content-Type header")
	httpReq.Header.Set("Content-Type", "application/json")

	// Set additional headers if provided
	if len(req.Headers) > 0 {
		log.Printf("Setting additional headers: %v", req.Headers)
		for key, value := range req.Headers {
			httpReq.Header.Set(key, value)
		}
	}

	// Create response recorder
	log.Printf("Creating response recorder")
	rec := httptest.NewRecorder()

	// Call handler
	log.Printf("Calling handler function")
	req.HandlerFunc(rec, httpReq)
	log.Printf("Handler function completed. Status code: %d", rec.Code)

	// Decode response based on type
	log.Printf("Decoding response body")
	var responseBody T
	if err := json.NewDecoder(rec.Body).Decode(&responseBody); err != nil {
		log.Printf("Failed to decode response body: %v", err)
		return nil, err
	}
	log.Printf("Response body decoded successfully")

	// Return response
	log.Printf("Returning test response")
	return &TestResponse[T]{
		StatusCode: rec.Code,
		Body:       responseBody,
		Recorder:   rec,
	}, nil
}

// MustSendTestRequest is a helper that panics if SendTestRequest fails
func MustSendTestRequest[T any](req TestRequest) *TestResponse[T] {
	log.Printf("Starting MustSendTestRequest")
	resp, err := SendTestRequest[T](req)
	if err != nil {
		log.Fatalf("Failed to send test request: %v", err)
	}
	log.Printf("Test request completed successfully")
	return resp
}
