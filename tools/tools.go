// Package tools registers MCP tools for Microsoft To-Do operations.
package tools

import (
	"github.com/mark3labs/mcp-go/server"

	"github.com/michMartineau/ms-todo-mcp/auth"
	"github.com/michMartineau/ms-todo-mcp/client"
)

// Register adds all Microsoft To-Do tools to the MCP server.
func Register(srv *server.MCPServer, graphClient *client.GraphClient, tokenManager *auth.TokenManager) {
	srv.AddTools(
		loginTool(tokenManager),
		loginCompleteTool(tokenManager),
		listTodoListsTool(graphClient),
		listTasksTool(graphClient),
		createTaskTool(graphClient),
		completeTaskTool(graphClient),
		deleteTaskTool(graphClient),
	)
}