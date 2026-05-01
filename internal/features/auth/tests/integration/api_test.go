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

	"rest-api-blueprint/internal/config"
	"rest-api-blueprint/internal/email"
	"rest-api-blueprint/internal/features/auth/controller"
	"rest-api-blueprint/internal/features/auth/dto"
	authModel "rest-api-blueprint/internal/features/auth/model"
	"rest-api-blueprint/internal/features/auth/repository"
	"rest-api-blueprint/internal/features/auth/service"
)

// Helper to get cookie value by name from response
func getCookieValue(w *httptest.ResponseRecorder, name string) string {
	for _, c := range w.Result().Cookies() {
		if c.Name == name {
			return c.Value
		}
	}
	return ""
}

func TestAuthIntegration(t *testing.T) {
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

	// PostgreSQL
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

	// Migrate users table
	if err := gormDB.AutoMigrate(&authModel.User{}); err != nil {
		t.Skipf("failed to migrate: %v", err)
	}

	// Clean previous test data (hard delete)
	gormDB.Unscoped().Where("email = ?", "integration@example.com").Delete(&authModel.User{})

	// Redis
	redisAddr := os.Getenv("REDIS_URL")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}
	redisClient := redis.NewClient(&redis.Options{Addr: redisAddr})
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := redisClient.Ping(ctx).Err(); err != nil {
		t.Skipf("redis unavailable: %v", err)
	}
	defer redisClient.Close()

	// Config for JWT and refresh token
	cfg := &config.Config{
		JWTSecret:          "test-secret",
		JWTExpiry:          15 * time.Minute,
		RefreshTokenExpiry: 168 * time.Hour, // 7 days
	}

	// Wire dependencies
	emailSender := &email.MockSender{}
	repo := repository.NewRepository(gormDB)
	svc := service.NewService(repo, cfg, redisClient, emailSender)
	ctrl := controller.NewAuthController(svc, cfg)

	// ============================================================
	// ACT & ASSERT – Register (should return 202 Accepted)
	// ============================================================
	registerReq := dto.RegisterRequest{
		Email:    "integration@example.com",
		Password: "testpass123",
		Username: "integrationuser",
	}
	body, _ := json.Marshal(registerReq)
	httpReq := httptest.NewRequest("POST", "/api/v1/auth/register", bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	ctrl.Register(w, httpReq)
	if w.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d", w.Code)
	}

	// ============================================================
	// Retrieve OTP from Redis
	// ============================================================
	otp, err := redisClient.Get(context.Background(), fmt.Sprintf("otp:%s", registerReq.Email)).Result()
	if err != nil {
		t.Fatalf("failed to get OTP from Redis: %v", err)
	}
	t.Logf("OTP for %s: %s", registerReq.Email, otp)

	// ============================================================
	// ACT & ASSERT – Verify OTP (should return 200 and set cookies)
	// ============================================================
	verifyReq := dto.VerifyOTPRequest{
		Email: registerReq.Email,
		OTP:   otp,
	}
	body, _ = json.Marshal(verifyReq)
	httpReq = httptest.NewRequest("POST", "/api/v1/auth/verify-otp", bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	ctrl.VerifyOtp(w, httpReq)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	// Decode JSON response (only metadata, no access_token)
	var verifyResp dto.OTPResponse
	if err := json.NewDecoder(w.Body).Decode(&verifyResp); err != nil {
		t.Fatal(err)
	}
	if verifyResp.ExpiresIn <= 0 {
		t.Error("expires_in not set in response")
	}
	if verifyResp.User.Email != registerReq.Email {
		t.Errorf("expected user email %s, got %s", registerReq.Email, verifyResp.User.Email)
	}
	// Check that cookies are set
	accessTokenCookie := getCookieValue(w, "access_token")
	if accessTokenCookie == "" {
		t.Error("access_token cookie not set")
	}
	refreshTokenCookie := getCookieValue(w, "refresh_token")
	if refreshTokenCookie == "" {
		t.Error("refresh_token cookie not set")
	}

	// ============================================================
	// ACT & ASSERT – Login with the same credentials
	// ============================================================
	loginReq := dto.LoginRequest{
		Email:    registerReq.Email,
		Password: registerReq.Password,
	}
	body, _ = json.Marshal(loginReq)
	httpReq = httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	ctrl.Login(w, httpReq)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var loginResp dto.LoginResponse
	if err := json.NewDecoder(w.Body).Decode(&loginResp); err != nil {
		t.Fatal(err)
	}
	if loginResp.ExpiresIn <= 0 {
		t.Error("expires_in not set")
	}
	if loginResp.User.Email != registerReq.Email {
		t.Errorf("expected user email %s, got %s", registerReq.Email, loginResp.User.Email)
	}
	// Check that new cookies are set (should be refreshed)
	newAccessCookie := getCookieValue(w, "access_token")
	if newAccessCookie == "" {
		t.Error("access_token cookie not set after login")
	}
	newRefreshCookie := getCookieValue(w, "refresh_token")
	if newRefreshCookie == "" {
		t.Error("refresh_token cookie not set after login")
	}
	// They should be different from the previous ones? Not required to check but okay.

	// ============================================================
	// ACT & ASSERT – Duplicate registration (should return 409)
	// ============================================================
	body, _ = json.Marshal(registerReq)
	httpReq = httptest.NewRequest("POST", "/api/v1/auth/register", bytes.NewReader(body))
	w = httptest.NewRecorder()
	ctrl.Register(w, httpReq)
	if w.Code != http.StatusConflict {
		t.Errorf("expected 409, got %d", w.Code)
	}
}
