package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"rest-api-blueprint/internal/errors"
)

func TestErrorIntegration(t *testing.T) {
	// Test handler that returns a 400 Bad Request with domain error
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := errors.BadRequestError("test error detail")
		errors.WriteProblem(w, r, err, "integration-test-id")
	})

	server := httptest.NewServer(handler)
	defer server.Close()

	resp, err := http.Get(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", resp.StatusCode)
	}
	if ct := resp.Header.Get("Content-Type"); ct != "application/problem+json" {
		t.Errorf("expected Content-Type application/problem+json, got %s", ct)
	}

	var problem errors.ProblemDetails
	if err := json.NewDecoder(resp.Body).Decode(&problem); err != nil {
		t.Fatal(err)
	}

	if problem.Type != "bad_request" {
		t.Errorf("Type: got %s, want bad_request", problem.Type)
	}
	if problem.Title != "Bad request" {
		t.Errorf("Title: got %s, want Bad request", problem.Title)
	}
	if problem.Status != http.StatusBadRequest {
		t.Errorf("Status: got %d, want %d", problem.Status, http.StatusBadRequest)
	}
	if problem.Detail != "test error detail" {
		t.Errorf("Detail: got %s, want test error detail", problem.Detail)
	}
	if problem.Instance != "integration-test-id" {
		t.Errorf("Instance: got %s, want integration-test-id", problem.Instance)
	}
}

func TestWriteProblemSimpleIntegration(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		errors.WriteProblemSimple(w, r, http.StatusTooManyRequests, "Rate Limited", "Slow down", "ratelimit-123")
	})

	server := httptest.NewServer(handler)
	defer server.Close()

	resp, err := http.Get(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusTooManyRequests {
		t.Errorf("expected status 429, got %d", resp.StatusCode)
	}

	var problem errors.ProblemDetails
	if err := json.NewDecoder(resp.Body).Decode(&problem); err != nil {
		t.Fatal(err)
	}

	if problem.Type != "about:blank" {
		t.Errorf("Type: got %s, want about:blank", problem.Type)
	}
	if problem.Title != "Rate Limited" {
		t.Errorf("Title: got %s, want Rate Limited", problem.Title)
	}
	if problem.Status != http.StatusTooManyRequests {
		t.Errorf("Status: got %d, want %d", problem.Status, http.StatusTooManyRequests)
	}
	if problem.Detail != "Slow down" {
		t.Errorf("Detail: got %s, want Slow down", problem.Detail)
	}
	if problem.Instance != "ratelimit-123" {
		t.Errorf("Instance: got %s, want ratelimit-123", problem.Instance)
	}
}
