package unit

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"rest-api-blueprint/internal/errors"
	"rest-api-blueprint/internal/features/health/controller"
	"rest-api-blueprint/internal/features/health/dto"
	"rest-api-blueprint/internal/gen"
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
			if w.Code != tt.wantStatus {
				t.Errorf("HTTP status = %d, want %d", w.Code, tt.wantStatus)
			}

			if tt.wantStatus == http.StatusOK || tt.wantStatus == http.StatusServiceUnavailable {
				var resp gen.HealthResponse
				if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
					t.Fatal(err)
				}
				if resp.Status != tt.wantRespStatus {
					t.Errorf("resp.Status = %s, want %s", resp.Status, tt.wantRespStatus)
				}
				if resp.Data.Status != tt.wantDataStatus {
					t.Errorf("resp.Data.Status = %s, want %s", resp.Data.Status, tt.wantDataStatus)
				}
				if tt.mockReturn != nil && tt.mockReturn.Checks != nil {
					if resp.Data.Checks == nil {
						t.Error("checks map missing")
					} else {
						for k, v := range tt.mockReturn.Checks {
							if (*resp.Data.Checks)[k] != v {
								t.Errorf("check[%s] = %s, want %s", k, (*resp.Data.Checks)[k], v)
							}
						}
					}
				}
			} else if tt.wantStatus == http.StatusInternalServerError {
				// Expect RFC 7807 problem details
				var problem errors.ProblemDetails
				if err := json.NewDecoder(w.Body).Decode(&problem); err != nil {
					t.Fatalf("failed to decode error response: %v", err)
				}
				if problem.Status != http.StatusInternalServerError {
					t.Errorf("problem.Status = %d, want %d", problem.Status, http.StatusInternalServerError)
				}
				if problem.Title != "Internal Server Error" {
					t.Errorf("problem.Title = %s, want 'Internal Server Error'", problem.Title)
				}
				if problem.Detail != "Failed to check health" {
					t.Errorf("problem.Detail = %s, want 'Failed to check health'", problem.Detail)
				}
				if problem.Instance == "" {
					t.Error("problem.Instance is empty")
				}
			}
		})
	}
}
