// internal/app/dto_resolver.go
package app

import (
	"net/http"
	"rest-api-blueprint/internal/features/admin"
	"rest-api-blueprint/internal/features/auth"
	"rest-api-blueprint/internal/features/user"
)

// DTOResolver returns the request DTO for validation, if any.
// It checks each feature's resolver in order (auth → user → admin).
func DTOResolver(r *http.Request) (any, bool) {
	if dto, ok := auth.Resolver(r); ok {
		return dto, true
	}
	if dto, ok := user.Resolver(r); ok {
		return dto, true
	}
	if dto, ok := admin.Resolver(r); ok {
		return dto, true
	}
	return nil, false
}
