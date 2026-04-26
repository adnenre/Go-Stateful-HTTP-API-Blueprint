package main

import (
	"log"
	"log/slog"
	"net/http"
	"os"
	"rest-api-blueprint/internal/cache"
	"rest-api-blueprint/internal/config"
	"rest-api-blueprint/internal/database"
	"rest-api-blueprint/internal/features/admin"
	adminController "rest-api-blueprint/internal/features/admin/controller"
	adminRepository "rest-api-blueprint/internal/features/admin/repository"
	adminService "rest-api-blueprint/internal/features/admin/service"
	"rest-api-blueprint/internal/features/auth"
	authController "rest-api-blueprint/internal/features/auth/controller"
	authRepository "rest-api-blueprint/internal/features/auth/repository"
	authService "rest-api-blueprint/internal/features/auth/service"
	healthController "rest-api-blueprint/internal/features/health/controller"
	healthRepository "rest-api-blueprint/internal/features/health/repository"
	healthService "rest-api-blueprint/internal/features/health/service"
	"rest-api-blueprint/internal/features/user"
	userController "rest-api-blueprint/internal/features/user/controller"
	userRepository "rest-api-blueprint/internal/features/user/repository"
	userService "rest-api-blueprint/internal/features/user/service"
	"rest-api-blueprint/internal/gen"
	"rest-api-blueprint/internal/logger"
	"rest-api-blueprint/internal/middleware"
)

// combinedServer implements all generated ServerInterface methods by embedding the feature controllers.
type combinedServer struct {
	*healthController.HealthController
	*authController.AuthController
	*userController.UserController
	*adminController.AdminController
}

// dtoResolver combines all feature resolvers for the validation middleware.
func dtoResolver(r *http.Request) (any, bool) {
	if dto, ok := auth.Resolver(r); ok {
		return dto, true
	}
	if dto, ok := user.Resolver(r); ok {
		return dto, true
	}
	if dto, ok := admin.Resolver(r); ok {
		return dto, true
	}
	return nil, false
}

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
	database.Connect(cfg.DatabaseURL)

	// ============================================================
	// 4. CONNECT TO REDIS (for rate limiting, health checks, cache)
	// ============================================================
	cache.InitRedis(cfg.RedisURL)

	// ============================================================
	// 5. RUN DATABASE MIGRATIONS (users and preferences tables)
	// ============================================================
	auth.Migrate() // creates/updates users table
	user.Migrate() // creates/updates user_preferences table

	// ============================================================
	// 6. WIRE DEPENDENCIES FOR ALL FEATURES
	// ============================================================
	// Health feature
	healthRepo := healthRepository.NewRepository(database.DB, cache.Client)
	healthSvc := healthService.NewService(healthRepo)
	healthCtrl := healthController.NewHealthController(healthSvc)

	// Auth feature
	authRepo := authRepository.NewRepository(database.DB)
	authSvc := authService.NewService(authRepo, cfg)
	authCtrl := authController.NewAuthController(authSvc)

	// User feature
	userRepo := userRepository.NewRepository(database.DB)
	userSvc := userService.NewService(userRepo)
	userCtrl := userController.NewUserController(userSvc)

	// Admin feature
	adminRepo := adminRepository.NewRepository(database.DB)
	adminSvc := adminService.NewService(adminRepo)
	adminCtrl := adminController.NewAdminController(adminSvc)

	// ============================================================
	// 7. COMBINE ALL CONTROLLERS INTO A SINGLE SERVER INTERFACE
	// ============================================================
	server := &combinedServer{
		HealthController: healthCtrl,
		AuthController:   authCtrl,
		UserController:   userCtrl,
		AdminController:  adminCtrl,
	}

	// ============================================================
	// 8. REGISTER ROUTES (generated from OpenAPI spec)
	// ============================================================
	mux := http.NewServeMux()
	handler := gen.HandlerFromMux(server, mux)

	// ============================================================
	// 9. SETUP RATE LIMITING
	// ============================================================
	rateLimiter := middleware.NewRateLimiter(cache.Client, cfg.RateLimitPerSecond)

	// ============================================================
	// 10. APPLY MIDDLEWARES – ORDER MATTERS!
	//
	// Execution order (outermost first):
	//   SecurityHeaders → CORS → RequestID → Logging → ValidateRequest → JWTAuth → RateLimiter → baseHandler
	//
	// Public paths that do NOT require JWT (used by JWTAuthMiddleware):
	//   - /api/v1/health
	//   - /api/v1/auth/login
	//   - /api/v1/auth/register
	// ============================================================
	publicPaths := map[string]bool{
		"/api/v1/health":        true,
		"/api/v1/auth/login":    true,
		"/api/v1/auth/register": true,
	}

	// Build chain from innermost to outermost:
	// (a) Rate limiting (innermost)
	handler = rateLimiter.Middleware(middleware.DefaultIPKeyFunc)(handler)

	// (b) JWT authentication – validates token, injects claims (skips public paths)
	handler = middleware.JWTAuthMiddleware(cfg, publicPaths)(handler)

	// (c) Global validation – decodes, validates, and restores request body
	handler = middleware.ValidateRequest(dtoResolver)(handler)

	// (d) Logging – logs request method, path, status, latency, request ID
	handler = middleware.Logging(handler)

	// (e) Request ID – generates/forwards X-Request-Id and stores in context
	handler = middleware.RequestIDMiddleware(handler)

	// (f) CORS – handles preflight requests and sets CORS headers
	corsMiddleware := middleware.NewCORS(middleware.CORSConfig{
		AllowedOrigins:   cfg.CORSAllowedOrigins,
		AllowedMethods:   cfg.CORSAllowedMethods,
		AllowedHeaders:   cfg.CORSAllowedHeaders,
		AllowCredentials: cfg.CORSAllowCredentials,
	})
	handler = corsMiddleware(handler)

	// (g) Security headers (outermost) – adds security headers
	securityMiddleware := middleware.SecurityHeaders(middleware.SecurityHeadersConfig{
		HSTSMaxAge: cfg.HSTSMaxAge,
	})
	handler = securityMiddleware(handler)

	// ============================================================
	// 11. START HTTP SERVER
	// ============================================================
	addr := ":" + cfg.ServerPort
	slog.Info("server starting", "address", addr)
	if err := http.ListenAndServe(addr, handler); err != nil {
		slog.Error("server failed", "error", err)
		os.Exit(1)
	}
}
