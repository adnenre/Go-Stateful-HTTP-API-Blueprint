package unit

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"rest-api-blueprint/internal/config"
	"rest-api-blueprint/internal/errors"
	"rest-api-blueprint/internal/features/auth/controller"
	"rest-api-blueprint/internal/features/auth/dto"
)

// Updated mock service to match the current service.Service interface.
type mockAuthService struct {
	registerFunc             func(ctx context.Context, email, username, password string, avatar *string) error
	loginFunc                func(ctx context.Context, email, password string) (*dto.LoginResponse, error)
	verifyOTPFunc            func(ctx context.Context, email, otp string) (*dto.OTPResponse, error)
	refreshAccessTokenFunc   func(ctx context.Context, refreshToken string) (*dto.RefreshResponse, error)
	revokeAllUserTokensFunc  func(ctx context.Context, userID string) error
	getSessionFunc           func(ctx context.Context, userID string) (*dto.UserResponse, error)
	requestPasswordResetFunc func(ctx context.Context, email string) error
	confirmPasswordResetFunc func(ctx context.Context, token, newPassword string) error
}

func (m *mockAuthService) Register(ctx context.Context, email, username, password string, avatar *string) error {
	return m.registerFunc(ctx, email, username, password, avatar)
}
func (m *mockAuthService) Login(ctx context.Context, email, password string) (*dto.LoginResponse, error) {
	return m.loginFunc(ctx, email, password)
}
func (m *mockAuthService) VerifyOtp(ctx context.Context, email, otp string) (*dto.OTPResponse, error) {
	return m.verifyOTPFunc(ctx, email, otp)
}
func (m *mockAuthService) RefreshAccessToken(ctx context.Context, refreshToken string) (*dto.RefreshResponse, error) {
	return m.refreshAccessTokenFunc(ctx, refreshToken)
}
func (m *mockAuthService) RevokeAllUserTokens(ctx context.Context, userID string) error {
	return m.revokeAllUserTokensFunc(ctx, userID)
}
func (m *mockAuthService) GetSession(ctx context.Context, userID string) (*dto.UserResponse, error) {
	return m.getSessionFunc(ctx, userID)
}
func (m *mockAuthService) RequestPasswordReset(ctx context.Context, email string) error {
	return m.requestPasswordResetFunc(ctx, email)
}
func (m *mockAuthService) ConfirmPasswordReset(ctx context.Context, token, newPassword string) error {
	return m.confirmPasswordResetFunc(ctx, token, newPassword)
}

func TestAuthController_Register(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		mockRegister   func(ctx context.Context, email, username, password string, avatar *string) error
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
			mockRegister: func(ctx context.Context, email, username, password string, avatar *string) error {
				return nil
			},
			expectedStatus: http.StatusAccepted,
		},
		{
			name: "email already exists",
			requestBody: dto.RegisterRequest{
				Email:    "existing@example.com",
				Password: "pass",
				Username: "newuser",
			},
			mockRegister: func(ctx context.Context, email, username, password string, avatar *string) error {
				return errors.ConflictError("email")
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
			mockSvc := &mockAuthService{
				registerFunc:             tt.mockRegister,
				loginFunc:                func(ctx context.Context, email, password string) (*dto.LoginResponse, error) { return nil, nil },
				verifyOTPFunc:            func(ctx context.Context, email, otp string) (*dto.OTPResponse, error) { return nil, nil },
				refreshAccessTokenFunc:   func(ctx context.Context, refreshToken string) (*dto.RefreshResponse, error) { return nil, nil },
				revokeAllUserTokensFunc:  func(ctx context.Context, userID string) error { return nil },
				getSessionFunc:           func(ctx context.Context, userID string) (*dto.UserResponse, error) { return nil, nil },
				requestPasswordResetFunc: func(ctx context.Context, email string) error { return nil },
				confirmPasswordResetFunc: func(ctx context.Context, token, newPassword string) error { return nil },
			}
			cfg := &config.Config{
				JWTSecret:          "test",
				JWTExpiry:          15 * time.Minute,
				RefreshTokenExpiry: 168 * time.Hour,
			}
			ctrl := controller.NewAuthController(mockSvc, cfg)

			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest("POST", "/api/v1/auth/register", bytes.NewReader(body))
			req.Header.Set("X-Request-ID", "test-id")
			w := httptest.NewRecorder()

			ctrl.Register(w, req)

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
		mockLogin      func(ctx context.Context, email, password string) (*dto.LoginResponse, error)
		expectedStatus int
	}{
		{
			name:        "success",
			requestBody: dto.LoginRequest{Email: "test@example.com", Password: "correct"},
			mockLogin: func(ctx context.Context, email, password string) (*dto.LoginResponse, error) {
				return &dto.LoginResponse{
					AccessToken:  "jwt-token",
					RefreshToken: "refresh-token",
					ExpiresIn:    3600,
					TokenType:    "Bearer",
					User: dto.UserResponse{
						ID:    "123",
						Email: "test@example.com",
					},
				}, nil
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "invalid credentials",
			requestBody: dto.LoginRequest{Email: "wrong", Password: "wrong"},
			mockLogin: func(ctx context.Context, email, password string) (*dto.LoginResponse, error) {
				return nil, errors.UnauthorizedError("invalid credentials")
			},
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := &mockAuthService{
				loginFunc:                tt.mockLogin,
				registerFunc:             func(ctx context.Context, email, username, password string, avatar *string) error { return nil },
				verifyOTPFunc:            func(ctx context.Context, email, otp string) (*dto.OTPResponse, error) { return nil, nil },
				refreshAccessTokenFunc:   func(ctx context.Context, refreshToken string) (*dto.RefreshResponse, error) { return nil, nil },
				revokeAllUserTokensFunc:  func(ctx context.Context, userID string) error { return nil },
				getSessionFunc:           func(ctx context.Context, userID string) (*dto.UserResponse, error) { return nil, nil },
				requestPasswordResetFunc: func(ctx context.Context, email string) error { return nil },
				confirmPasswordResetFunc: func(ctx context.Context, token, newPassword string) error { return nil },
			}
			cfg := &config.Config{
				JWTSecret:          "test",
				JWTExpiry:          15 * time.Minute,
				RefreshTokenExpiry: 168 * time.Hour,
			}
			ctrl := controller.NewAuthController(mockSvc, cfg)
			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewReader(body))
			req.Header.Set("X-Request-ID", "test-id")
			w := httptest.NewRecorder()

			ctrl.Login(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedStatus == http.StatusOK {
				cookies := w.Result().Cookies()
				assert.NotNil(t, cookies)
				var foundAccess, foundRefresh bool
				for _, c := range cookies {
					if c.Name == "access_token" {
						foundAccess = true
					}
					if c.Name == "refresh_token" {
						foundRefresh = true
					}
				}
				assert.True(t, foundAccess, "access_token cookie should be set")
				assert.True(t, foundRefresh, "refresh_token cookie should be set")
			}
		})
	}
}
