package integration

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"rest-api-blueprint/internal/database"
	"rest-api-blueprint/internal/features/health/controller"
	"rest-api-blueprint/internal/features/health/repository"
	"rest-api-blueprint/internal/features/health/service"
	"rest-api-blueprint/internal/gen"
	"testing"
)

// This integration test assumes a running PostgreSQL instance.
// Use `make docker-up` before running, or run it with the database available.

func TestHealthIntegration(t *testing.T) {
	// Connect to the real database (use test database or the same config)
	// For simplicity, we assume DATABASE_URL is set in the environment.
	dbURL := getEnvOrDefault("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/rest_api_blueprint?sslmode=disable")
	if err := database.Connect(dbURL); err != nil {
		t.Skipf("Skipping integration test, database not available: %v", err)
		return
	}
	defer func() {
		sqlDB, _ := database.DB.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
	}()

	// Wire dependencies
	repo := repository.NewRepository(database.DB)
	svc := service.NewService(repo)
	ctrl := controller.NewHealthController(svc)

	req := httptest.NewRequest("GET", "/api/v1/health", nil)
	req = req.WithContext(context.Background())
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
	// Check that the database check is present and "ok"
	if resp.Data.Checks == nil {
		t.Error("expected checks to be present, got nil")
	} else if dbStatus, ok := (*resp.Data.Checks)["database"]; !ok {
		t.Error("expected database check in checks map")
	} else if dbStatus != "ok" {
		t.Errorf("expected database check 'ok', got %s", dbStatus)
	}
}

func getEnvOrDefault(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
