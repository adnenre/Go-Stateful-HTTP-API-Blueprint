package user

import (
	"net/http"
	"rest-api-blueprint/internal/features/user/dto"
)

var routeDTOMap = map[string]func() any{
	"PATCH /api/v1/users/me/preferences": func() any { return &dto.UpdatePreferencesRequest{} },
}

func Resolver(r *http.Request) (any, bool) {
	key := r.Method + " " + r.URL.Path
	if factory, ok := routeDTOMap[key]; ok {
		return factory(), true
	}
	return nil, false
}
