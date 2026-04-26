package errors

import (
	"fmt"
	"strings"
)

// baseURL holds the base URL for error documentation (e.g., "https://api.example.com").
// If empty, relative paths (e.g., "/errors/validation.html") are used.
var baseURL string

// Init initializes the error package with the base URL for error documentation.
// Call this once in main after loading configuration.
func Init(url string) {
	// Trim trailing slash to avoid double slashes when building the type URI.
	baseURL = strings.TrimSuffix(url, "/")
}

// DomainError represents an application error with RFC 7807 fields.
type DomainError struct {
	Type    string            // URI / error type (e.g., "/errors/validation.html")
	Title   string            // short human‑readable summary
	Status  int               // HTTP status code
	Detail  string            // detailed explanation
	Details map[string]string // optional field‑specific errors (for validation)
}

func (e *DomainError) Error() string {
	return e.Detail
}

// buildType constructs the full URI for an error path.
func buildType(path string) string {
	if baseURL == "" {
		return path
	}
	return baseURL + path
}

// ----------------------------------------------------------------------------
// 4xx Client Error Factories
// ----------------------------------------------------------------------------

// BadRequestError returns a 400 Bad Request error.
func BadRequestError(detail string) *DomainError {
	return &DomainError{
		Type:   buildType("/errors/bad_request.html"),
		Title:  "Bad request",
		Status: 400,
		Detail: detail,
	}
}

// UnauthorizedError returns a 401 Unauthorized error.
func UnauthorizedError(detail string) *DomainError {
	return &DomainError{
		Type:   buildType("/errors/unauthorized.html"),
		Title:  "Unauthorized",
		Status: 401,
		Detail: detail,
	}
}

// ForbiddenError returns a 403 Forbidden error.
func ForbiddenError(detail string) *DomainError {
	return &DomainError{
		Type:   buildType("/errors/forbidden.html"),
		Title:  "Forbidden",
		Status: 403,
		Detail: detail,
	}
}

// NotFoundError returns a 404 Not Found error for a specific resource.
func NotFoundError(resource string) *DomainError {
	return &DomainError{
		Type:   buildType("/errors/not_found.html"),
		Title:  "Resource not found",
		Status: 404,
		Detail: fmt.Sprintf("%s not found", resource),
	}
}

// NotFoundErrorCustom returns a 404 Not Found with custom detail.
func NotFoundErrorCustom(detail string) *DomainError {
	return &DomainError{
		Type:   buildType("/errors/not_found.html"),
		Title:  "Resource not found",
		Status: 404,
		Detail: detail,
	}
}

// ConflictError returns a 409 Conflict error.
func ConflictError(resource string) *DomainError {
	return &DomainError{
		Type:   buildType("/errors/conflict.html"),
		Title:  "Conflict",
		Status: 409,
		Detail: fmt.Sprintf("%s already exists", resource),
	}
}

// ConflictErrorCustom returns a 409 Conflict with custom detail.
func ConflictErrorCustom(detail string) *DomainError {
	return &DomainError{
		Type:   buildType("/errors/conflict.html"),
		Title:  "Conflict",
		Status: 409,
		Detail: detail,
	}
}

// UnprocessableEntityError returns a 422 Unprocessable Entity error.
func UnprocessableEntityError(detail string, fieldErrors map[string]string) *DomainError {
	return &DomainError{
		Type:    buildType("/errors/validation.html"),
		Title:   "Validation failed",
		Status:  422,
		Detail:  detail,
		Details: fieldErrors,
	}
}

// TooManyRequestsError returns a 429 Too Many Requests error.
func TooManyRequestsError(detail string) *DomainError {
	return &DomainError{
		Type:   buildType("/errors/too_many_requests.html"),
		Title:  "Too many requests",
		Status: 429,
		Detail: detail,
	}
}

// ----------------------------------------------------------------------------
// 5xx Server Error Factories
// ----------------------------------------------------------------------------

// InternalError returns a 500 Internal Server Error.
func InternalError(detail string) *DomainError {
	return &DomainError{
		Type:   buildType("/errors/internal.html"),
		Title:  "Internal server error",
		Status: 500,
		Detail: detail,
	}
}

// ServiceUnavailableError returns a 503 Service Unavailable.
func ServiceUnavailableError(detail string) *DomainError {
	return &DomainError{
		Type:   buildType("/errors/service_unavailable.html"),
		Title:  "Service unavailable",
		Status: 503,
		Detail: detail,
	}
}
