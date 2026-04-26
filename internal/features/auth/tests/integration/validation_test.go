package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"rest-api-blueprint/internal/config"
	"rest-api-blueprint/internal/features/auth/controller"
	"rest-api-blueprint/internal/features/auth/dto"
	authModel "rest-api-blueprint/internal/features/auth/model"
	"rest-api-blueprint/internal/features/auth/repository"
	"rest-api-blueprint/internal/features/auth/service"
	"rest-api-blueprint/internal/middleware"
)

func TestValidationIntegration_Register(t *testing.T) {
	// ============================================================
	// ARRANGE – set up real DB (skip if not available)
	// ============================================================
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/rest_api_blueprint?sslmode=disable"
	}
	gormDB, err := gorm.Open(postgres.Open(dbURL), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Skipf("database unavailable: %v", err)
	}
	sqlDB, _ := gormDB.DB()
	defer sqlDB.Close()

	// Clean any leftover test data
	gormDB.Where("email = ?", "integration@example.com").Delete(&authModel.User{})

	// Create auth service and controller (using test config)
	authRepo := repository.NewRepository(gormDB)
	cfg := &config.Config{JWTSecret: "test", JWTExpiry: 15 * time.Minute}
	authSvc := service.NewService(authRepo, cfg)
	authCtrl := controller.NewAuthController(authSvc)

	// Prepare request with missing username (invalid)
	invalidBody := dto.RegisterRequest{
		Email:    "integration@example.com",
		Password: "Strong123",
		// Username missing
	}
	bodyBytes, _ := json.Marshal(invalidBody)
	req := httptest.NewRequest("POST", "/api/v1/auth/register", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Create resolver that returns a fresh DTO for this route
	resolver := func(r *http.Request) (any, bool) {
		if r.URL.Path == "/api/v1/auth/register" && r.Method == "POST" {
			return &dto.RegisterRequest{}, true
		}
		return nil, false
	}

	// Apply validation middleware to the controller handler
	handler := middleware.ValidateRequest(resolver)(http.HandlerFunc(authCtrl.Register))

	// ============================================================
	// ACT
	// ============================================================
	handler.ServeHTTP(w, req)

	// ============================================================
	// ASSERT
	// ============================================================
	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	var problem map[string]interface{}
	err = json.NewDecoder(w.Body).Decode(&problem)
	assert.NoError(t, err)
	assert.Equal(t, "/errors/validation.html", problem["type"]) // changed
	assert.Contains(t, problem["errors"], "username")
}
