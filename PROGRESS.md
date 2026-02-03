# MS Todo MCP Server - Implementation Progress

## Current Status
**Date:** 2025-01-27
**Phase:** Step 4 of 9 - Token refresh logic (waiting for human contribution)

## Completed Files

| File | Status | Description |
|------|--------|-------------|
| `go.mod` | ✅ Done | Go module definition |
| `types/types.go` | ✅ Done | All data structures for Graph API |
| `auth/oauth.go` | 🔶 Partial | Device code flow complete, **token refresh needs implementation** |

## Pending Files

| File | Status |
|------|--------|
| `client/graph.go` | ⏳ Not started |
| `tools/tools.go` | ⏳ Not started |
| `tools/list_todo_lists.go` | ⏳ Not started |
| `tools/list_tasks.go` | ⏳ Not started |
| `tools/create_task.go` | ⏳ Not started |
| `tools/complete_task.go` | ⏳ Not started |
| `tools/delete_task.go` | ⏳ Not started |
| `main.go` | ⏳ Not started |

## 🎯 Your Task: Implement Token Refresh

In `auth/oauth.go`, find the `GetValidToken` function and the `TODO(human)` comment.

**What to implement:** When the access token is expired, use the refresh token to get a new one.

**HTTP Request details:**
- **URL:** `tokenURL` (already defined as constant)
- **Method:** POST
- **Content-Type:** `application/x-www-form-urlencoded`
- **Body parameters:**
  - `grant_type`: `"refresh_token"`
  - `client_id`: `tm.clientID`
  - `refresh_token`: `tokens.RefreshToken`
  - `scope`: `scopes` (constant already defined)

**Expected behavior:**
1. Make the POST request
2. Parse response into `types.TokenResponse`
3. Create new `StoredTokens` and save with `tm.SaveTokens()`
4. Return the new access token

**If refresh fails:** Return an error telling user to re-authenticate.

## Next Steps After Token Refresh

Once you implement the token refresh:
1. Run `go mod tidy` to fetch dependencies
2. Run `go build` to verify compilation
3. Continue with `client/graph.go`

## Azure App Setup Reminder

Before testing, ensure you have:
1. Azure App Registration with Client ID
2. "Allow public client flows" enabled
3. `Tasks.ReadWrite` API permission added

Set client ID via environment variable: `MS_TODO_CLIENT_ID`
