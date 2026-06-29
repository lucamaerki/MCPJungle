package model

import (
	"time"

	"github.com/mcpjungle/mcpjungle/pkg/types"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// UpstreamOAuthPendingSession stores an in-progress OAuth authorization flow for
// an upstream MCP server registration.
type UpstreamOAuthPendingSession struct {
	gorm.Model

	SessionID string `json:"session_id" gorm:"uniqueIndex;not null"`

	ServerName string                   `json:"server_name" gorm:"index;not null"`
	Transport  types.McpServerTransport `json:"transport" gorm:"type:varchar(30);not null"`

	// ServerInput stores the original RegisterServerInput payload so registration
	// can be resumed after the OAuth callback completes.
	ServerInput datatypes.JSON `json:"server_input" gorm:"type:jsonb;not null"`

	Force bool `json:"force" gorm:"not null;default:false"`

	RedirectURI  string         `json:"redirect_uri"`
	ClientID     string         `json:"client_id"`
	ClientSecret string         `json:"client_secret"`
	Scopes       datatypes.JSON `json:"scopes" gorm:"type:jsonb"`

	State        string    `json:"state" gorm:"not null"`
	CodeVerifier string    `json:"code_verifier" gorm:"not null"`
	ExpiresAt    time.Time `json:"expires_at" gorm:"index;not null"`

	InitiatedBy string `json:"initiated_by"`
}

// UpstreamOAuthToken stores OAuth credentials for a registered upstream MCP server.
// When UserID is nil, the record is a shared gateway-level token used as fallback.
// When UserID is set, the record is a per-user token that takes priority over the gateway token.
type UpstreamOAuthToken struct {
	gorm.Model

	ServerName string                   `json:"server_name" gorm:"not null;uniqueIndex:idx_oauth_token_server_user"`
	UserID     *uint                    `json:"user_id,omitempty" gorm:"uniqueIndex:idx_oauth_token_server_user"`
	Transport  types.McpServerTransport `json:"transport" gorm:"type:varchar(30);not null"`

	ClientID     string         `json:"client_id"`
	ClientSecret string         `json:"client_secret"`
	RedirectURI  string         `json:"redirect_uri"`
	Scopes       datatypes.JSON `json:"scopes" gorm:"type:jsonb"`

	AccessToken  string    `json:"access_token"`
	TokenType    string    `json:"token_type"`
	RefreshToken string    `json:"refresh_token"`
	Scope        string    `json:"scope"`
	ExpiresAt    time.Time `json:"expires_at"`
}

// UserUpstreamOAuthPendingSession stores an in-progress per-user OAuth authorization
// flow for an upstream MCP server. Separate from UpstreamOAuthPendingSession which
// handles admin-initiated gateway-level registration flows.
type UserUpstreamOAuthPendingSession struct {
	gorm.Model

	SessionID  string `json:"session_id" gorm:"uniqueIndex;not null"`
	ServerName string `json:"server_name" gorm:"index;not null"`
	UserID     uint   `json:"user_id" gorm:"index;not null"`

	// Copied from the gateway-level UpstreamOAuthToken at session start
	RedirectURI  string         `json:"redirect_uri"`
	ClientID     string         `json:"client_id"`
	ClientSecret string         `json:"client_secret"`
	Scopes       datatypes.JSON `json:"scopes" gorm:"type:jsonb"`

	State        string    `json:"state" gorm:"not null"`
	CodeVerifier string    `json:"code_verifier" gorm:"not null"`
	ExpiresAt    time.Time `json:"expires_at" gorm:"index;not null"`
}
