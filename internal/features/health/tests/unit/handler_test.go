package unit

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"rest-api-blueprint/internal/features/health/controller"
	"rest-api-blueprint/internal/features/health/dto"
	"rest-api-blueprint/internal/gen"
	"testing"
)

type mockService struct {
	getHealthFunc func(ctx context.Context) (*dto.HealthData, error)
}

func (m *mockService) GetHealth(ctx context.Context) (*dto.HealthData, error) {
	return m.getHealthFunc(ctx)
}

func TestHealthController_GetHealth(t *testing.T) {
	tests := []struct {
		name           string
		mockReturn     *dto.HealthData
		mockErr        error
		wantStatus     int
		wantRespStatus gen.HealthResponseStatus
		wantDataStatus gen.HealthResponseDataStatus
	}{
		{
			name: "success_healthy_with_both_checks",
			mockReturn: &dto.HealthData{
				Status:    "healthy",
				Timestamp: "2026-04-22T12:00:00Z",
				Uptime:    "1s",
				Version:   "dev",
				Checks:    map[string]string{"database": "ok", "redis": "ok"},
			},
			mockErr:        nil,
			wantStatus:     http.StatusOK,
			wantRespStatus: gen.Success,
			wantDataStatus: gen.Healthy,
		},
		{
			name:           "service_error",
			mockReturn:     nil,
			mockErr:        fmt.Errorf("something went wrong"),
			wantStatus:     http.StatusInternalServerError,
			wantRespStatus: "",
			wantDataStatus: "",
		},
		{
			name: "unhealthy_database",
			mockReturn: &dto.HealthData{
				Status:    "unhealthy",
				Timestamp: "2026-04-22T12:00:00Z",
				Uptime:    "1s",
				Version:   "dev",
				Checks:    map[string]string{"database": "connection refused", "redis": "ok"},
			},
			mockErr:        nil,
			wantStatus:     http.StatusServiceUnavailable,
			wantRespStatus: gen.Success,
			wantDataStatus: gen.Unhealthy,
		},
		{
			name: "unhealthy_redis",
			mockReturn: &dto.HealthData{
				Status:    "unhealthy",
				Timestamp: "2026-04-22T12:00:00Z",
				Uptime:    "1s",
				Version:   "dev",
				Checks:    map[string]string{"database": "ok", "redis": "connection refused"},
			},
			mockErr:        nil,
			wantStatus:     http.StatusServiceUnavailable,
			wantRespStatus: gen.Success,
			wantDataStatus: gen.Unhealthy,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// ============================================================
			// ARRANGE
			// ============================================================
			mockSvc := &mockService{
				getHealthFunc: func(ctx context.Context) (*dto.HealthData, error) {
					return tt.mockReturn, tt.mockErr
				},
			}
			ctrl := controller.NewHealthController(mockSvc)

			req := httptest.NewRequest("GET", "/api/v1/health", nil)
			w := httptest.NewRecorder()

			// ============================================================
			// ACT
			// ============================================================
			ctrl.GetHealth(w, req)

			// ============================================================
			// ASSERT
			// ============================================================

			// Assert HTTP status code
			if w.Code != tt.wantStatus {
				t.Errorf("HTTP status = %d, want %d", w.Code, tt.wantStatus)
			}

			// For successful or unhealthy responses (non-500), validate JSON response
			if tt.wantStatus == http.StatusOK || tt.wantStatus == http.StatusServiceUnavailable {
				var resp gen.HealthResponse
				if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
					t.Fatal("failed to decode response JSON:", err)
				}

				// Assert envelope status
				if resp.Status != tt.wantRespStatus {
					t.Errorf("resp.Status = %s, want %s", resp.Status, tt.wantRespStatus)
				}

				// Assert data.status
				if resp.Data.Status != tt.wantDataStatus {
					t.Errorf("resp.Data.Status = %s, want %s", resp.Data.Status, tt.wantDataStatus)
				}

				// Assert checks map is present and matches expected values
				if tt.mockReturn != nil && tt.mockReturn.Checks != nil {
					if resp.Data.Checks == nil {
						t.Error("resp.Data.Checks is nil, expected non-nil")
					} else {
						for key, expectedValue := range tt.mockReturn.Checks {
							if actualValue, ok := (*resp.Data.Checks)[key]; !ok {
								t.Errorf("missing check key %q", key)
							} else if actualValue != expectedValue {
								t.Errorf("check[%q] = %s, want %s", key, actualValue, expectedValue)
							}
						}
					}
				}
			} else if tt.wantStatus == http.StatusInternalServerError {
				// Assert error response body
				expectedBody := "Internal server error\n"
				if w.Body.String() != expectedBody {
					t.Errorf("error body = %q, want %q", w.Body.String(), expectedBody)
				}
			}
		})
	}
}
