package unit

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"rest-api-blueprint/internal/errors"
)

func TestDomainErrorFactories(t *testing.T) {
	tests := []struct {
		name     string
		factory  func() *errors.DomainError
		expected *errors.DomainError
	}{
		{
			name:     "BadRequestError",
			factory:  func() *errors.DomainError { return errors.BadRequestError("bad input") },
			expected: &errors.DomainError{Type: "bad_request", Title: "Bad request", Status: 400, Detail: "bad input"},
		},
		{
			name:     "UnauthorizedError",
			factory:  func() *errors.DomainError { return errors.UnauthorizedError("missing token") },
			expected: &errors.DomainError{Type: "unauthorized", Title: "Unauthorized", Status: 401, Detail: "missing token"},
		},
		{
			name:     "ForbiddenError",
			factory:  func() *errors.DomainError { return errors.ForbiddenError("insufficient role") },
			expected: &errors.DomainError{Type: "forbidden", Title: "Forbidden", Status: 403, Detail: "insufficient role"},
		},
		{
			name:     "NotFoundError",
			factory:  func() *errors.DomainError { return errors.NotFoundError("user") },
			expected: &errors.DomainError{Type: "not_found", Title: "Resource not found", Status: 404, Detail: "user not found"},
		},
		{
			name:     "NotFoundErrorCustom",
			factory:  func() *errors.DomainError { return errors.NotFoundErrorCustom("custom not found") },
			expected: &errors.DomainError{Type: "not_found", Title: "Resource not found", Status: 404, Detail: "custom not found"},
		},
		{
			name:     "ConflictError",
			factory:  func() *errors.DomainError { return errors.ConflictError("email") },
			expected: &errors.DomainError{Type: "conflict", Title: "Conflict", Status: 409, Detail: "email already exists"},
		},
		{
			name:     "ConflictErrorCustom",
			factory:  func() *errors.DomainError { return errors.ConflictErrorCustom("custom conflict") },
			expected: &errors.DomainError{Type: "conflict", Title: "Conflict", Status: 409, Detail: "custom conflict"},
		},
		{
			name:     "UnprocessableEntityError no fields",
			factory:  func() *errors.DomainError { return errors.UnprocessableEntityError("invalid data", nil) },
			expected: &errors.DomainError{Type: "validation", Title: "Validation failed", Status: 422, Detail: "invalid data", Details: nil},
		},
		{
			name: "UnprocessableEntityError with fields",
			factory: func() *errors.DomainError {
				return errors.UnprocessableEntityError("invalid data", map[string]string{"email": "required"})
			},
			expected: &errors.DomainError{Type: "validation", Title: "Validation failed", Status: 422, Detail: "invalid data", Details: map[string]string{"email": "required"}},
		},
		{
			name:     "TooManyRequestsError",
			factory:  func() *errors.DomainError { return errors.TooManyRequestsError("slow down") },
			expected: &errors.DomainError{Type: "too_many_requests", Title: "Too many requests", Status: 429, Detail: "slow down"},
		},
		{
			name:     "InternalError",
			factory:  func() *errors.DomainError { return errors.InternalError("DB error") },
			expected: &errors.DomainError{Type: "internal", Title: "Internal server error", Status: 500, Detail: "DB error"},
		},
		{
			name:     "ServiceUnavailableError",
			factory:  func() *errors.DomainError { return errors.ServiceUnavailableError("redis down") },
			expected: &errors.DomainError{Type: "service_unavailable", Title: "Service unavailable", Status: 503, Detail: "redis down"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.factory()
			if err.Type != tt.expected.Type {
				t.Errorf("Type: got %s, want %s", err.Type, tt.expected.Type)
			}
			if err.Title != tt.expected.Title {
				t.Errorf("Title: got %s, want %s", err.Title, tt.expected.Title)
			}
			if err.Status != tt.expected.Status {
				t.Errorf("Status: got %d, want %d", err.Status, tt.expected.Status)
			}
			if err.Detail != tt.expected.Detail {
				t.Errorf("Detail: got %s, want %s", err.Detail, tt.expected.Detail)
			}
			if len(err.Details) != len(tt.expected.Details) {
				t.Errorf("Details length: got %d, want %d", len(err.Details), len(tt.expected.Details))
			}
			for k, v := range tt.expected.Details {
				if err.Details[k] != v {
					t.Errorf("Details[%s]: got %s, want %s", k, err.Details[k], v)
				}
			}
		})
	}
}

func TestWriteProblemSimpleJSON(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	errors.WriteProblemSimple(rec, req, http.StatusNotFound, "Not Found", "The resource does not exist", "test-instance-123")

	var problem errors.ProblemDetails
	if err := json.NewDecoder(rec.Body).Decode(&problem); err != nil {
		t.Fatal(err)
	}

	if problem.Type != "about:blank" {
		t.Errorf("Type: got %s, want about:blank", problem.Type)
	}
	if problem.Title != "Not Found" {
		t.Errorf("Title: got %s, want Not Found", problem.Title)
	}
	if problem.Status != http.StatusNotFound {
		t.Errorf("Status: got %d, want %d", problem.Status, http.StatusNotFound)
	}
	if problem.Detail != "The resource does not exist" {
		t.Errorf("Detail: got %s, want The resource does not exist", problem.Detail)
	}
	if problem.Instance != "test-instance-123" {
		t.Errorf("Instance: got %s, want test-instance-123", problem.Instance)
	}
}

func TestWriteProblemWithDomainError(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	err := errors.BadRequestError("test detail")
	errors.WriteProblem(rec, req, err, "req-123")

	var problem errors.ProblemDetails
	if err := json.NewDecoder(rec.Body).Decode(&problem); err != nil {
		t.Fatal(err)
	}

	if problem.Type != "bad_request" {
		t.Errorf("Type: got %s, want bad_request", problem.Type)
	}
	if problem.Title != "Bad request" {
		t.Errorf("Title: got %s, want Bad request", problem.Title)
	}
	if problem.Status != 400 {
		t.Errorf("Status: got %d, want 400", problem.Status)
	}
	if problem.Detail != "test detail" {
		t.Errorf("Detail: got %s, want test detail", problem.Detail)
	}
	if problem.Instance != "req-123" {
		t.Errorf("Instance: got %s, want req-123", problem.Instance)
	}
}
