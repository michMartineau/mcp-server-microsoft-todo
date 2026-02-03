package main

import (
	"fmt"
	"log"
	"os"

	"github.com/mark3labs/mcp-go/server"

	"github.com/michMartineau/ms-todo-mcp/auth"
	"github.com/michMartineau/ms-todo-mcp/client"
	"github.com/michMartineau/ms-todo-mcp/tools"
)

func main() {
	clientID := os.Getenv("MS_TODO_CLIENT_ID")
	if clientID == "" {
		fmt.Fprintln(os.Stderr, "Error: MS_TODO_CLIENT_ID environment variable is required")
		os.Exit(1)
	}

	tokenManager, err := auth.NewTokenManager(clientID)
	if err != nil {
		log.Fatalf("Failed to create token manager: %v", err)
	}

	graphClient := client.NewGraphClient(tokenManager)

	mcpServer := server.NewMCPServer(
		"microsoft-todo",
		"0.1.0",
	)

	tools.Register(mcpServer, graphClient, tokenManager)

	if err := server.ServeStdio(mcpServer); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}