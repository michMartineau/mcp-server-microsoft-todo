package tools

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/michMartineau/mcp-server-microsoft-todo/client"
)

func deleteTaskTool(graphClient *client.GraphClient) server.ServerTool {
	tool := mcp.NewTool(
		"delete_task",
		mcp.WithDescription("Delete a task from a Microsoft To-Do task list"),
		mcp.WithString(
			"list_id",
			mcp.Description("The ID of the task list containing the task"),
			mcp.Required(),
		),
		mcp.WithString(
			"task_id",
			mcp.Description("The ID of the task to delete"),
			mcp.Required(),
		),
	)

	handler := func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		listID := request.GetString("list_id", "")
		taskID := request.GetString("task_id", "")
		if listID == "" || taskID == "" {
			return &mcp.CallToolResult{
				Content: []mcp.Content{mcp.TextContent{Type: "text", Text: "Error: list_id and task_id are required"}},
				IsError: true,
			}, nil
		}

		err := graphClient.DeleteTask(ctx, listID, taskID)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{mcp.TextContent{Type: "text", Text: fmt.Sprintf("Error: %s", err)}},
				IsError: true,
			}, nil
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{mcp.TextContent{Type: "text", Text: "Task deleted successfully."}},
		}, nil
	}

	return server.ServerTool{Tool: tool, Handler: handler}
}