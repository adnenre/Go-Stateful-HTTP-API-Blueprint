package main

import (
	"log"
	"net/http"
	"rest-api-blueprint/internal/features/health/controller"
	"rest-api-blueprint/internal/features/health/repository"
	"rest-api-blueprint/internal/features/health/service"
	"rest-api-blueprint/internal/gen"
)

func main() {
	// Health feature wiring
	healthRepo := repository.NewRepository()
	healthSvc := service.NewService(healthRepo)
	healthCtrl := controller.NewHealthController(healthSvc)

	mux := http.NewServeMux()
	gen.HandlerFromMux(healthCtrl, mux)

	log.Println("Server starting on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
