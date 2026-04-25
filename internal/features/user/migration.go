package user

import (
	"log/slog"
	"os"
	"rest-api-blueprint/internal/database"
	"rest-api-blueprint/internal/features/user/model"
)

// Migrate creates or updates the user_preferences table.
func Migrate() {
	if err := database.DB.AutoMigrate(&model.UserPreferences{}); err != nil {
		slog.Error("failed to migrate user preferences table", "error", err)
		os.Exit(1)
	}
	slog.Info("user preferences table migration completed")
}
