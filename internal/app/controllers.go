// internal/app/controllers.go
package app

import (
	"rest-api-blueprint/internal/cache"
	"rest-api-blueprint/internal/config"
	"rest-api-blueprint/internal/database"
	"rest-api-blueprint/internal/email"

	adminController "rest-api-blueprint/internal/features/admin/controller"
	adminRepository "rest-api-blueprint/internal/features/admin/repository"
	adminService "rest-api-blueprint/internal/features/admin/service"

	authController "rest-api-blueprint/internal/features/auth/controller"
	authRepository "rest-api-blueprint/internal/features/auth/repository"
	authService "rest-api-blueprint/internal/features/auth/service"

	healthController "rest-api-blueprint/internal/features/health/controller"
	healthRepository "rest-api-blueprint/internal/features/health/repository"
	healthService "rest-api-blueprint/internal/features/health/service"

	userController "rest-api-blueprint/internal/features/user/controller"
	userRepository "rest-api-blueprint/internal/features/user/repository"
	userService "rest-api-blueprint/internal/features/user/service"
)

// ============================================================
// BUILD CONTROLLERS – WIRES ALL FEATURE DEPENDENCIES
// ============================================================

// BuildControllers creates all repositories, services, and controllers.
// It returns the combined server that implements the generated interface.
func BuildControllers(cfg *config.Config, emailSender email.Sender) *CombinedServer {
	// ============================================================
	// HEALTH FEATURE
	// ============================================================
	healthRepo := healthRepository.NewRepository(database.DB, cache.Client)
	healthSvc := healthService.NewService(healthRepo)
	healthCtrl := healthController.NewHealthController(healthSvc)

	// ============================================================
	// AUTH FEATURE
	// ============================================================
	authRepo := authRepository.NewRepository(database.DB)
	authSvc := authService.NewService(authRepo, cfg, cache.Client, emailSender)
	authCtrl := authController.NewAuthController(authSvc)

	// ============================================================
	// USER FEATURE
	// ============================================================
	userRepo := userRepository.NewRepository(database.DB)
	userSvc := userService.NewService(userRepo)
	userCtrl := userController.NewUserController(userSvc)

	// ============================================================
	// ADMIN FEATURE
	// ============================================================
	adminRepo := adminRepository.NewRepository(database.DB)
	adminSvc := adminService.NewService(adminRepo)
	adminCtrl := adminController.NewAdminController(adminSvc)

	// ============================================================
	// COMBINED SERVER (implements generated ServerInterface)
	// ============================================================
	return &CombinedServer{
		HealthController: healthCtrl,
		AuthController:   authCtrl,
		UserController:   userCtrl,
		AdminController:  adminCtrl,
	}
}
