// Package client provides an HTTP client for the Microsoft Graph API.
package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/michMartineau/ms-todo-mcp/auth"
	"github.com/michMartineau/ms-todo-mcp/types"
)

const baseURL = "https://graph.microsoft.com/v1.0"

// GraphClient makes authenticated requests to the Microsoft Graph API.
type GraphClient struct {
	tokenManager *auth.TokenManager
	httpClient   *http.Client
}

// NewGraphClient creates a new Graph API client.
func NewGraphClient(tm *auth.TokenManager) *GraphClient {
	return &GraphClient{
		tokenManager: tm,
		httpClient:   &http.Client{Timeout: 30 * time.Second},
	}
}

// doRequest performs an authenticated HTTP request and returns the response body.
func (c *GraphClient) doRequest(ctx context.Context, method, url string, body io.Reader) ([]byte, error) {
	token, err := c.tokenManager.GetValidToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting auth token: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var graphErr types.GraphError
		if json.Unmarshal(respBody, &graphErr) == nil && graphErr.Error.Message != "" {
			return nil, fmt.Errorf("Graph API error (%s): %s", graphErr.Error.Code, graphErr.Error.Message)
		}
		return nil, fmt.Errorf("Graph API returned status %d: %s", resp.StatusCode, respBody)
	}

	return respBody, nil
}

// ListTodoLists returns all the user's To-Do task lists.
func (c *GraphClient) ListTodoLists(ctx context.Context) ([]types.TodoTaskList, error) {
	body, err := c.doRequest(ctx, "GET", baseURL+"/me/todo/lists", nil)
	if err != nil {
		return nil, err
	}

	var resp types.TodoTaskListsResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("parsing task lists: %w", err)
	}
	return resp.Value, nil
}

// ListTasks returns all tasks in a specific task list.
func (c *GraphClient) ListTasks(ctx context.Context, listID string) ([]types.TodoTask, error) {
	url := fmt.Sprintf("%s/me/todo/lists/%s/tasks", baseURL, listID)
	body, err := c.doRequest(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	var resp types.TodoTasksResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("parsing tasks: %w", err)
	}
	return resp.Value, nil
}

// CreateTask creates a new task in the specified list and returns the created task.
func (c *GraphClient) CreateTask(ctx context.Context, listID string, title string, body string, importance string, dueDate string) (*types.TodoTask, error) {
	url := fmt.Sprintf("%s/me/todo/lists/%s/tasks", baseURL, listID)

	update := map[string]string{"title": title, "importance": importance, "dueDate": dueDate}
	payload, err := json.Marshal(update)
	respBody, err := c.doRequest(ctx, "POST", url, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	var task types.TodoTask
	if err := json.Unmarshal(respBody, &task); err != nil {
		return nil, fmt.Errorf("parsing updated task: %w", err)
	}
	return &task, nil
}

// CompleteTask marks a task as completed.
func (c *GraphClient) CompleteTask(ctx context.Context, listID, taskID string) (*types.TodoTask, error) {
	url := fmt.Sprintf("%s/me/todo/lists/%s/tasks/%s", baseURL, listID, taskID)

	update := map[string]string{"status": "completed"}
	payload, err := json.Marshal(update)
	if err != nil {
		return nil, fmt.Errorf("marshaling update: %w", err)
	}

	respBody, err := c.doRequest(ctx, "PATCH", url, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}

	var task types.TodoTask
	if err := json.Unmarshal(respBody, &task); err != nil {
		return nil, fmt.Errorf("parsing updated task: %w", err)
	}
	return &task, nil
}

// DeleteTask removes a task from a list.
func (c *GraphClient) DeleteTask(ctx context.Context, listID, taskID string) error {
	url := fmt.Sprintf("%s/me/todo/lists/%s/tasks/%s", baseURL, listID, taskID)
	_, err := c.doRequest(ctx, "DELETE", url, nil)
	return err
}
