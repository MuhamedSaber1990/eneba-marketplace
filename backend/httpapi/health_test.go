package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// These tests run without a database: NewRouter(nil, Config{}) is the "DATABASE_URL is
// missing" case.

func TestHealthWithoutDatabase(t *testing.T) {
	router := NewRouter(nil, Config{})

	request := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusServiceUnavailable)
	}
	var body healthResponse
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("response is not JSON: %v", err)
	}
	if body.Status != "degraded" || body.Database != "not configured" {
		t.Fatalf("body = %+v", body)
	}
}

func TestUnknownRouteReturnsJSON404(t *testing.T) {
	router := NewRouter(nil, Config{})

	request := httptest.NewRequest(http.MethodGet, "/api/does-not-exist", nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusNotFound)
	}
	var body map[string]string
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("response is not JSON: %v", err)
	}
	if body["error"] == "" {
		t.Fatalf("body has no error message: %v", body)
	}
}
