package status

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"cribeapp.com/cribe-server/internal/utils"
)

func TestStatusIntegration(t *testing.T) {
	req, err := http.NewRequest("GET", "/status", nil)

	if err != nil {
		t.Fatalf("Could not create request: %v", err)
	}

	rec := httptest.NewRecorder()

	Handler(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, rec.Code)
	}

	expected := `{"message":"ok"}`
	result := utils.SanitizeJSONString(rec.Body.String())

	if result != expected {
		t.Errorf("Expected body %s, got %s", expected, rec.Body.String())
	}

}
