package unit

import (
	"fmt"
	"net/http"
	"testing"

	"rest-api-blueprint/internal/features/auth"

	"github.com/stretchr/testify/assert"
)

func TestAuthResolver(t *testing.T) {
	tests := []struct {
		name     string
		method   string
		path     string
		wantDTO  bool
		wantType string
	}{
		{
			name:     "register route",
			method:   "POST",
			path:     "/api/v1/auth/register",
			wantDTO:  true,
			wantType: "*dto.RegisterRequest",
		},
		{
			name:     "login route",
			method:   "POST",
			path:     "/api/v1/auth/login",
			wantDTO:  true,
			wantType: "*dto.LoginRequest",
		},
		{
			name:    "unknown route",
			method:  "GET",
			path:    "/api/v1/auth/unknown",
			wantDTO: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest(tt.method, tt.path, nil)
			got, ok := auth.Resolver(req)
			assert.Equal(t, tt.wantDTO, ok)
			if ok {
				assert.Equal(t, tt.wantType, fmt.Sprintf("%T", got))
			}
		})
	}
}
