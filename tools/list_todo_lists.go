package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/michMartineau/mcp-server-microsoft-todo/client"
)

func listTodoListsTool(graphClient *client.GraphClient) server.ServerTool {
	tool := mcp.NewTool(
		"list_todo_lists",
		mcp.WithDescription("List Microsoft To-Do task lists. Optionally filter by name."),
		mcp.WithString(
			"name",
			mcp.Description("Optional list name to filter by (exact match)"),
		),
	)

	handler := func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		name := request.GetString("name", "")
		filter := ""
		if name != "" {
			filter = "displayName eq '" + name + "'"
		}
		lists, err := graphClient.ListTodoLists(ctx, filter)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{mcp.TextContent{Type: "text", Text: fmt.Sprintf("Error: %s", err)}},
				IsError: true,
			}, nil
		}

		if len(lists) == 0 {
			return &mcp.CallToolResult{
				Content: []mcp.Content{mcp.TextContent{Type: "text", Text: "No task lists found."}},
			}, nil
		}

		var sb strings.Builder
		for _, list := range lists {
			sb.WriteString(fmt.Sprintf("- **%s** (ID: `%s`)\n", list.DisplayName, list.ID))
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{mcp.TextContent{Type: "text", Text: sb.String()}},
		}, nil
	}

	return server.ServerTool{Tool: tool, Handler: handler}
}
