# MS Todo MCP Server - Design Document

## Overview

This MCP (Model Context Protocol) server allows Claude to interact with Microsoft To-Do on your behalf. When you ask Claude to "add a task to my todo list" or "show me my tasks", Claude calls this server which then communicates with Microsoft's Graph API.

```
┌─────────┐      MCP Protocol       ┌──────────────┐      Graph API      ┌─────────────┐
│ Claude  │ ◄───────────────────► │ This Server  │ ◄─────────────────► │ Microsoft   │
│         │   (JSON-RPC over       │ (Go binary)  │   (HTTPS REST)      │ To-Do       │
└─────────┘    stdio)              └──────────────┘                     └─────────────┘
```

## Project Structure

```
ms-todo/
├── main.go              # Entry point - initializes MCP server
├── go.mod               # Go module definition
│
├── auth/
│   └── oauth.go         # OAuth2 authentication (device code flow + token refresh)
│
├── client/
│   └── graph.go         # Microsoft Graph API HTTP client
│
├── tools/
│   ├── tools.go         # Tool registration with MCP server
│   ├── list_todo_lists.go
│   ├── list_tasks.go
│   ├── create_task.go
│   ├── complete_task.go
│   └── delete_task.go
│
├── types/
│   └── types.go         # Data structures (API responses, tokens)
│
└── docs/
    ├── DESIGN.md        # This file
    └── OAUTH.md         # OAuth2 flow explanation
```

## Component Responsibilities

### 1. `main.go` - Entry Point

```go
func main() {
    // 1. Read client ID from environment
    clientID := os.Getenv("MS_TODO_CLIENT_ID")

    // 2. Create token manager (handles OAuth)
    tokenManager := auth.NewTokenManager(clientID)

    // 3. Create Graph client (makes API calls)
    graphClient := client.NewGraphClient(tokenManager)

    // 4. Create MCP server and register tools
    server := mcp.NewServer()
    tools.RegisterAll(server, graphClient, tokenManager)

    // 5. Start serving (reads from stdin, writes to stdout)
    server.Serve()
}
```

### 2. `auth/oauth.go` - Authentication

Manages the OAuth2 lifecycle:

| Function | Purpose | When Called |
|----------|---------|-------------|
| `NewTokenManager()` | Creates manager, sets up token file path | Once at startup |
| `DeviceCodeLogin()` | First-time auth via browser | User runs "login" tool |
| `GetValidToken()` | Returns valid token, refreshes if expired | Before every API call |
| `LoadTokens()` | Reads tokens from disk | Internal helper |
| `SaveTokens()` | Writes tokens to disk | After login or refresh |
| `ClearTokens()` | Deletes token file | User runs "logout" tool |

### 3. `client/graph.go` - API Client

Makes HTTP requests to Microsoft Graph API:

```go
type GraphClient struct {
    tokenManager *auth.TokenManager
    httpClient   *http.Client
    baseURL      string  // "https://graph.microsoft.com/v1.0"
}

// All methods follow this pattern:
func (c *GraphClient) ListTasks(ctx context.Context, listID string) ([]types.TodoTask, error) {
    // 1. Get valid token
    token, err := c.tokenManager.GetValidToken(ctx)

    // 2. Make HTTP request with Authorization header
    req.Header.Set("Authorization", "Bearer " + token)

    // 3. Parse response into types.TodoTask structs
    return tasks, nil
}
```

### 4. `tools/*.go` - MCP Tool Handlers

Each tool is a function Claude can call:

```go
// Example: list_tasks tool
var ListTasksTool = mcp.Tool{
    Name:        "list_tasks",
    Description: "List tasks in a Microsoft To-Do list",
    InputSchema: /* JSON Schema for parameters */,
    Handler: func(ctx context.Context, args map[string]any) (any, error) {
        listID := args["list_id"].(string)
        return graphClient.ListTasks(ctx, listID)
    },
}
```

## Data Flow Example

When you ask Claude: *"Show me my tasks in the Groceries list"*

```
1. Claude decides to call "list_tasks" tool with list_id="abc123"
         │
         ▼
2. MCP Server receives JSON-RPC request via stdin
         │
         ▼
3. tools/list_tasks.go handler is invoked
         │
         ▼
4. graphClient.ListTasks(ctx, "abc123") is called
         │
         ▼
5. tokenManager.GetValidToken(ctx) is called
         │
         ├── Token valid? Return it
         │
         └── Token expired? ──► Refresh using refresh_token (YOUR TASK!)
                                     │
                                     ▼
                              POST to Microsoft token endpoint
                                     │
                                     ▼
                              Save new tokens, return access_token
         │
         ▼
6. HTTP GET https://graph.microsoft.com/v1.0/me/todo/lists/abc123/tasks
   Headers: Authorization: Bearer <access_token>
         │
         ▼
7. Parse JSON response into []types.TodoTask
         │
         ▼
8. Return result to Claude via MCP protocol
         │
         ▼
9. Claude formats and presents the tasks to you
```

## Error Handling Strategy

| Error Type | Handling |
|------------|----------|
| No tokens stored | Return "Please run login first" |
| Token refresh fails | Return "Session expired, please login again" |
| API rate limited | Return error with retry suggestion |
| Network error | Return error with details |
| Invalid list ID | Return "List not found" |

## Security Considerations

1. **Tokens stored locally** in `~/.config/mcp-server-microsoft-todo/tokens.json` with `0600` permissions
2. **No client secret** - uses public client (device code flow)
3. **Minimal scopes** - only `Tasks.ReadWrite` and `offline_access`
4. **Refresh tokens** allow long-term access without re-authenticating
