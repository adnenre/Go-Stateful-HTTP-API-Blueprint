// main.go
package main

import (
	"embed"
	"log/slog"
	"net/http"
	"rest-api-blueprint/internal/app"
	"rest-api-blueprint/internal/cache"
	"rest-api-blueprint/internal/config"
	"rest-api-blueprint/internal/database"
	"rest-api-blueprint/internal/email"
	"rest-api-blueprint/internal/errors"
	"rest-api-blueprint/internal/features/auth"
	"rest-api-blueprint/internal/features/user"
	"rest-api-blueprint/internal/logger"
)

//go:embed web/docs web/errors
var staticFS embed.FS

func main() {
	// 1. Load configuration (exits on error)
	cfg := config.Load()

	// 2. Initialize logging
	logger.InitJSONLogger()
	slog.Info("starting application", "port", cfg.ServerPort)

	// 3. Initialize error documentation base URL
	errors.Init(cfg.ErrorDocsBaseURL)

	// 4. Connect to database (exits on error)
	database.Connect(cfg.DatabaseURL)

	// 5. Connect to Redis (exits on error)
	cache.InitRedis(cfg.RedisURL)

	// 6. Initialize email sender
	emailSender := email.InitSender(cfg)

	// 7. Run database migrations
	auth.Migrate()
	user.Migrate()

	// 8. Build all controllers (repositories, services, controllers)
	combined := app.BuildControllers(cfg, emailSender)

	// 9. Setup HTTP server (routes, middleware)
	handler, err := app.SetupServer(cfg, combined, app.DTOResolver, staticFS)
	if err != nil {
		slog.Error("failed to setup server", "error", err)
		// In original main, similar errors would have called os.Exit(1)
		// We preserve that behaviour by exiting here.
		return
	}

	// 10. Start HTTP server
	addr := ":" + cfg.ServerPort
	slog.Info("server starting", "address", addr)
	if err := http.ListenAndServe(addr, handler); err != nil {
		slog.Error("server failed", "error", err)
	}
}
