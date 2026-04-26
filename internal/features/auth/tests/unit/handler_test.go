package unit

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"rest-api-blueprint/internal/errors"
	"rest-api-blueprint/internal/features/auth/controller"
	"rest-api-blueprint/internal/features/auth/dto"
)

type mockAuthService struct {
	registerFunc func(ctx context.Context, email, username, password string, avatar *string) (string, error)
	loginFunc    func(ctx context.Context, email, password string) (string, error)
}

func (m *mockAuthService) Register(ctx context.Context, email, username, password string, avatar *string) (string, error) {
	return m.registerFunc(ctx, email, username, password, avatar)
}
func (m *mockAuthService) Login(ctx context.Context, email, password string) (string, error) {
	return m.loginFunc(ctx, email, password)
}

func TestAuthController_Register(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		mockRegister   func(ctx context.Context, email, username, password string, avatar *string) (string, error)
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name: "success",
			requestBody: dto.RegisterRequest{
				Email:    "test@example.com",
				Password: "password123",
				Username: "testuser",
			},
			mockRegister: func(ctx context.Context, email, username, password string, avatar *string) (string, error) {
				return "user-123", nil
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "email already exists",
			requestBody: dto.RegisterRequest{
				Email:    "existing@example.com",
				Password: "pass",
				Username: "newuser",
			},
			mockRegister: func(ctx context.Context, email, username, password string, avatar *string) (string, error) {
				return "", errors.ConflictError("email") // changed
			},
			expectedStatus: http.StatusConflict,
			expectedBody: map[string]interface{}{
				"title":  "Conflict",
				"status": float64(409),
				"detail": "email already exists",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// ARRANGE
			mockSvc := &mockAuthService{registerFunc: tt.mockRegister}
			ctrl := controller.NewAuthController(mockSvc)

			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest("POST", "/api/v1/auth/register", bytes.NewReader(body))
			req.Header.Set("X-Request-ID", "test-id") // fallback for GetRequestID
			w := httptest.NewRecorder()

			// ACT
			ctrl.Register(w, req)

			// ASSERT
			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedBody != nil {
				var resp map[string]interface{}
				json.NewDecoder(w.Body).Decode(&resp)
				for k, v := range tt.expectedBody {
					assert.Equal(t, v, resp[k])
				}
			}
		})
	}
}

func TestAuthController_Login(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    dto.LoginRequest
		mockLogin      func(ctx context.Context, email, password string) (string, error)
		expectedStatus int
		expectedToken  string
	}{
		{
			name:           "success",
			requestBody:    dto.LoginRequest{Email: "test@example.com", Password: "correct"},
			mockLogin:      func(ctx context.Context, email, password string) (string, error) { return "jwt-token", nil },
			expectedStatus: http.StatusOK,
			expectedToken:  "jwt-token",
		},
		{
			name:        "invalid credentials",
			requestBody: dto.LoginRequest{Email: "wrong", Password: "wrong"},
			mockLogin: func(ctx context.Context, email, password string) (string, error) {
				return "", errors.UnauthorizedError("invalid credentials")
			}, // changed
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// ARRANGE
			mockSvc := &mockAuthService{loginFunc: tt.mockLogin}
			ctrl := controller.NewAuthController(mockSvc)
			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewReader(body))
			req.Header.Set("X-Request-ID", "test-id")
			w := httptest.NewRecorder()

			// ACT
			ctrl.Login(w, req)

			// ASSERT
			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedToken != "" {
				var resp dto.LoginResponse
				json.NewDecoder(w.Body).Decode(&resp)
				assert.Equal(t, tt.expectedToken, resp.Token)
			}
		})
	}
}
