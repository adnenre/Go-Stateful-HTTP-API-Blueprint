package main

import (
	"log"
	"log/slog"
	"net/http"
	"os"
	"rest-api-blueprint/internal/config"
	"rest-api-blueprint/internal/features/health/controller"
	"rest-api-blueprint/internal/features/health/repository"
	"rest-api-blueprint/internal/features/health/service"
	"rest-api-blueprint/internal/gen"
	"rest-api-blueprint/internal/logger"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config error: %v", err)
	}
	logger.InitJSONLogger()
	slog.Info("starting application", "port", cfg.ServerPort)

	// Health feature wiring
	healthRepo := repository.NewRepository()
	healthSvc := service.NewService(healthRepo)
	healthCtrl := controller.NewHealthController(healthSvc)

	mux := http.NewServeMux()
	gen.HandlerFromMux(healthCtrl, mux)

	addr := ":" + cfg.ServerPort
	slog.Info("server starting", "address", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		slog.Error("server failed", "error", err)
		os.Exit(1)
	}
}
