package app

import (
	"net/http"
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

// Implement missing methods from generated interface to satisfy gen.ServerInterface.
// These forward to the embedded auth controller.

func (c *CombinedServer) RefreshToken(w http.ResponseWriter, r *http.Request) {
	c.AuthController.Refresh(w, r)
}

func (c *CombinedServer) Logout(w http.ResponseWriter, r *http.Request) {
	c.AuthController.Logout(w, r)
}

// Add GetSession method to satisfy the generated interface (operationId: getSession)
func (c *CombinedServer) GetSession(w http.ResponseWriter, r *http.Request) {
	c.AuthController.Session(w, r)
}
