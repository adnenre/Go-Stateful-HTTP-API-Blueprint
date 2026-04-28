// internal/app/server.go
package app

import (
	"io/fs"
	"net/http"
	"os"
	"rest-api-blueprint/internal/cache"
	"rest-api-blueprint/internal/config"
	"rest-api-blueprint/internal/gen"
	"rest-api-blueprint/internal/middleware"
)

// ============================================================
// SETUP SERVER – CREATES ROUTER, REGISTERS ROUTES, BUILDS MIDDLEWARE CHAIN
// ============================================================

// SetupServer creates a fully configured HTTP handler with all routes and middleware.
// It returns the final handler ready to be passed to http.ListenAndServe.
// On any unrecoverable error (e.g., missing static files), it returns an error.
func SetupServer(
	cfg *config.Config,
	combined *CombinedServer,
	dtoResolver func(*http.Request) (any, bool),
	staticFS fs.FS,
) (http.Handler, error) {
	// ============================================================
	// 1. CREATE ROUTER (SERVE MUX)
	// ============================================================
	mux := http.NewServeMux()

	// ============================================================
	// 2. SERVE OPENAPI SPECIFICATION (openapi.yaml from disk)
	// ============================================================
	openapiSpec, err := os.ReadFile("api/openapi.yaml")
	if err != nil {
		return nil, err
	}
	mux.HandleFunc("GET /openapi.yaml", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/yaml")
		w.Write(openapiSpec)
	})

	// ============================================================
	// 3. SERVE SWAGGER UI (embedded static files)
	// ============================================================
	docsSub, err := fs.Sub(staticFS, "web/docs")
	if err != nil {
		return nil, err
	}
	mux.Handle("GET /docs/", http.StripPrefix("/docs/", http.FileServer(http.FS(docsSub))))

	// ============================================================
	// 4. SERVE ERROR DOCUMENTATION PAGES (embedded)
	// ============================================================
	errorsSub, err := fs.Sub(staticFS, "web/errors")
	if err != nil {
		return nil, err
	}
	mux.Handle("GET /errors/", http.StripPrefix("/errors/", http.FileServer(http.FS(errorsSub))))

	// ============================================================
	// 5. REGISTER GENERATED API ROUTES (from OpenAPI spec)
	// ============================================================
	gen.HandlerFromMuxWithBaseURL(combined, mux, "/api/v1")

	// ============================================================
	// 6. BUILD MIDDLEWARE CHAIN – ORDER MATTERS!
	//
	// Execution order (outermost first):
	//   PanicRecovery → SecurityHeaders → CORS → RequestID → Logging → ValidateRequest → JWTAuth → RateLimiter → baseHandler
	// ============================================================

	// 6a. Rate limiting (innermost)
	rateLimiter := middleware.NewRateLimiter(cache.Client, cfg.RateLimitPerSecond)
	handler := rateLimiter.Middleware(middleware.DefaultIPKeyFunc)(mux)

	// 6b. JWT authentication – validates token, injects claims (skips public paths)
	handler = middleware.JWTAuthMiddleware(cfg)(handler)

	// 6c. Global validation – decodes, validates, and restores request body
	handler = middleware.ValidateRequest(dtoResolver)(handler)

	// 6d. Logging – logs request method, path, status, latency, request ID
	handler = middleware.Logging(handler)

	// 6e. Request ID – generates/forwards X-Request-Id and stores in context
	handler = middleware.RequestIDMiddleware(handler)

	// 6f. CORS – handles preflight requests and sets CORS headers
	corsMiddleware := middleware.NewCORS(middleware.CORSConfig{
		AllowedOrigins:   cfg.CORSAllowedOrigins,
		AllowedMethods:   cfg.CORSAllowedMethods,
		AllowedHeaders:   cfg.CORSAllowedHeaders,
		AllowCredentials: cfg.CORSAllowCredentials,
	})
	handler = corsMiddleware(handler)

	// 6g. Security headers – adds security headers
	securityMiddleware := middleware.SecurityHeaders(middleware.SecurityHeadersConfig{
		HSTSMaxAge: cfg.HSTSMaxAge,
	})
	handler = securityMiddleware(handler)

	// 6h. Panic recovery (outermost) – catches panics and returns RFC 7807 error
	handler = middleware.PanicRecovery(handler)

	return handler, nil
}
