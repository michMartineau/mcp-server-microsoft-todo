# mcp-server-microsoft-todo

A Vibecoded mcp server for Microsoft TODO written with claude code ("outputStyle": "Learning").

## Features

- **List your task lists** — Browse all your Microsoft To-Do lists
- **List tasks** — View tasks within any list, with details like due dates, importance, and status
- **Create tasks** — Add new tasks with titles, descriptions, due dates, and importance levels
- **Complete tasks** — Mark tasks as completed
- **Delete tasks** — Remove tasks you no longer need

## Prerequisites

- [Go](https://go.dev/) 1.23 or later
- A Microsoft account (personal or work/school)
- An Azure app registration (see [Setup](#azure-app-registration))

## Azure App Registration

1. Go to the [Azure Portal](https://portal.azure.com/) → **Microsoft Entra ID** → **App registrations**
2. Click **New registration**
   - **Name:** `MS Todo MCP` (or any name you prefer)
   - **Supported account types:** "Personal Microsoft accounts only" (or include organizational if needed)
   - **Redirect URI:** Leave blank (not needed for device code flow)
3. Click **Register**
4. On the app overview page, copy the **Application (client) ID** — you'll need this later
5. Go to **Authentication** → Under **Advanced settings**, set **Allow public client flows** to **Yes** → Click **Save**
6. Go to **API permissions** → **Add a permission** → **Microsoft Graph** → **Delegated permissions** → Search for `Tasks.ReadWrite` → **Add permissions**

## Installation

```bash
git clone https://github.com/michMartineau/ms-todo-mcp.git
cd ms-todo-mcp
go build -o ms-todo-mcp
```

## Configuration

### Claude Desktop

Add the following to your Claude Desktop configuration file:

- **macOS:** `~/Library/Application Support/Claude/claude_desktop_config.json`
- **Windows:** `%APPDATA%\Claude\claude_desktop_config.json`

```json
{
  "mcpServers": {
    "microsoft-todo": {
      "command": "/absolute/path/to/ms-todo-mcp",
      "env": {
        "MS_TODO_CLIENT_ID": "your-azure-client-id-here"
      }
    }
  }
}
```

Replace `/absolute/path/to/ms-todo-mcp` with the actual path to the compiled binary and `your-azure-client-id-here` with the client ID from your Azure app registration.

### Authentication

On first use, the server will initiate the **device code flow**:

1. Claude will display a URL and a one-time code
2. Open the URL in your browser and enter the code
3. Sign in with your Microsoft account and grant permissions
4. The server stores tokens locally at `~/.config/ms-todo-mcp/tokens.json`

Tokens refresh automatically — you should only need to authenticate once.

## Available Tools

| Tool | Description |
|------|-------------|
| `list_todo_lists` | List all your Microsoft To-Do task lists |
| `list_tasks` | List tasks in a specific task list |
| `create_task` | Create a new task in a list |
| `complete_task` | Mark a task as completed |
| `delete_task` | Delete a task from a list |

## Architecture

```
Claude ←→ MCP Protocol (stdio) ←→ ms-todo-mcp ←→ Microsoft Graph API
```

The server communicates with Claude over stdin/stdout using the [Model Context Protocol](https://modelcontextprotocol.io/), and calls the [Microsoft Graph API](https://learn.microsoft.com/en-us/graph/api/resources/todo-overview) to manage To-Do tasks.

Built with [mcp-go](https://github.com/mark3labs/mcp-go).

## Development

```bash
# Install dependencies
go mod tidy

# Build
go build -o ms-todo-mcp

# Run directly (for testing)
MS_TODO_CLIENT_ID=your-client-id ./ms-todo-mcp
```

See [docs/DESIGN.md](docs/DESIGN.md) for architecture details and [docs/OAUTH.md](docs/OAUTH.md) for the full OAuth2 flow documentation.

## Security

- Tokens are stored with restricted permissions (`0600`) at `~/.config/ms-todo-mcp/tokens.json`
- No client secret is required (public client using device code flow)
- Only the `Tasks.ReadWrite` scope is requested — the server cannot access mail, calendar, or other data

## License

MIT