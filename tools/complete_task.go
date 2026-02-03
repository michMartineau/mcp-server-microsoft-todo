package tools

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/michMartineau/ms-todo-mcp/client"
)

func completeTaskTool(graphClient *client.GraphClient) server.ServerTool {
	tool := mcp.NewTool(
		"complete_task",
		mcp.WithDescription("Mark a Microsoft To-Do task as completed"),
		mcp.WithString(
			"list_id",
			mcp.Description("The ID of the task list containing the task"),
			mcp.Required(),
		),
		mcp.WithString(
			"task_id",
			mcp.Description("The ID of the task to complete"),
			mcp.Required(),
		),
	)

	handler := func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		listID, _ := request.Params.Arguments["list_id"].(string)
		taskID, _ := request.Params.Arguments["task_id"].(string)
		if listID == "" || taskID == "" {
			return &mcp.CallToolResult{
				Content: []mcp.Content{mcp.TextContent{Type: "text", Text: "Error: list_id and task_id are required"}},
				IsError: true,
			}, nil
		}

		task, err := graphClient.CompleteTask(ctx, listID, taskID)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{mcp.TextContent{Type: "text", Text: fmt.Sprintf("Error: %s", err)}},
				IsError: true,
			}, nil
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{mcp.TextContent{Type: "text", Text: fmt.Sprintf("Task \"%s\" marked as completed.", task.Title)}},
		}, nil
	}

	return server.ServerTool{Tool: tool, Handler: handler}
}