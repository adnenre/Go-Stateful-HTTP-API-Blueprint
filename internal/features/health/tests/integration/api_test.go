package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"rest-api-blueprint/internal/features/health/controller"
	"rest-api-blueprint/internal/features/health/repository"
	"rest-api-blueprint/internal/features/health/service"
	"rest-api-blueprint/internal/gen"
	"testing"
)

func TestHealthIntegration(t *testing.T) {
	repo := repository.NewRepository()
	svc := service.NewService(repo)
	ctrl := controller.NewHealthController(svc)

	req := httptest.NewRequest("GET", "/v1/health", nil)
	w := httptest.NewRecorder()
	ctrl.GetHealth(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp gen.HealthResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}
	if resp.Status != gen.Success {
		t.Errorf("expected status success, got %s", resp.Status)
	}
	if resp.Data.Status != gen.Healthy {
		t.Errorf("expected healthy, got %s", resp.Data.Status)
	}
}
