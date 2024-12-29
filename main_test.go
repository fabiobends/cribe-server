package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestStatusHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "/status", nil)
	
	if err != nil {
		t.Fatalf("Could not create request: %v", err)
	}

	rec := httptest.NewRecorder()
	statusHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, rec.Code)
	}

	expected := `{"message":"ok"}`
	result := strings.Split(rec.Body.String(), "\n")[0]

	if result != expected {
		t.Errorf("Expected body %s, got %s", expected, rec.Body.String())
	}

}

func TestNotFoundHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "/whatever", nil)

	if err != nil {
		t.Fatalf("Could not create request: %v", err)
	}

	rec := httptest.NewRecorder()
	notFoundHandler(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("Expected status code %d, got %d", http.StatusNotFound, rec.Code)
	}

	expected := `{"message":"not found"}`
	result := strings.Split(rec.Body.String(), "\n")[0]

	if result != expected {
		t.Errorf("Expected body %s, got %s", expected, rec.Body.String())
	}
}