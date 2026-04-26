package admin

import (
	"net/http"
	"rest-api-blueprint/internal/features/admin/dto"
	"strings"
)

var exactRoutes = map[string]func() any{
	"POST /api/v1/admin/users": func() any { return &dto.CreateUserRequest{} },
}

func Resolver(r *http.Request) (any, bool) {
	key := r.Method + " " + r.URL.Path
	// Exact match first
	if factory, ok := exactRoutes[key]; ok {
		return factory(), true
	}
	// For update (PUT) with ID in path
	if r.Method == http.MethodPut && strings.HasPrefix(r.URL.Path, "/api/v1/admin/users/") {
		return &dto.UpdateUserRequest{}, true
	}
	// DELETE usually has no body, so no DTO needed
	return nil, false
}
