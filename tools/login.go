package tools

import (
	"context"
	"fmt"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/michMartineau/ms-todo-mcp/types"

	"github.com/michMartineau/ms-todo-mcp/auth"
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
		if tm.PendingDeviceCode == nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{mcp.TextContent{Type: "text", Text: "No login in progress. Call 'login' first."}},
				IsError: true,
			}, nil
		}

		tokenResp, err := tm.PollForToken(ctx, tm.PendingDeviceCode)
		if err != nil {
			tm.PendingDeviceCode = nil
			return &mcp.CallToolResult{
				Content: []mcp.Content{mcp.TextContent{Type: "text", Text: fmt.Sprintf("Authentication failed: %s", err)}},
				IsError: true,
			}, nil
		}

		tokens := &types.StoredTokens{
			AccessToken:  tokenResp.AccessToken,
			RefreshToken: tokenResp.RefreshToken,
			ExpiresAt:    time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second),
		}
		if err := tm.SaveTokens(tokens); err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{mcp.TextContent{Type: "text", Text: fmt.Sprintf("Failed to save tokens: %s", err)}},
				IsError: true,
			}, nil
		}

		tm.PendingDeviceCode = nil
		return &mcp.CallToolResult{
			Content: []mcp.Content{mcp.TextContent{Type: "text", Text: "Authentication successful! You can now use Microsoft To-Do tools."}},
		}, nil
	}

	return server.ServerTool{Tool: tool, Handler: handler}
}
