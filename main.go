package main

import (
	"embed"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"rest-api-blueprint/internal/cache"
	"rest-api-blueprint/internal/config"
	"rest-api-blueprint/internal/database"
	"rest-api-blueprint/internal/email"
	"rest-api-blueprint/internal/errors"
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

//go:embed web/docs web/errors
var staticFS embed.FS

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
	cfg := config.Load()

	// ============================================================
	// 2. INITIALIZE STRUCTURED LOGGING (JSON output)
	// ============================================================
	logger.InitJSONLogger()
	slog.Info("starting application", "port", cfg.ServerPort)

	// ============================================================
	// 2.1 INITIALIZE ERROR DOCUMENTATION BASE URL (RFC 7807)
	// ============================================================
	errors.Init(cfg.ErrorDocsBaseURL)

	// ============================================================
	// 3. CONNECT TO DATABASE (PostgreSQL via GORM)
	// ============================================================
	database.Connect(cfg.DatabaseURL)

	// ============================================================
	// 4. CONNECT TO REDIS (for rate limiting, health checks, cache)
	// ============================================================
	cache.InitRedis(cfg.RedisURL)
	// ============================================================
	// 4.1 INITIALIZE EMAIL SENDER
	// ============================================================
	emailSender := email.InitSender(cfg)
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
	authSvc := authService.NewService(authRepo, cfg, cache.Client, emailSender)
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
	// 8. CREATE ROUTER AND REGISTER STATIC DOCUMENTATION ROUTES
	// ============================================================
	mux := http.NewServeMux()

	// Serve OpenAPI spec
	openapiSpec, err := os.ReadFile("api/openapi.yaml")
	if err != nil {
		slog.Error("failed to read openapi.yaml", "error", err)
		os.Exit(1)
	}
	mux.HandleFunc("GET /openapi.yaml", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/yaml")
		w.Write(openapiSpec)
	})

	// Serve Swagger UI (static files)
	docsSub, err := fs.Sub(staticFS, "web/docs")
	if err != nil {
		slog.Error("failed to get docs subfolder", "error", err)
		os.Exit(1)
	}
	mux.Handle("GET /docs/", http.StripPrefix("/docs/", http.FileServer(http.FS(docsSub))))

	// Serve error documentation pages
	errorsSub, err := fs.Sub(staticFS, "web/errors")
	if err != nil {
		slog.Error("failed to get errors subfolder", "error", err)
		os.Exit(1)
	}
	mux.Handle("GET /errors/", http.StripPrefix("/errors/", http.FileServer(http.FS(errorsSub))))

	// ============================================================
	// 9. REGISTER API ROUTES (generated from OpenAPI spec)
	// ============================================================
	handler := gen.HandlerFromMuxWithBaseURL(server, mux, "/api/v1") // mux already contains static routes

	// ============================================================
	// 10. SETUP RATE LIMITING
	// ============================================================
	rateLimiter := middleware.NewRateLimiter(cache.Client, cfg.RateLimitPerSecond)

	// ============================================================
	// 11. APPLY MIDDLEWARES – ORDER MATTERS!
	//
	// Execution order (outermost first):
	//   PanicRecovery → SecurityHeaders → CORS → RequestID → Logging → ValidateRequest → JWTAuth → RateLimiter → baseHandler
	//

	// Build chain from innermost to outermost:
	// (a) Rate limiting (innermost)
	handler = rateLimiter.Middleware(middleware.DefaultIPKeyFunc)(handler)

	// (b) JWT authentication – validates token, injects claims (skips public paths)
	handler = middleware.JWTAuthMiddleware(cfg)(handler)

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

	// (g) Security headers – adds security headers
	securityMiddleware := middleware.SecurityHeaders(middleware.SecurityHeadersConfig{
		HSTSMaxAge: cfg.HSTSMaxAge,
	})
	handler = securityMiddleware(handler)

	// (h) Panic recovery (outermost) – catches panics and returns RFC 7807 error
	handler = middleware.PanicRecovery(handler)

	// ============================================================
	// 12. START HTTP SERVER
	// ============================================================
	addr := ":" + cfg.ServerPort
	slog.Info("server starting", "address", addr)
	if err := http.ListenAndServe(addr, handler); err != nil {
		slog.Error("server failed", "error", err)
		os.Exit(1)
	}
}
