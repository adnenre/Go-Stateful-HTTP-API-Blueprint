package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"rest-api-blueprint/internal/auth"
	"rest-api-blueprint/internal/config"
	authModel "rest-api-blueprint/internal/features/auth/model"
	authRepo "rest-api-blueprint/internal/features/auth/repository"
	authService "rest-api-blueprint/internal/features/auth/service"
	userController "rest-api-blueprint/internal/features/user/controller"
	userModel "rest-api-blueprint/internal/features/user/model"
	userRepo "rest-api-blueprint/internal/features/user/repository"
	userService "rest-api-blueprint/internal/features/user/service"
	"rest-api-blueprint/internal/middleware"
)

func TestUserIntegration(t *testing.T) {
	// ============================================================
	// ARRANGE
	// ============================================================
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Skip("cannot determine test file path")
	}
	projectRoot := filepath.Join(filepath.Dir(currentFile), "../../../../..")
	envPath := filepath.Join(projectRoot, ".env")
	_ = godotenv.Load(envPath)

	// DB connection
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

	gormDB, err := gorm.Open(postgres.Open(dbURL), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Skipf("database unavailable: %v", err)
	}
	sqlDB, _ := gormDB.DB()
	defer sqlDB.Close()

	// Migrate tables (users and preferences)
	if err := gormDB.AutoMigrate(&authModel.User{}, &userModel.UserPreferences{}); err != nil {
		t.Skipf("failed to migrate: %v", err)
	}

	// ============================================================
	// HARD DELETE ANY PREVIOUS TEST DATA FOR USER-INTEGRATION
	// ============================================================
	gormDB.Unscoped().Where("email = ?", "user-integration@example.com").Delete(&authModel.User{})
	gormDB.Unscoped().Where("user_id = ?", "user-integration-id").Delete(&userModel.UserPreferences{})

	// Redis (not required for user, but we still connect to avoid panic)
	redisAddr := os.Getenv("REDIS_URL")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}
	redisClient := redis.NewClient(&redis.Options{Addr: redisAddr})
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_ = redisClient.Ping(ctx).Err() // ignore, user feature doesn't use Redis
	defer redisClient.Close()

	// Config
	cfg := &config.Config{
		JWTSecret: "test-secret",
		JWTExpiry: 15 * time.Minute,
	}

	// Create a test user via auth service
	authRepoInstance := authRepo.NewRepository(gormDB)
	authSvc := authService.NewService(authRepoInstance, cfg)
	userID, err := authSvc.Register(context.Background(), "user-integration@example.com", "userint", "testpass", nil)
	if err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}
	_ = userID

	// Login to get token and claims
	token, err := authSvc.Login(context.Background(), "user-integration@example.com", "testpass")
	if err != nil {
		t.Fatalf("failed to login: %v", err)
	}
	claims, err := auth.ValidateToken(token, cfg.JWTSecret)
	if err != nil {
		t.Fatalf("failed to validate token: %v", err)
	}

	// Wire user controller
	userRepoInstance := userRepo.NewRepository(gormDB)
	userSvc := userService.NewService(userRepoInstance)
	ctrl := userController.NewUserController(userSvc)

	// ============================================================
	// ACT & ASSERT – GET /users/me
	// ============================================================
	req := httptest.NewRequest("GET", "/api/v1/users/me", nil)
	// Inject claims and request ID into context (simulate middleware)
	ctxWithValues := context.WithValue(req.Context(), middleware.UserKey, claims)
	ctxWithValues = context.WithValue(ctxWithValues, middleware.RequestIDKey, "test-request-id")
	req = req.WithContext(ctxWithValues)
	w := httptest.NewRecorder()
	ctrl.GetMe(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var profile map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&profile); err != nil {
		t.Fatal(err)
	}
	if profile["email"] != "user-integration@example.com" {
		t.Errorf("expected email user-integration@example.com, got %v", profile["email"])
	}
	if profile["username"] != "userint" {
		t.Errorf("expected username userint, got %v", profile["username"])
	}

	// ============================================================
	// ACT & ASSERT – PATCH /users/me/preferences
	// ============================================================
	prefsPayload := map[string]interface{}{
		"notifications": true,
		"language":      "fr",
	}
	body, _ := json.Marshal(prefsPayload)
	req = httptest.NewRequest("PATCH", "/api/v1/users/me/preferences", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(ctxWithValues) // same context with claims
	w = httptest.NewRecorder()
	ctrl.UpdatePreferences(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var prefsResp map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&prefsResp); err != nil {
		t.Fatal(err)
	}
	if prefsResp["notifications"] != true {
		t.Errorf("expected notifications true, got %v", prefsResp["notifications"])
	}
	if prefsResp["language"] != "fr" {
		t.Errorf("expected language fr, got %v", prefsResp["language"])
	}
}
