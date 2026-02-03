package tools

import (
	"context"
	"fmt"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/michMartineau/ms-todo-mcp/auth"
	"github.com/michMartineau/ms-todo-mcp/types"
)

func loginTool(tm *auth.TokenManager) server.ServerTool {
	tool := mcp.NewTool(
		"login",
		mcp.WithDescription("Start Microsoft authentication. Returns a URL and code for the user to complete sign-in."),
	)

	handler := func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		deviceCode, err := tm.RequestDeviceCode(ctx)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{mcp.TextContent{Type: "text", Text: fmt.Sprintf("Error requesting device code: %s", err)}},
				IsError: true,
			}, nil
		}

		tm.PendingDeviceCode = deviceCode

		msg := fmt.Sprintf(
			"Please visit: %s\nEnter code: %s\n\nOnce you have entered the code, call the 'login_complete' tool to finish authentication.",
			deviceCode.VerificationURI,
			deviceCode.UserCode,
		)

		return &mcp.CallToolResult{
			Content: []mcp.Content{mcp.TextContent{Type: "text", Text: msg}},
		}, nil
	}

	return server.ServerTool{Tool: tool, Handler: handler}
}

func loginCompleteTool(tm *auth.TokenManager) server.ServerTool {
	tool := mcp.NewTool(
		"login_complete",
		mcp.WithDescription("Complete Microsoft authentication after the user has entered the device code."),
	)

	handler := func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// TODO(human): Implement the login completion logic.
		// Use tm.PendingDeviceCode, tm.PollForToken, and tm.SaveTokens.
		// Clear tm.PendingDeviceCode when done.
		// Return a success or error message as a CallToolResult.
		return nil, nil
	}

	return server.ServerTool{Tool: tool, Handler: handler}
}
