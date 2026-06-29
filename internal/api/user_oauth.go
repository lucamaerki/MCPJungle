package api

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mcpjungle/mcpjungle/internal/model"
	"github.com/mcpjungle/mcpjungle/pkg/apierrors"
)

// getAuthenticatedUser returns the User stored in the gin context, or nil if not present.
func getAuthenticatedUser(c *gin.Context) *model.User {
	if u, ok := c.Get("user"); ok {
		if user, ok := u.(*model.User); ok {
			return user
		}
	}
	return nil
}

// startUserOAuthHandler handles POST /api/v0/servers/:name/user-oauth/start.
// Initiates a per-user OAuth flow for an upstream MCP server.
// The server must already have a gateway-level OAuth token registered by an admin.
func (s *Server) startUserOAuthHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		serverName := c.Param("name")

		u := getAuthenticatedUser(c)
		if u == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
			return
		}

		var input struct {
			RedirectURI string `json:"redirect_uri"`
		}
		_ = c.ShouldBindJSON(&input)

		result, err := s.mcpService.StartUserUpstreamOAuth(c.Request.Context(), serverName, u.ID, input.RedirectURI)
		if err != nil {
			handleUserOAuthError(c, err)
			return
		}
		c.JSON(http.StatusAccepted, gin.H{
			"session_id":        result.SessionID,
			"authorization_url": result.AuthorizationURL,
			"expires_at":        result.ExpiresAt,
		})
	}
}

// getUserOAuthStatusHandler handles GET /api/v0/servers/:name/user-oauth/status.
// Returns whether the authenticated user has a linked OAuth token for the given server.
func (s *Server) getUserOAuthStatusHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		serverName := c.Param("name")

		u := getAuthenticatedUser(c)
		if u == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
			return
		}

		token, err := s.mcpService.GetUserUpstreamOAuthStatus(serverName, u.ID)
		if err != nil {
			if errors.Is(err, apierrors.ErrNotFound) {
				c.JSON(http.StatusOK, gin.H{"linked": false})
				return
			}
			handleUserOAuthError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"linked":     true,
			"expires_at": token.ExpiresAt,
			"scope":      token.Scope,
		})
	}
}

// revokeUserOAuthHandler handles DELETE /api/v0/servers/:name/user-oauth.
// Revokes the authenticated user's OAuth token for the given server.
func (s *Server) revokeUserOAuthHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		serverName := c.Param("name")

		u := getAuthenticatedUser(c)
		if u == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
			return
		}

		if err := s.mcpService.RevokeUserUpstreamOAuthToken(c.Request.Context(), serverName, u.ID); err != nil {
			handleUserOAuthError(c, err)
			return
		}
		c.Status(http.StatusNoContent)
	}
}

// completeUserOAuthSessionHandler handles POST /api/v0/user-oauth/sessions/:id/complete.
// Exchanges the OAuth callback code for a user-scoped token.
func (s *Server) completeUserOAuthSessionHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionID := c.Param("id")
		var input struct {
			Code  string `json:"code"`
			State string `json:"state"`
		}
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if err := s.mcpService.CompleteUserUpstreamOAuthSession(c.Request.Context(), sessionID, input.Code, input.State); err != nil {
			handleUserOAuthError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "linked"})
	}
}

// userOAuthCallbackHandler handles GET /oauth/user-callback.
// Public endpoint that receives the OAuth provider redirect after user authorization.
func (s *Server) userOAuthCallbackHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		code := c.Query("code")
		state := c.Query("state")
		oauthError := c.Query("error")
		errorDescription := c.Query("error_description")

		if oauthError != "" {
			renderDashboardOAuthHTML(c, http.StatusBadRequest,
				"Authorization failed", safeOAuthErrorMessage(oauthError, errorDescription))
			return
		}
		if code == "" || state == "" {
			renderDashboardOAuthHTML(c, http.StatusBadRequest,
				"Authorization failed", "Missing required OAuth callback parameters.")
			return
		}

		session, err := s.mcpService.GetUserUpstreamOAuthPendingSessionByState(c.Request.Context(), state)
		if err != nil {
			renderDashboardOAuthHTML(c, http.StatusBadRequest,
				"Authorization failed", "Session not found or has expired.")
			return
		}

		if err := s.mcpService.CompleteUserUpstreamOAuthSession(c.Request.Context(), session.SessionID, code, state); err != nil {
			renderDashboardOAuthHTML(c, http.StatusInternalServerError,
				"Authorization failed", safeOAuthCallbackError(err))
			return
		}

		renderDashboardOAuthHTML(c, http.StatusOK,
			"Authorization successful",
			"Your account is now linked. You can close this tab and return to MCPJungle.")
	}
}

func handleUserOAuthError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, apierrors.ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.Is(err, apierrors.ErrInvalidInput):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}
