// Package migrations provides database migration functionality for the MCPJungle application.
package migrations

import (
	"fmt"

	"github.com/mcpjungle/mcpjungle/internal/model"
	"gorm.io/gorm"
)

// Migrate performs the database migration for the application.
func Migrate(db *gorm.DB) error {
	if err := db.AutoMigrate(&model.McpServer{}); err != nil {
		return fmt.Errorf("auto‑migration failed for McpServer model: %v", err)
	}
	if err := db.AutoMigrate(&model.Tool{}); err != nil {
		return fmt.Errorf("auto‑migration failed for Tool model: %v", err)
	}
	if err := db.AutoMigrate(&model.ServerConfig{}); err != nil {
		return fmt.Errorf("auto‑migration failed for ServerConfig model: %v", err)
	}
	if err := db.AutoMigrate(&model.User{}); err != nil {
		return fmt.Errorf("auto‑migration failed for User model: %v", err)
	}
	if err := db.AutoMigrate(&model.McpClient{}); err != nil {
		return fmt.Errorf("auto‑migration failed for McpClient model: %v", err)
	}
	if err := db.AutoMigrate(&model.ToolGroup{}); err != nil {
		return fmt.Errorf("auto‑migration failed for ToolGroup model: %v", err)
	}
	if err := db.AutoMigrate(&model.Prompt{}); err != nil {
		return fmt.Errorf("auto‑migration failed for Prompt model: %v", err)
	}
	if err := db.AutoMigrate(&model.Resource{}); err != nil {
		return fmt.Errorf("auto‑migration failed for Resource model: %v", err)
	}
	if err := db.AutoMigrate(&model.UpstreamOAuthPendingSession{}); err != nil {
		return fmt.Errorf("auto-migration failed for UpstreamOAuthPendingSession model: %v", err)
	}
	if err := db.AutoMigrate(&model.UpstreamOAuthToken{}); err != nil {
		return fmt.Errorf("auto-migration failed for UpstreamOAuthToken model: %v", err)
	}
	if err := db.AutoMigrate(&model.UserUpstreamOAuthPendingSession{}); err != nil {
		return fmt.Errorf("auto-migration failed for UserUpstreamOAuthPendingSession model: %v", err)
	}

	// Migrate the single-column unique index on server_name to a composite index
	// (server_name, user_id) to support per-user tokens alongside gateway tokens.
	// COALESCE(user_id, 0) is used as a SQLite-compatible workaround since SQLite
	// does not treat two NULL values as distinct in a unique index.
	if err := migrateUpstreamOAuthTokenIndex(db); err != nil {
		return fmt.Errorf("failed to migrate upstream OAuth token index: %v", err)
	}

	return nil
}

// migrateUpstreamOAuthTokenIndex replaces the old single-column unique index on
// server_name with a composite index on (server_name, COALESCE(user_id, 0)).
// The operation is idempotent and safe to run on every startup.
func migrateUpstreamOAuthTokenIndex(db *gorm.DB) error {
	db.Exec("DROP INDEX IF EXISTS idx_upstream_oauth_tokens_server_name")
	db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS idx_oauth_token_server_user
		ON upstream_oauth_tokens(server_name, COALESCE(user_id, 0))
		WHERE deleted_at IS NULL`)
	return nil
}
