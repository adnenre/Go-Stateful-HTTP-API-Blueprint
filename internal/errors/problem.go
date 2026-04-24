package errors

import (
	"encoding/json"
	"net/http"
)

// ProblemDetails represents the RFC 7807 error response.
type ProblemDetails struct {
	Type     string            `json:"type"`
	Title    string            `json:"title"`
	Status   int               `json:"status"`
	Detail   string            `json:"detail"`
	Instance string            `json:"instance"`
	Errors   map[string]string `json:"errors,omitempty"`
}

// WriteProblem writes an RFC 7807 response from a DomainError.
// instance should be the request ID (e.g., from middleware.GetRequestID(ctx)).
func WriteProblem(w http.ResponseWriter, r *http.Request, err *DomainError, instance string) {
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(err.Status)

	if instance == "" {
		instance = "unknown"
	}

	problem := ProblemDetails{
		Type:     err.Type,
		Title:    err.Title,
		Status:   err.Status,
		Detail:   err.Detail,
		Instance: instance,
		Errors:   err.Details,
	}
	if encodeErr := json.NewEncoder(w).Encode(problem); encodeErr != nil {
		// Fallback – should not happen
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// WriteProblemSimple writes a generic RFC 7807 response for non‑domain errors.
func WriteProblemSimple(w http.ResponseWriter, r *http.Request, status int, title, detail, instance string) {
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(status)

	if instance == "" {
		instance = "unknown"
	}

	problem := ProblemDetails{
		Type:     "about:blank",
		Title:    title,
		Status:   status,
		Detail:   detail,
		Instance: instance,
	}
	if encodeErr := json.NewEncoder(w).Encode(problem); encodeErr != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}
