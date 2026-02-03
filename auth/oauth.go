// Package auth handles OAuth2 authentication with Microsoft identity platform.
package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/michMartineau/ms-todo-mcp/types"
)

const (
	// Microsoft identity platform endpoints for consumer accounts
	deviceCodeURL = "https://login.microsoftonline.com/consumers/oauth2/v2.0/devicecode"
	tokenURL      = "https://login.microsoftonline.com/consumers/oauth2/v2.0/token"

	// Required scopes for Microsoft To-Do access
	scopes = "Tasks.ReadWrite offline_access"
)

// TokenManager handles OAuth token lifecycle.
type TokenManager struct {
	clientID         string
	tokensPath       string
	httpClient       *http.Client
	PendingDeviceCode *types.DeviceCodeResponse
}

// NewTokenManager creates a new token manager.
func NewTokenManager(clientID string) (*TokenManager, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return nil, fmt.Errorf("getting config dir: %w", err)
	}

	tokensDir := filepath.Join(configDir, "ms-todo-mcp")
	if err := os.MkdirAll(tokensDir, 0700); err != nil {
		return nil, fmt.Errorf("creating tokens dir: %w", err)
	}

	return &TokenManager{
		clientID:   clientID,
		tokensPath: filepath.Join(tokensDir, "tokens.json"),
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}, nil
}

// LoadTokens reads tokens from disk.
func (tm *TokenManager) LoadTokens() (*types.StoredTokens, error) {
	data, err := os.ReadFile(tm.tokensPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // No tokens stored yet
		}
		return nil, fmt.Errorf("reading tokens file: %w", err)
	}

	var tokens types.StoredTokens
	if err := json.Unmarshal(data, &tokens); err != nil {
		return nil, fmt.Errorf("parsing tokens: %w", err)
	}
	return &tokens, nil
}

// SaveTokens persists tokens to disk.
func (tm *TokenManager) SaveTokens(tokens *types.StoredTokens) error {
	data, err := json.MarshalIndent(tokens, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling tokens: %w", err)
	}

	if err := os.WriteFile(tm.tokensPath, data, 0600); err != nil {
		return fmt.Errorf("writing tokens file: %w", err)
	}
	return nil
}

// GetValidToken returns a valid access token, refreshing if necessary.
func (tm *TokenManager) GetValidToken(ctx context.Context) (string, error) {
	tokens, err := tm.LoadTokens()
	if err != nil {
		return "", err
	}
	if tokens == nil {
		return "", errors.New("not authenticated - run 'login' command first")
	}

	// Check if token is still valid (with 5 minute buffer)
	if time.Now().Add(5 * time.Minute).Before(tokens.ExpiresAt) {
		return tokens.AccessToken, nil
	}

	data := url.Values{
		"grant_type":    {"urn:ietf:params:oauth:grant_type=refresh_token"},
		"client_id":     {tm.clientID},
		"refresh_token": {tokens.RefreshToken},
		"scope":         {scopes},
	}

	req, err := http.NewRequestWithContext(ctx, "POST", tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := tm.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("reading refresh response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("token refresh failed (user may need to re-authenticate): %s", body)
	}

	var tokenResp types.TokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return "", fmt.Errorf("parsing refresh response: %w", err)
	}

	// Save the new tokens
	newTokens := &types.StoredTokens{
		AccessToken:  tokenResp.AccessToken,
		RefreshToken: tokenResp.RefreshToken,
		ExpiresAt:    time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second),
	}

	if err := tm.SaveTokens(newTokens); err != nil {
		return "", fmt.Errorf("saving refreshed tokens: %w", err)
	}

	return newTokens.AccessToken, nil
}

// DeviceCodeLogin initiates the device code authentication flow.
// It returns immediately with instructions for the user.
func (tm *TokenManager) DeviceCodeLogin(ctx context.Context) error {
	// Step 1: Request device code
	deviceCode, err := tm.RequestDeviceCode(ctx)
	if err != nil {
		return fmt.Errorf("requesting device code: %w", err)
	}

	// Display instructions to user
	fmt.Println("\n" + deviceCode.Message)
	fmt.Printf("\nWaiting for authentication (expires in %d seconds)...\n", deviceCode.ExpiresIn)

	// Step 2: Poll for token
	tokens, err := tm.PollForToken(ctx, deviceCode)
	if err != nil {
		return fmt.Errorf("polling for token: %w", err)
	}

	// Step 3: Save tokens
	storedTokens := &types.StoredTokens{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		ExpiresAt:    time.Now().Add(time.Duration(tokens.ExpiresIn) * time.Second),
	}

	if err := tm.SaveTokens(storedTokens); err != nil {
		return fmt.Errorf("saving tokens: %w", err)
	}

	fmt.Println("\n✓ Authentication successful! Tokens saved.")
	return nil
}

// RequestDeviceCode initiates the device code flow and returns the device code response.
func (tm *TokenManager) RequestDeviceCode(ctx context.Context) (*types.DeviceCodeResponse, error) {
	data := url.Values{
		"client_id": {tm.clientID},
		"scope":     {scopes},
	}

	req, err := http.NewRequestWithContext(ctx, "POST", deviceCodeURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := tm.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("device code request failed: %s", body)
	}

	var deviceCode types.DeviceCodeResponse
	if err := json.Unmarshal(body, &deviceCode); err != nil {
		return nil, err
	}
	return &deviceCode, nil
}

// PollForToken polls Microsoft's token endpoint until the user completes authentication.
func (tm *TokenManager) PollForToken(ctx context.Context, deviceCode *types.DeviceCodeResponse) (*types.TokenResponse, error) {
	interval := time.Duration(deviceCode.Interval) * time.Second
	if interval == 0 {
		interval = 5 * time.Second
	}

	deadline := time.Now().Add(time.Duration(deviceCode.ExpiresIn) * time.Second)

	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(interval):
		}

		tokens, err := tm.tryGetToken(ctx, deviceCode.DeviceCode)
		if err == nil {
			return tokens, nil
		}

		// Check if it's an expected "pending" error
		if strings.Contains(err.Error(), "authorization_pending") {
			continue
		}
		if strings.Contains(err.Error(), "slow_down") {
			interval += 5 * time.Second
			continue
		}

		// Unexpected error
		return nil, err
	}

	return nil, errors.New("authentication timed out")
}

func (tm *TokenManager) tryGetToken(ctx context.Context, deviceCode string) (*types.TokenResponse, error) {
	data := url.Values{
		"grant_type":  {"urn:ietf:params:oauth:grant-type:device_code"},
		"client_id":   {tm.clientID},
		"device_code": {deviceCode},
	}

	req, err := http.NewRequestWithContext(ctx, "POST", tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := tm.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token request failed: %s", body)
	}

	var tokens types.TokenResponse
	if err := json.Unmarshal(body, &tokens); err != nil {
		return nil, err
	}
	return &tokens, nil
}

// ClearTokens removes stored tokens (logout).
func (tm *TokenManager) ClearTokens() error {
	err := os.Remove(tm.tokensPath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("removing tokens: %w", err)
	}
	return nil
}
