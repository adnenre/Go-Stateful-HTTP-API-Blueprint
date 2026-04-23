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

	"rest-api-blueprint/internal/cache"
	"rest-api-blueprint/internal/database"
	"rest-api-blueprint/internal/features/health/controller"
	"rest-api-blueprint/internal/features/health/repository"
	"rest-api-blueprint/internal/features/health/service"
	"rest-api-blueprint/internal/gen"

	"github.com/joho/godotenv"
)

func TestHealthIntegration(t *testing.T) {
	// ============================================================
	// ARRANGE
	// ============================================================

	// Find project root and load .env
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Skip("cannot determine test file path")
	}
	projectRoot := filepath.Join(filepath.Dir(currentFile), "../../../../..")
	envPath := filepath.Join(projectRoot, ".env")
	_ = godotenv.Load(envPath)

	// Get credentials
	user := os.Getenv("POSTGRES_USER")
	password := os.Getenv("POSTGRES_PASSWORD")
	dbname := os.Getenv("POSTGRES_DB")
	port := os.Getenv("POSTGRES_PORT")
	if port == "" {
		port = "5432"
	}
	if user == "" || password == "" || dbname == "" {
		t.Skip("PostgreSQL credentials missing")
	}
	dbURL := fmt.Sprintf("postgres://%s:%s@localhost:%s/%s?sslmode=disable", user, password, port, dbname)

	// Connect to DB
	if err := database.Connect(dbURL); err != nil {
		t.Skipf("database unavailable: %v", err)
	}
	defer func() {
		if db, _ := database.DB.DB(); db != nil {
			db.Close()
		}
	}()

	// Connect to Redis
	redisAddr := os.Getenv("REDIS_URL")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}
	if err := cache.InitRedis(redisAddr); err != nil {
		t.Skipf("redis unavailable: %v", err)
	}
	// No need to close Redis client globally

	// Wire dependencies
	repo := repository.NewRepository(database.DB, cache.Client)
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
		t.Error("checks map missing")
	} else {
		dbCheck, okDB := (*resp.Data.Checks)["database"]
		if !okDB || dbCheck != "ok" {
			t.Errorf("database check expected 'ok', got %v", dbCheck)
		}
		redisCheck, okRedis := (*resp.Data.Checks)["redis"]
		if !okRedis || redisCheck != "ok" {
			t.Errorf("redis check expected 'ok', got %v", redisCheck)
		}
	}
}
