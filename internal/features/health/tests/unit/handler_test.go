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
			name: "success_healthy",
			mockReturn: &dto.HealthData{
				Status:    "healthy",
				Timestamp: "2026-04-22T12:00:00Z",
				Uptime:    "1s",
				Version:   "dev",
				Checks:    map[string]string{"database": "ok"},
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
				Checks:    map[string]string{"database": "connection refused"},
			},
			mockErr:        nil,
			wantStatus:     http.StatusServiceUnavailable,
			wantRespStatus: gen.Success,
			wantDataStatus: gen.Unhealthy,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := &mockService{
				getHealthFunc: func(ctx context.Context) (*dto.HealthData, error) {
					return tt.mockReturn, tt.mockErr
				},
			}
			ctrl := controller.NewHealthController(mockSvc)

			req := httptest.NewRequest("GET", "/api/v1/health", nil)
			w := httptest.NewRecorder()
			ctrl.GetHealth(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("got status %d, want %d", w.Code, tt.wantStatus)
			}

			if tt.wantStatus == http.StatusOK || tt.wantStatus == http.StatusServiceUnavailable {
				var resp gen.HealthResponse
				if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
					t.Fatal(err)
				}
				if resp.Status != tt.wantRespStatus {
					t.Errorf("got status field %s, want %s", resp.Status, tt.wantRespStatus)
				}
				if resp.Data.Status != tt.wantDataStatus {
					t.Errorf("got data.status %s, want %s", resp.Data.Status, tt.wantDataStatus)
				}
				// Optionally check that checks are included
				if tt.mockReturn != nil && tt.mockReturn.Checks != nil {
					if resp.Data.Checks == nil {
						t.Error("expected checks to be present, got nil")
					} else if (*resp.Data.Checks)["database"] != tt.mockReturn.Checks["database"] {
						t.Errorf("expected database check %s, got %s", tt.mockReturn.Checks["database"], (*resp.Data.Checks)["database"])
					}
				}
			} else if tt.wantStatus == http.StatusInternalServerError {
				// No JSON response, just error text
				if w.Body.String() != "Internal server error\n" {
					t.Errorf("unexpected error body: %s", w.Body.String())
				}
			}
		})
	}
}
