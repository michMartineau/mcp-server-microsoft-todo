package tools

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/michMartineau/ms-todo-mcp/client"
)

func createTaskTool(graphClient *client.GraphClient) server.ServerTool {
	tool := mcp.NewTool(
		"create_task",
		mcp.WithDescription("Create a new task in a Microsoft To-Do task list"),
		mcp.WithString(
			"list_id",
			mcp.Description("The ID of the task list. Use list_todo_lists to find it."),
			mcp.Required(),
		),
		mcp.WithString(
			"title",
			mcp.Description("The title of the task"),
			mcp.Required(),
		),
		mcp.WithString(
			"body",
			mcp.Description("Optional description/notes for the task"),
		),
		mcp.WithString(
			"importance",
			mcp.Description("Task importance level"),
			mcp.Enum("low", "normal", "high"),
		),
		mcp.WithString(
			"due_date",
			mcp.Description("Optional due date in ISO 8601 format (e.g. 2025-12-31T17:00:00)"),
		),
	)

	// TODO(human): Implement the create_task handler
	handler := func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return nil, fmt.Errorf("not implemented")
	}

	return server.ServerTool{Tool: tool, Handler: handler}
}