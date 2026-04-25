package unit

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	auth "rest-api-blueprint/internal/auth"
	"rest-api-blueprint/internal/features/admin/controller"
	authModel "rest-api-blueprint/internal/features/auth/model"
	"rest-api-blueprint/internal/gen"
	"rest-api-blueprint/internal/middleware"

	"github.com/stretchr/testify/assert"
)

type mockAdminService struct {
	listUsersFunc  func(ctx context.Context, limit, offset int) ([]authModel.User, error)
	createUserFunc func(ctx context.Context, email, username, password, role string, avatar *string) (*authModel.User, error)
	getUserFunc    func(ctx context.Context, id string) (*authModel.User, error)
	updateUserFunc func(ctx context.Context, id string, email, username, role *string, password *string, avatar *string) error
	deleteUserFunc func(ctx context.Context, id string) error
}

func (m *mockAdminService) ListUsers(ctx context.Context, limit, offset int) ([]authModel.User, error) {
	return m.listUsersFunc(ctx, limit, offset)
}
func (m *mockAdminService) CreateUser(ctx context.Context, email, username, password, role string, avatar *string) (*authModel.User, error) {
	return m.createUserFunc(ctx, email, username, password, role, avatar)
}
func (m *mockAdminService) GetUser(ctx context.Context, id string) (*authModel.User, error) {
	return m.getUserFunc(ctx, id)
}
func (m *mockAdminService) UpdateUser(ctx context.Context, id string, email, username, role *string, password *string, avatar *string) error {
	return m.updateUserFunc(ctx, id, email, username, role, password, avatar)
}
func (m *mockAdminService) DeleteUser(ctx context.Context, id string) error {
	return m.deleteUserFunc(ctx, id)
}

func TestAdminController_ListUsers(t *testing.T) {
	tests := []struct {
		name           string
		claims         *auth.Claims
		params         gen.ListUsersParams
		mockList       func(ctx context.Context, limit, offset int) ([]authModel.User, error)
		expectedStatus int
	}{
		{
			name:   "success admin",
			claims: &auth.Claims{Role: "admin"},
			params: gen.ListUsersParams{Limit: intPtr(10), Offset: intPtr(0)},
			mockList: func(ctx context.Context, limit, offset int) ([]authModel.User, error) {
				return []authModel.User{{ID: "1", Username: "admin"}}, nil
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "forbidden user",
			claims:         &auth.Claims{Role: "user"},
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := &mockAdminService{listUsersFunc: tt.mockList}
			ctrl := controller.NewAdminController(mockSvc)

			req := httptest.NewRequest("GET", "/api/v1/admin/users", nil)
			ctx := context.Background()
			if tt.claims != nil {
				ctx = context.WithValue(ctx, middleware.UserKey, tt.claims)
			}
			ctx = context.WithValue(ctx, middleware.RequestIDKey, "test-id")
			req = req.WithContext(ctx)
			w := httptest.NewRecorder()

			ctrl.ListUsers(w, req, tt.params)
			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func intPtr(i int) *int { return &i }
