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
	"rest-api-blueprint/internal/middleware"
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
	// 4. CONNECT TO REDIS (for rate limiting, health checks, cache)
	// ============================================================
	if err := cache.InitRedis(cfg.RedisURL); err != nil {
		slog.Error("redis connection failed", "error", err)
		os.Exit(1)
	}
	slog.Info("redis connected", "url", cfg.RedisURL)

	// ============================================================
	// 5. SETUP RATE LIMITING (distributed, using Redis)
	// ============================================================
	rateLimiter := middleware.NewRateLimiter(cache.Client, cfg.RateLimitPerSecond)

	// ============================================================
	// 6. WIRE DEPENDENCIES FOR THE HEALTH FEATURE
	// ============================================================
	healthRepo := repository.NewRepository(database.DB, cache.Client)
	healthSvc := service.NewService(healthRepo)
	healthCtrl := controller.NewHealthController(healthSvc)

	// ============================================================
	// 7. REGISTER ROUTES (generated from OpenAPI spec)
	// ============================================================
	mux := http.NewServeMux()
	handler := gen.HandlerFromMux(healthCtrl, mux)

	// ============================================================
	// 8. APPLY MIDDLEWARES (order: RequestID -> Logging -> RateLimiter)
	// ============================================================
	handler = middleware.RequestIDMiddleware(handler) // Generate/read request ID
	handler = middleware.Logging(handler)             // Log request with request ID
	handler = rateLimiter.Middleware(middleware.DefaultIPKeyFunc)(handler)

	// ============================================================
	// 9. START HTTP SERVER
	// ============================================================
	addr := ":" + cfg.ServerPort
	slog.Info("server starting", "address", addr)
	if err := http.ListenAndServe(addr, handler); err != nil {
		slog.Error("server failed", "error", err)
		os.Exit(1)
	}
}
