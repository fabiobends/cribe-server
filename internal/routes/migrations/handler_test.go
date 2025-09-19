package migrations

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"cribeapp.com/cribe-server/internal/utils"
)

var handler = NewMockMigrationHandlerReady()

func TestMigrationsHandler_GetMigrations(t *testing.T) {
	expected := []Migration{
		{
			ID:        2,
			Name:      "000002_second",
			CreatedAt: utils.MockGetCurrentTime(),
		},
		{
			ID:        1,
			Name:      "000001_initial",
			CreatedAt: utils.MockGetCurrentTime(),
		},
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/migrations", nil)

	handler.HandleRequest(w, r)

	result, err := utils.DecodeResponse[[]Migration](w.Body.String())
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if fmt.Sprint(result) != fmt.Sprint(expected) {
		t.Errorf("expected %q, got %q", expected, result)
	}

}

func TestMigrationsHandler_PostMigrations(t *testing.T) {
	// First run
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/migrations", nil)

	handler.HandleRequest(w, r)

	firstResult, err := utils.DecodeResponse[[]Migration](w.Body.String())
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	firstExpected := []Migration{
		{
			ID:        2,
			Name:      "000002_second",
			CreatedAt: utils.MockGetCurrentTime(),
		},
		{
			ID:        1,
			Name:      "000001_initial",
			CreatedAt: utils.MockGetCurrentTime(),
		},
	}

	if fmt.Sprint(firstResult) != fmt.Sprint(firstExpected) {
		t.Errorf("expected %q, got %q", firstExpected, firstResult)
	}

	// Second run
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodPost, "/migrations", nil)

	handler.HandleRequest(w, r)

	secondResult, err := utils.DecodeResponse[[]Migration](w.Body.String())
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	secondExpected := []Migration{}

	if fmt.Sprint(secondResult) != fmt.Sprint(secondExpected) {
		t.Errorf("expected %q, got %q", secondExpected, secondResult)
	}
}
