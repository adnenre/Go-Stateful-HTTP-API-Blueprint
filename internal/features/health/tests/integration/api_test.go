package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"rest-api-blueprint/internal/database"
	"rest-api-blueprint/internal/features/health/controller"
	"rest-api-blueprint/internal/features/health/repository"
	"rest-api-blueprint/internal/features/health/service"
	"rest-api-blueprint/internal/gen"

	"github.com/joho/godotenv"
)

// TestHealthIntegration verifies the health endpoint works with a real database.
func TestHealthIntegration(t *testing.T) {
	// ============================================================
	// ARRANGE
	// ============================================================

	// 1. Find project root directory (where .env is located)
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Skip("cannot determine test file path, skipping")
	}
	// Navigate up: internal/features/health/tests/integration/api_test.go -> project root
	projectRoot := filepath.Join(filepath.Dir(currentFile), "../../../../..")
	envPath := filepath.Join(projectRoot, ".env")

	// 2. Load .env if exists (ignore error if not found)
	_ = godotenv.Load(envPath)

	// 3. Read PostgreSQL credentials
	user := os.Getenv("POSTGRES_USER")
	password := os.Getenv("POSTGRES_PASSWORD")
	dbname := os.Getenv("POSTGRES_DB")
	port := os.Getenv("POSTGRES_PORT")
	if port == "" {
		port = "5432"
	}

	if user == "" || password == "" || dbname == "" {
		t.Skip("POSTGRES_USER, POSTGRES_PASSWORD, or POSTGRES_DB not set, skipping integration test")
	}

	// 4. Build DATABASE_URL
	dbURL := fmt.Sprintf("postgres://%s:%s@localhost:%s/%s?sslmode=disable", user, password, port, dbname)

	// 5. Connect to database
	if err := database.Connect(dbURL); err != nil {
		t.Skipf("database unavailable: %v", err)
	}
	defer func() {
		sqlDB, _ := database.DB.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
	}()

	// 6. Wire dependencies
	repo := repository.NewRepository(database.DB)
	svc := service.NewService(repo)
	ctrl := controller.NewHealthController(svc)

	// ============================================================
	// ACT
	// ============================================================
	req := httptest.NewRequest("GET", "/api/v1/health", nil)
	req = req.WithContext(context.Background())
	w := httptest.NewRecorder()
	ctrl.GetHealth(w, req)

	// ============================================================
	// ASSERT
	// ============================================================
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
	if resp.Data.Checks == nil {
		t.Error("expected checks to be present, got nil")
	} else if dbStatus, ok := (*resp.Data.Checks)["database"]; !ok {
		t.Error("expected database check in checks map")
	} else if dbStatus != "ok" {
		t.Errorf("expected database check 'ok', got %s", dbStatus)
	}
}
