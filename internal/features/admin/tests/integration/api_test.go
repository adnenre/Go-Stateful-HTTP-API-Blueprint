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
	adminController "rest-api-blueprint/internal/features/admin/controller"
	adminDto "rest-api-blueprint/internal/features/admin/dto"
	adminRepo "rest-api-blueprint/internal/features/admin/repository"
	adminService "rest-api-blueprint/internal/features/admin/service"
	authModel "rest-api-blueprint/internal/features/auth/model"
	authRepo "rest-api-blueprint/internal/features/auth/repository"
	authService "rest-api-blueprint/internal/features/auth/service"
	userModel "rest-api-blueprint/internal/features/user/model"
	"rest-api-blueprint/internal/gen"
	"rest-api-blueprint/internal/middleware"
)

func TestAdminIntegration(t *testing.T) {
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

	// Clean previous test data
	gormDB.Where("email = ?", "admin@example.com").Delete(&authModel.User{})
	gormDB.Where("email = ?", "newuser@example.com").Delete(&authModel.User{})
	gormDB.Where("email = ?", "admin-integration@example.com").Delete(&authModel.User{})

	// Redis (not strictly required for admin)
	redisAddr := os.Getenv("REDIS_URL")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}
	redisClient := redis.NewClient(&redis.Options{Addr: redisAddr})
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_ = redisClient.Ping(ctx).Err()
	defer redisClient.Close()

	// Config
	cfg := &config.Config{
		JWTSecret: "test-secret",
		JWTExpiry: 15 * time.Minute,
	}

	// Create an admin user via auth service (role "admin")
	authRepoInstance := authRepo.NewRepository(gormDB)
	authSvc := authService.NewService(authRepoInstance, cfg)
	adminID, err := authSvc.Register(context.Background(), "admin@example.com", "adminuser", "adminpass", nil)
	if err != nil {
		t.Fatalf("failed to create admin user: %v", err)
	}
	// Promote user to admin (direct DB update, because registration only creates "user")
	if err := gormDB.Model(&authModel.User{}).Where("id = ?", adminID).Update("role", "admin").Error; err != nil {
		t.Fatalf("failed to promote to admin: %v", err)
	}

	// Login as admin to obtain token and claims
	token, err := authSvc.Login(context.Background(), "admin@example.com", "adminpass")
	if err != nil {
		t.Fatalf("admin login failed: %v", err)
	}
	claims, err := auth.ValidateToken(token, cfg.JWTSecret)
	if err != nil {
		t.Fatalf("failed to validate admin token: %v", err)
	}

	// Wire admin controller
	adminRepoInstance := adminRepo.NewRepository(gormDB)
	adminSvc := adminService.NewService(adminRepoInstance)
	ctrl := adminController.NewAdminController(adminSvc)

	// Helper to inject claims into request context
	injectClaims := func(r *http.Request) *http.Request {
		ctx := context.WithValue(r.Context(), middleware.UserKey, claims)
		// Also inject a dummy request ID (optional)
		ctx = context.WithValue(ctx, middleware.RequestIDKey, "test-request-id")
		return r.WithContext(ctx)
	}

	// ============================================================
	// ACT & ASSERT – Create a new user via admin endpoint
	// ============================================================
	createReq := adminDto.CreateUserRequest{
		Email:    "newuser@example.com",
		Password: "newpass123",
		Role:     "user",
		Username: "newuser",
	}
	body, _ := json.Marshal(createReq)
	httpReq := httptest.NewRequest("POST", "/api/v1/admin/users", bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq = injectClaims(httpReq)
	w := httptest.NewRecorder()
	ctrl.CreateUser(w, httpReq)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Code)
	}
	var createdUser map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&createdUser); err != nil {
		t.Fatal(err)
	}
	userID := createdUser["id"].(string)

	// ============================================================
	// ACT & ASSERT – List users
	// ============================================================
	httpReq = httptest.NewRequest("GET", "/api/v1/admin/users?limit=10", nil)
	httpReq = injectClaims(httpReq)
	w = httptest.NewRecorder()
	ctrl.ListUsers(w, httpReq, gen.ListUsersParams{Limit: intPtr(10), Offset: intPtr(0)})

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var users []map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&users); err != nil {
		t.Fatal(err)
	}
	var found bool
	for _, u := range users {
		if u["id"] == userID {
			found = true
			break
		}
	}
	if !found {
		t.Error("created user not found in list")
	}

	// ============================================================
	// ACT & ASSERT – Get user by ID
	// ============================================================
	httpReq = httptest.NewRequest("GET", "/api/v1/admin/users/"+userID, nil)
	httpReq = injectClaims(httpReq)
	w = httptest.NewRecorder()
	ctrl.GetUser(w, httpReq, userID)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var fetchedUser map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&fetchedUser); err != nil {
		t.Fatal(err)
	}
	if fetchedUser["email"] != createReq.Email {
		t.Errorf("expected email %s, got %v", createReq.Email, fetchedUser["email"])
	}

	// ============================================================
	// ACT & ASSERT – Update user
	// ============================================================
	updatePayload := map[string]interface{}{
		"username": "updateduser",
	}
	body, _ = json.Marshal(updatePayload)
	httpReq = httptest.NewRequest("PUT", "/api/v1/admin/users/"+userID, bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq = injectClaims(httpReq)
	w = httptest.NewRecorder()
	ctrl.UpdateUser(w, httpReq, userID)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	// Verify update
	httpReq = httptest.NewRequest("GET", "/api/v1/admin/users/"+userID, nil)
	httpReq = injectClaims(httpReq)
	w = httptest.NewRecorder()
	ctrl.GetUser(w, httpReq, userID)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if err := json.NewDecoder(w.Body).Decode(&fetchedUser); err != nil {
		t.Fatal(err)
	}
	if fetchedUser["username"] != "updateduser" {
		t.Errorf("expected username updateduser, got %v", fetchedUser["username"])
	}

	// ============================================================
	// ACT & ASSERT – Delete user
	// ============================================================
	httpReq = httptest.NewRequest("DELETE", "/api/v1/admin/users/"+userID, nil)
	httpReq = injectClaims(httpReq)
	w = httptest.NewRecorder()
	ctrl.DeleteUser(w, httpReq, userID)

	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", w.Code)
	}

	// Verify deletion
	httpReq = httptest.NewRequest("GET", "/api/v1/admin/users/"+userID, nil)
	httpReq = injectClaims(httpReq)
	w = httptest.NewRecorder()
	ctrl.GetUser(w, httpReq, userID)
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404 after deletion, got %d", w.Code)
	}
}

func intPtr(i int) *int {
	return &i
}
