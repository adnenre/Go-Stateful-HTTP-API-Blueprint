package auth

import (
	"log/slog"
	"os"
	"rest-api-blueprint/internal/database"
	"rest-api-blueprint/internal/features/auth/model"
)

// Migrate creates or updates the users table.
func Migrate() {
	if err := database.DB.AutoMigrate(&model.User{}); err != nil {
		slog.Error("failed to migrate users table", "error", err)
		os.Exit(1)
	}
	slog.Info("users table migration completed")
}
