package unit

import (
	"fmt"
	"net/http"
	"testing"

	"rest-api-blueprint/internal/features/admin"

	"github.com/stretchr/testify/assert"
)

func TestAdminResolver(t *testing.T) {
	tests := []struct {
		name     string
		method   string
		path     string
		wantDTO  bool
		wantType string
	}{
		{
			name:     "create user",
			method:   "POST",
			path:     "/api/v1/admin/users",
			wantDTO:  true,
			wantType: "*dto.CreateUserRequest",
		},
		{
			name:     "update user (with ID)",
			method:   "PUT",
			path:     "/api/v1/admin/users/123",
			wantDTO:  true,
			wantType: "*dto.UpdateUserRequest",
		},
		{
			name:    "list users",
			method:  "GET",
			path:    "/api/v1/admin/users",
			wantDTO: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest(tt.method, tt.path, nil)
			got, ok := admin.Resolver(req)
			assert.Equal(t, tt.wantDTO, ok)
			if ok {
				assert.Equal(t, tt.wantType, fmt.Sprintf("%T", got))
			}
		})
	}
}
