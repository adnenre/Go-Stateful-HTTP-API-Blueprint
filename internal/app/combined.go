package app

import (
	admin "rest-api-blueprint/internal/features/admin/controller"
	auth "rest-api-blueprint/internal/features/auth/controller"
	health "rest-api-blueprint/internal/features/health/controller"
	user "rest-api-blueprint/internal/features/user/controller"
)

// CombinedServer implements the generated ServerInterface by embedding all feature controllers.
type CombinedServer struct {
	*health.HealthController
	*auth.AuthController
	*user.UserController
	*admin.AdminController
}
