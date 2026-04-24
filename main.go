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
	// 8. APPLY MIDDLEWARES – ORDER MATTERS!
	//
	// The middlewares are applied from innermost to outermost.
	// Execution order (outermost first) will be:
	//   SecurityHeaders → CORS → RequestID → Logging → RateLimiter → baseHandler
	//
	// This ensures that RequestID and Logging run before the RateLimiter,
	// so even when a request is rate‑limited (429), the X-Request-Id header
	// is already set, the request ID is logged, and the error response's
	// 'instance' field contains the correct value.
	// ============================================================

	// 8.1 Rate limiting (innermost)
	handler = rateLimiter.Middleware(middleware.DefaultIPKeyFunc)(handler)

	// 8.2 Logging – logs request method, path, status, latency, and request ID
	handler = middleware.Logging(handler)

	// 8.3 Request ID – generates/forwards X-Request-Id and stores it in context
	handler = middleware.RequestIDMiddleware(handler)

	// 8.4 CORS – handles preflight requests and sets CORS headers
	corsMiddleware := middleware.NewCORS(middleware.CORSConfig{
		AllowedOrigins:   cfg.CORSAllowedOrigins,
		AllowedMethods:   cfg.CORSAllowedMethods,
		AllowedHeaders:   cfg.CORSAllowedHeaders,
		AllowCredentials: cfg.CORSAllowCredentials,
	})
	handler = corsMiddleware(handler)

	// 8.5 Security headers (outermost) – adds X‑Content‑Type‑Options, etc.
	securityMiddleware := middleware.SecurityHeaders(middleware.SecurityHeadersConfig{
		HSTSMaxAge: cfg.HSTSMaxAge,
	})
	handler = securityMiddleware(handler)

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
