package mcp

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mcpjungle/mcpjungle/internal/model"
)

// ProxyToolFilter filters tools exposed by MCP proxy for enterprise mode based on client allow-list.
func ProxyToolFilter(ctx context.Context, tools []mcp.Tool) []mcp.Tool {
	serverMode, ok := ctx.Value("mode").(model.ServerMode)
	if !ok {
		// Missing/invalid mode in context: fail closed.
		return nil
	}
	if !model.IsEnterpriseMode(serverMode) {
		// In non-enterprise mode, there are no access restrictions, so return all tools
		return tools
	}

	// Branch 1: McpClient in context — filter tools by server allowlist.
	if c, ok := ctx.Value("client").(*model.McpClient); ok && c != nil {
		var filteredTools []mcp.Tool
		allowedServers := make(map[string]bool)
		for _, tool := range tools {
			serverName, _, _ := splitServerToolName(tool.Name)
			allowed, cached := allowedServers[serverName]
			if !cached {
				allowed = c.CheckHasServerAccess(serverName)
				allowedServers[serverName] = allowed
			}
			if allowed {
				filteredTools = append(filteredTools, tool)
			}
		}
		return filteredTools
	}

	// Branch 2: User in context — user sees all tools (per-tool ACL is enforced by
	// the upstream system via their own OAuth token, not by the gateway).
	if _, ok := ctx.Value("user").(*model.User); ok {
		return tools
	}

	// Enterprise mode requires an authenticated identity; fail closed if absent.
	return nil
}
