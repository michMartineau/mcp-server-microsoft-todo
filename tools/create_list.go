package tools

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/michMartineau/mcp-server-microsoft-todo/client"
)

func createListTool(graphClient *client.GraphClient) server.ServerTool {
	tool := mcp.NewTool(
		"create_list",
		mcp.WithDescription("Create a new Microsoft To-Do task list"),
		mcp.WithString(
			"display_name",
			mcp.Description("The name of the new task list"),
			mcp.Required(),
		),
	)

	handler := func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		displayName := request.GetString("display_name", "")
		if displayName == "" {
			return &mcp.CallToolResult{
				Content: []mcp.Content{mcp.TextContent{Type: "text", Text: "Error: display_name is required"}},
				IsError: true,
			}, nil
		}
		list, err := graphClient.CreateList(ctx, displayName)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{mcp.TextContent{Type: "text", Text: fmt.Sprintf("Error: %s", err)}},
				IsError: true,
			}, nil
		}
		return &mcp.CallToolResult{
			Content: []mcp.Content{mcp.TextContent{Type: "text", Text: fmt.Sprintf("List \"%s\" created successfully. (ID: %s)", list.DisplayName, list.ID)}},
		}, nil
	}

	return server.ServerTool{Tool: tool, Handler: handler}
}
