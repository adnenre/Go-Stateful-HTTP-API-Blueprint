package unit

import (
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
	getHealthFunc func() (*dto.HealthData, error)
}

func (m *mockService) GetHealth() (*dto.HealthData, error) {
	return m.getHealthFunc()
}

func TestHealthController_GetHealth(t *testing.T) {
	tests := []struct {
		name           string
		mockReturn     *dto.HealthData
		mockErr        error
		wantStatus     int
		wantRespStatus gen.HealthResponseStatus
	}{
		{
			name: "success",
			mockReturn: &dto.HealthData{
				Status:    "healthy",
				Timestamp: "2026-04-22T12:00:00Z",
				Uptime:    "1s",
				Version:   "dev",
				Checks:    nil,
			},
			mockErr:        nil,
			wantStatus:     http.StatusOK,
			wantRespStatus: gen.Success,
		},
		{
			name:           "service error",
			mockReturn:     nil,
			mockErr:        fmt.Errorf("something went wrong"),
			wantStatus:     http.StatusInternalServerError,
			wantRespStatus: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := &mockService{
				getHealthFunc: func() (*dto.HealthData, error) {
					return tt.mockReturn, tt.mockErr
				},
			}
			ctrl := controller.NewHealthController(mockSvc)

			req := httptest.NewRequest("GET", "/v1/health", nil)
			w := httptest.NewRecorder()
			ctrl.GetHealth(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("got status %d, want %d", w.Code, tt.wantStatus)
			}

			if tt.wantStatus == http.StatusOK {
				var resp gen.HealthResponse
				if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
					t.Fatal(err)
				}
				if resp.Status != tt.wantRespStatus {
					t.Errorf("got status field %s, want %s", resp.Status, tt.wantRespStatus)
				}
				if resp.Data.Status != gen.Healthy {
					t.Errorf("got data.status %s, want %s", resp.Data.Status, gen.Healthy)
				}
			}
		})
	}
}
