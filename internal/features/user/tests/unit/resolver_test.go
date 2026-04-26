package unit

import (
	"fmt"
	"net/http"
	"testing"

	"rest-api-blueprint/internal/features/user"

	"github.com/stretchr/testify/assert"
)

func TestUserResolver(t *testing.T) {
	tests := []struct {
		name     string
		method   string
		path     string
		wantDTO  bool
		wantType string
	}{
		{
			name:     "update preferences",
			method:   "PATCH",
			path:     "/api/v1/users/me/preferences",
			wantDTO:  true,
			wantType: "*dto.UpdatePreferencesRequest",
		},
		{
			name:    "unknown path",
			method:  "GET",
			path:    "/api/v1/users/me",
			wantDTO: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest(tt.method, tt.path, nil)
			got, ok := user.Resolver(req)
			assert.Equal(t, tt.wantDTO, ok)
			if ok {
				assert.Equal(t, tt.wantType, fmt.Sprintf("%T", got))
			}
		})
	}
}
