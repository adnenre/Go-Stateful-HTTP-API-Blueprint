package unit

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"rest-api-blueprint/internal/auth"
	authModel "rest-api-blueprint/internal/features/auth/model"
	"rest-api-blueprint/internal/features/user/controller"
	"rest-api-blueprint/internal/features/user/dto"
	userModel "rest-api-blueprint/internal/features/user/model"
	"rest-api-blueprint/internal/middleware"

	"github.com/stretchr/testify/assert"
)

type mockUserService struct {
	getProfileFunc        func(ctx context.Context, userID string) (*authModel.User, error)
	updatePreferencesFunc func(ctx context.Context, userID string, notifications *bool, language *string) error
	getPreferencesFunc    func(ctx context.Context, userID string) (*userModel.UserPreferences, error)
}

func (m *mockUserService) GetProfile(ctx context.Context, userID string) (*authModel.User, error) {
	return m.getProfileFunc(ctx, userID)
}
func (m *mockUserService) UpdatePreferences(ctx context.Context, userID string, notifications *bool, language *string) error {
	return m.updatePreferencesFunc(ctx, userID, notifications, language)
}
func (m *mockUserService) GetPreferences(ctx context.Context, userID string) (*userModel.UserPreferences, error) {
	return m.getPreferencesFunc(ctx, userID)
}

func TestUserController_GetMe(t *testing.T) {
	tests := []struct {
		name           string
		claims         *auth.Claims
		mockGetProfile func(ctx context.Context, userID string) (*authModel.User, error)
		expectedStatus int
	}{
		{
			name:   "success",
			claims: &auth.Claims{UserID: "user-123", Role: "user"},
			mockGetProfile: func(ctx context.Context, userID string) (*authModel.User, error) {
				return &authModel.User{ID: "user-123", Username: "test", Email: "test@example.com"}, nil
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "no claims",
			claims:         nil,
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := &mockUserService{getProfileFunc: tt.mockGetProfile}
			ctrl := controller.NewUserController(mockSvc)

			req := httptest.NewRequest("GET", "/api/v1/users/me", nil)
			ctx := context.Background()
			if tt.claims != nil {
				// Use exported UserKey
				ctx = context.WithValue(ctx, middleware.UserKey, tt.claims)
			}
			// Use exported RequestIDKey (though not strictly needed for test)
			ctx = context.WithValue(ctx, middleware.RequestIDKey, "test-id")
			req = req.WithContext(ctx)
			w := httptest.NewRecorder()

			ctrl.GetMe(w, req)
			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestUserController_UpdatePreferences(t *testing.T) {
	tests := []struct {
		name           string
		claims         *auth.Claims
		requestBody    dto.UpdatePreferencesRequest
		mockUpdate     func(ctx context.Context, userID string, notifications *bool, language *string) error
		mockGetPrefs   func(ctx context.Context, userID string) (*userModel.UserPreferences, error)
		expectedStatus int
	}{
		{
			name:   "success",
			claims: &auth.Claims{UserID: "user-123"},
			requestBody: dto.UpdatePreferencesRequest{
				Notifications: boolPtr(true),
				Language:      stringPtr("fr"),
			},
			mockUpdate: func(ctx context.Context, userID string, notifications *bool, language *string) error { return nil },
			mockGetPrefs: func(ctx context.Context, userID string) (*userModel.UserPreferences, error) {
				return &userModel.UserPreferences{UserID: "user-123", Notifications: boolPtr(true), Language: stringPtr("fr")}, nil
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "no claims",
			claims:         nil,
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := &mockUserService{
				updatePreferencesFunc: tt.mockUpdate,
				getPreferencesFunc:    tt.mockGetPrefs,
			}
			ctrl := controller.NewUserController(mockSvc)

			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest("PATCH", "/api/v1/users/me/preferences", bytes.NewReader(body))
			ctx := context.Background()
			if tt.claims != nil {
				ctx = context.WithValue(ctx, middleware.UserKey, tt.claims)
			}
			ctx = context.WithValue(ctx, middleware.RequestIDKey, "test-id")
			req = req.WithContext(ctx)
			w := httptest.NewRecorder()

			ctrl.UpdatePreferences(w, req)
			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func boolPtr(b bool) *bool       { return &b }
func stringPtr(s string) *string { return &s }
