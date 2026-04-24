package errors

import "fmt"

// DomainError represents an application error with RFC 7807 fields.
type DomainError struct {
	Type    string            // URI / error type (e.g., "validation", "not_found")
	Title   string            // short human‑readable summary
	Status  int               // HTTP status code
	Detail  string            // detailed explanation
	Details map[string]string // optional field‑specific errors (for validation)
}

func (e *DomainError) Error() string {
	return e.Detail
}

// ----------------------------------------------------------------------------
// 4xx Client Error Factories
// ----------------------------------------------------------------------------

// BadRequestError returns a 400 Bad Request error.
func BadRequestError(detail string) *DomainError {
	return &DomainError{
		Type:   "bad_request",
		Title:  "Bad request",
		Status: 400,
		Detail: detail,
	}
}

// UnauthorizedError returns a 401 Unauthorized error.
func UnauthorizedError(detail string) *DomainError {
	return &DomainError{
		Type:   "unauthorized",
		Title:  "Unauthorized",
		Status: 401,
		Detail: detail,
	}
}

// ForbiddenError returns a 403 Forbidden error.
func ForbiddenError(detail string) *DomainError {
	return &DomainError{
		Type:   "forbidden",
		Title:  "Forbidden",
		Status: 403,
		Detail: detail,
	}
}

// NotFoundError returns a 404 Not Found error for a specific resource.
func NotFoundError(resource string) *DomainError {
	return &DomainError{
		Type:   "not_found",
		Title:  "Resource not found",
		Status: 404,
		Detail: fmt.Sprintf("%s not found", resource),
	}
}

// NotFoundErrorCustom returns a 404 Not Found with custom detail.
func NotFoundErrorCustom(detail string) *DomainError {
	return &DomainError{
		Type:   "not_found",
		Title:  "Resource not found",
		Status: 404,
		Detail: detail,
	}
}

// ConflictError returns a 409 Conflict error.
func ConflictError(resource string) *DomainError {
	return &DomainError{
		Type:   "conflict",
		Title:  "Conflict",
		Status: 409,
		Detail: fmt.Sprintf("%s already exists", resource),
	}
}

// ConflictErrorCustom returns a 409 Conflict with custom detail.
func ConflictErrorCustom(detail string) *DomainError {
	return &DomainError{
		Type:   "conflict",
		Title:  "Conflict",
		Status: 409,
		Detail: detail,
	}
}

// UnprocessableEntityError returns a 422 Unprocessable Entity error.
// Use this for validation failures (e.g., missing required fields, invalid format).
// You can pass a map of field‑specific errors as the third argument.
func UnprocessableEntityError(detail string, fieldErrors map[string]string) *DomainError {
	return &DomainError{
		Type:    "validation",
		Title:   "Validation failed",
		Status:  422,
		Detail:  detail,
		Details: fieldErrors,
	}
}

// TooManyRequestsError returns a 429 Too Many Requests error.
// Note: rate limiting middleware typically handles this, but you may also return it from services.
func TooManyRequestsError(detail string) *DomainError {
	return &DomainError{
		Type:   "too_many_requests",
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
		Type:   "internal",
		Title:  "Internal server error",
		Status: 500,
		Detail: detail,
	}
}

// ServiceUnavailableError returns a 503 Service Unavailable.
func ServiceUnavailableError(detail string) *DomainError {
	return &DomainError{
		Type:   "service_unavailable",
		Title:  "Service unavailable",
		Status: 503,
		Detail: detail,
	}
}
