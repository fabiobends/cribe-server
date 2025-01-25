package migrations

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"cribeapp.com/cribe-server/internal/utils"
)

func TestMigrations_GetIntegration(t *testing.T) {
	req, err := http.NewRequest("GET", "/migrations", nil)

	if err != nil {
		t.Fatalf("Could not create request: %v", err)
	}

	rec := httptest.NewRecorder()

	Handler(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, rec.Code)
	}

	result := utils.DecodeResponse[[]Migration](rec.Body.String())

	if len(result) == 0 {
		t.Errorf("Expected at least one migration, got %d", len(result))
	}

}

func TestMigrations_PostIntegration(t *testing.T) {
	req, err := http.NewRequest("POST", "/migrations", nil)

	if err != nil {
		t.Fatalf("Could not create request: %v", err)
	}

	rec := httptest.NewRecorder()

	Handler(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("Expected status code %d, got %d", http.StatusCreated, rec.Code)
	}

	result := utils.DecodeResponse[[]Migration](rec.Body.String())

	if len(result) != 0 {
		t.Errorf("Expected no migrations, got %d", len(result))
	}

}
