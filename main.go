package main

import (
	"log"
	"log/slog"
	"net/http"
	"os"
	"rest-api-blueprint/internal/cache"
	"rest-api-blueprint/internal/config"
	"rest-api-blueprint/internal/database"
	"rest-api-blueprint/internal/features/health/controller"
	"rest-api-blueprint/internal/features/health/repository"
	"rest-api-blueprint/internal/features/health/service"
	"rest-api-blueprint/internal/gen"
	"rest-api-blueprint/internal/logger"
)

func main() {
	// ============================================================
	// 1. LOAD CONFIGURATION (fail‑fast if missing required vars)
	// ============================================================
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	// ============================================================
	// 2. INITIALIZE STRUCTURED LOGGING (JSON output)
	// ============================================================
	logger.InitJSONLogger()
	slog.Info("starting application", "port", cfg.ServerPort)

	// ============================================================
	// 3. CONNECT TO DATABASE (PostgreSQL via GORM)
	// ============================================================
	if err := database.Connect(cfg.DatabaseURL); err != nil {
		slog.Error("database connection failed", "error", err)
		os.Exit(1)
	}
	slog.Info("database connected")

	// ============================================================
	// 4. CONNECT TO REDIS (currently used for health checks only)
	//    TODO: add rate limiting, idempotency, caching in future phases
	// ============================================================
	if err := cache.InitRedis(cfg.RedisURL); err != nil {
		slog.Error("redis connection failed", "error", err)
		os.Exit(1)
	}
	slog.Info("redis connected", "url", cfg.RedisURL)

	// ============================================================
	// 5. WIRE DEPENDENCIES FOR THE HEALTH FEATURE
	// ============================================================
	// Repository receives both database and Redis clients
	healthRepo := repository.NewRepository(database.DB, cache.Client)
	// Service uses the repository to check dependencies
	healthSvc := service.NewService(healthRepo)
	// Controller handles the HTTP request
	healthCtrl := controller.NewHealthController(healthSvc)

	// ============================================================
	// 6. REGISTER ROUTES (generated from OpenAPI spec)
	// ============================================================
	mux := http.NewServeMux()
	gen.HandlerFromMux(healthCtrl, mux)

	// ============================================================
	// 7. START HTTP SERVER
	// ============================================================
	addr := ":" + cfg.ServerPort
	slog.Info("server starting", "address", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		slog.Error("server failed", "error", err)
		os.Exit(1)
	}
}
