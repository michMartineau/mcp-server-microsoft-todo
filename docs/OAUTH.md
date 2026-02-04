# OAuth2 Device Code Flow - Explained

## Why Device Code Flow?

MCP servers run as CLI tools without a browser. The **device code flow** is designed for exactly this:

1. Server shows you a code and URL
2. You open the URL in any browser and enter the code
3. Server polls Microsoft until you complete authentication
4. Microsoft returns tokens to the server

```
┌─────────────┐                    ┌─────────────┐                    ┌─────────────┐
│ MCP Server  │                    │  Microsoft  │                    │    User     │
│ (terminal)  │                    │   (Azure)   │                    │  (browser)  │
└──────┬──────┘                    └──────┬──────┘                    └──────┬──────┘
       │                                  │                                  │
       │ 1. POST /devicecode              │                                  │
       │  (client_id, scope)              │                                  │
       │─────────────────────────────────►│                                  │
       │                                  │                                  │
       │ 2. Returns device_code,          │                                  │
       │    user_code, verification_uri   │                                  │
       │◄─────────────────────────────────│                                  │
       │                                  │                                  │
       │ 3. Display to user:              │                                  │
       │    "Go to aka.ms/devicelogin     │                                  │
       │     Enter code: ABCD-EFGH"       │                                  │
       │─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─►│
       │                                  │                                  │
       │                                  │  4. User visits URL,             │
       │                                  │     enters code, logs in         │
       │                                  │◄─────────────────────────────────│
       │                                  │                                  │
       │ 5. Poll: POST /token             │                                  │
       │    (device_code)                 │                                  │
       │─────────────────────────────────►│                                  │
       │                                  │                                  │
       │ 6. "authorization_pending"       │                                  │
       │◄─────────────────────────────────│                                  │
       │                                  │                                  │
       │ ... (repeat polling) ...         │                                  │
       │                                  │                                  │
       │ 7. Returns access_token,         │                                  │
       │    refresh_token                 │                                  │
       │◄─────────────────────────────────│                                  │
       │                                  │                                  │
       │ 8. Save tokens to disk           │                                  │
       │                                  │                                  │
```

## The Two Tokens

Microsoft returns two tokens:

### Access Token
- **What:** A JWT (JSON Web Token) that proves your identity
- **Lifespan:** ~1 hour
- **Usage:** Sent with every API request in the `Authorization` header
- **Example:** `Authorization: Bearer eyJ0eXAiOiJKV1QiLCJhbGciOiJS...`

### Refresh Token
- **What:** A long-lived secret that can get new access tokens
- **Lifespan:** 90 days (extended each time you use it)
- **Usage:** Exchange for a new access token when the current one expires
- **Security:** Never sent to the Graph API, only to the token endpoint

## Token Refresh Flow (YOUR TASK!)

When the access token expires, use the refresh token to get a new one **without user interaction**:

```
┌─────────────┐                              ┌─────────────┐
│ MCP Server  │                              │  Microsoft  │
└──────┬──────┘                              └──────┬──────┘
       │                                            │
       │  GetValidToken() called                    │
       │  ├── Load tokens from disk                 │
       │  ├── Check: is token expired?              │
       │  │   └── YES, token expired                │
       │  │                                         │
       │  │  POST /token                            │
       │  │  Content-Type: application/x-www-form-urlencoded
       │  │  Body:                                  │
       │  │    grant_type=refresh_token             │
       │  │    client_id=<your-client-id>           │
       │  │    refresh_token=<stored-refresh-token> │
       │  │    scope=Tasks.ReadWrite offline_access │
       │  │─────────────────────────────────────────►
       │  │                                         │
       │  │  Response: {                            │
       │  │    "access_token": "new-token...",      │
       │  │    "refresh_token": "new-refresh...",   │
       │  │    "expires_in": 3600                   │
       │  │  }                                      │
       │  │◄─────────────────────────────────────────
       │  │                                         │
       │  ├── Save new tokens to disk               │
       │  └── Return new access_token               │
       │                                            │
```

## HTTP Request Details

### Endpoint
```
POST https://login.microsoftonline.com/consumers/oauth2/v2.0/token
```

### Headers
```
Content-Type: application/x-www-form-urlencoded
```

### Body (form-encoded, not JSON!)
```
grant_type=refresh_token
client_id=your-azure-app-client-id
refresh_token=the-stored-refresh-token
scope=Tasks.ReadWrite offline_access
```

### Success Response (200 OK)
```json
{
  "access_token": "eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiIs...", # gitleaks:allow
  "refresh_token": "0.AXYA...",
  "token_type": "Bearer",
  "expires_in": 3599,
  "scope": "Tasks.ReadWrite"
}
```

### Error Response (400 Bad Request)
```json
{
  "error": "invalid_grant",
  "error_description": "AADSTS700082: The refresh token has expired..."
}
```

## Go Code Pattern

Here's the pattern used elsewhere in the codebase (see `tryGetToken` function):

```go
// 1. Build form data
data := url.Values{
    "grant_type":    {"refresh_token"},
    "client_id":     {tm.clientID},
    "refresh_token": {tokens.RefreshToken},
    "scope":         {scopes},
}

// 2. Create request
req, err := http.NewRequestWithContext(ctx, "POST", tokenURL, strings.NewReader(data.Encode()))
if err != nil {
    return "", fmt.Errorf("creating request: %w", err)
}
req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

// 3. Send request
resp, err := tm.httpClient.Do(req)
if err != nil {
    return "", fmt.Errorf("sending request: %w", err)
}
defer resp.Body.Close()

// 4. Read response
body, err := io.ReadAll(resp.Body)
if err != nil {
    return "", fmt.Errorf("reading response: %w", err)
}

// 5. Check for errors
if resp.StatusCode != http.StatusOK {
    return "", fmt.Errorf("token refresh failed: %s - please run login again", body)
}

// 6. Parse response
var tokenResp types.TokenResponse
if err := json.Unmarshal(body, &tokenResp); err != nil {
    return "", fmt.Errorf("parsing response: %w", err)
}

// 7. Save new tokens
newTokens := &types.StoredTokens{
    AccessToken:  tokenResp.AccessToken,
    RefreshToken: tokenResp.RefreshToken,
    ExpiresAt:    time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second),
}
if err := tm.SaveTokens(newTokens); err != nil {
    return "", fmt.Errorf("saving tokens: %w", err)
}

// 8. Return new access token
return tokenResp.AccessToken, nil
```

## Common Errors and Solutions

| Error | Cause | Solution |
|-------|-------|----------|
| `invalid_grant` | Refresh token expired or revoked | User must login again |
| `invalid_client` | Wrong client_id | Check MS_TODO_CLIENT_ID env var |
| `invalid_scope` | Scope not configured in Azure | Add Tasks.ReadWrite permission in Azure portal |

## Testing Your Implementation

1. **First login:**
   ```bash
   # Build and run login
   go build -o mcp-server-microsoft-todo
   ./mcp-server-microsoft-todo login
   # Follow the browser instructions
   ```

2. **Check tokens saved:**
   ```bash
   cat ~/.config/mcp-server-microsoft-todo/tokens.json
   ```

3. **Force token expiry (for testing):**
   ```bash
   # Edit tokens.json, change expires_at to a past date
   ```

4. **Run a command that needs a token:**
   ```bash
   ./mcp-server-microsoft-todo list-lists
   # Should trigger refresh if token expired
   ```
