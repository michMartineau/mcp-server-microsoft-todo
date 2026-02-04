package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/michMartineau/ms-todo-mcp/client"
	"github.com/michMartineau/ms-todo-mcp/types"
)

func listTasksTool(graphClient *client.GraphClient) server.ServerTool {
	tool := mcp.NewTool(
		"list_tasks",
		mcp.WithDescription("List all tasks in a Microsoft To-Do task list"),
		mcp.WithString(
			"list_id",
			mcp.Description("The ID of the task list. Use list_todo_lists to find it."),
			mcp.Required(),
		),
	)

	handler := func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		listID := request.GetString("list_id", "")
		if listID == "" {
			return &mcp.CallToolResult{
				Content: []mcp.Content{mcp.TextContent{Type: "text", Text: "Error: list_id is required"}},
				IsError: true,
			}, nil
		}

		tasks, err := graphClient.ListTasks(ctx, listID)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{mcp.TextContent{Type: "text", Text: fmt.Sprintf("Error: %s", err)}},
				IsError: true,
			}, nil
		}

		if len(tasks) == 0 {
			return &mcp.CallToolResult{
				Content: []mcp.Content{mcp.TextContent{Type: "text", Text: "No tasks found in this list."}},
			}, nil
		}

		var sb strings.Builder
		for _, task := range tasks {
			sb.WriteString(formatTask(task))
			sb.WriteString("\n")
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{mcp.TextContent{Type: "text", Text: sb.String()}},
		}, nil
	}

	return server.ServerTool{Tool: tool, Handler: handler}
}

func formatTask(task types.TodoTask) string {
	var sb strings.Builder

	checkbox := "☐"
	if task.Status == "completed" {
		checkbox = "☑"
	}
	sb.WriteString(fmt.Sprintf("%s **%s** (ID: `%s`)\n", checkbox, task.Title, task.ID))

	if task.Importance != "" && task.Importance != "normal" {
		sb.WriteString(fmt.Sprintf("  Importance: %s\n", task.Importance))
	}
	if task.Status != "" {
		sb.WriteString(fmt.Sprintf("  Status: %s\n", task.Status))
	}
	if task.DueDateTime != nil {
		sb.WriteString(fmt.Sprintf("  Due: %s\n", task.DueDateTime.DateTime))
	}
	if task.Body != nil && task.Body.Content != "" {
		sb.WriteString(fmt.Sprintf("  Notes: %s\n", task.Body.Content))
	}

	return sb.String()
}