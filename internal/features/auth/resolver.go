package auth

import (
	"net/http"
	"rest-api-blueprint/internal/features/auth/dto"
)

var routeDTOMap = map[string]func() any{
	"POST /api/v1/auth/register": func() any { return &dto.RegisterRequest{} },
	"POST /api/v1/auth/login":    func() any { return &dto.LoginRequest{} },
}

func Resolver(r *http.Request) (any, bool) {
	key := r.Method + " " + r.URL.Path
	if factory, ok := routeDTOMap[key]; ok {
		return factory(), true
	}
	return nil, false
}
