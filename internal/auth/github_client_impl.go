package auth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/grigri/pv/internal/errors"
)

const (
	// GitHubAPIBaseURL is the base URL for GitHub API
	GitHubAPIBaseURL = "https://api.github.com"

	// defaultTimeout for HTTP requests
	defaultTimeout = 3 * time.Second
)

// githubClient implements the GitHubClient interface
type githubClient struct {
	baseURL string
	client  *http.Client
}

// NewGitHubClient creates a new GitHub API client
func NewGitHubClient() GitHubClient {
	return &githubClient{
		baseURL: GitHubAPIBaseURL,
		client: &http.Client{
			Timeout: defaultTimeout,
		},
	}
}

// GetAuthenticatedUser retrieves the authenticated user information
func (c *githubClient) GetAuthenticatedUser(token string) (*User, error) {
	req, err := http.NewRequest("GET", c.baseURL+"/user", nil)
	if err != nil {
		return nil, errors.NewAppError(errors.ErrNetwork, "failed to create request", err)
	}

	req.Header.Set("Authorization", "token "+token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, errors.ErrGitHubAPIUnavailable
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, errors.ErrInvalidToken
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.NewAppError(
			errors.ErrNetwork,
			fmt.Sprintf("GitHub API returned status %d", resp.StatusCode),
			nil,
		)
	}

	var user User
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, errors.NewAppError(errors.ErrNetwork, "failed to decode response", err)
	}

	return &user, nil
}

// ValidateScopes checks if the token has the required scopes
func (c *githubClient) ValidateScopes(token string) ([]string, error) {
	// Use HEAD request to minimize data transfer
	req, err := http.NewRequest("HEAD", c.baseURL+"/", nil)
	if err != nil {
		return nil, errors.NewAppError(errors.ErrNetwork, "failed to create request", err)
	}

	req.Header.Set("Authorization", "token "+token)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, errors.ErrGitHubAPIUnavailable
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, errors.ErrInvalidToken
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.NewAppError(
			errors.ErrNetwork,
			fmt.Sprintf("GitHub API returned status %d", resp.StatusCode),
			nil,
		)
	}

	// Extract scopes from X-OAuth-Scopes header
	scopesHeader := resp.Header.Get("X-OAuth-Scopes")
	if scopesHeader == "" {
		return []string{}, nil
	}

	// Parse comma-separated scopes
	scopes := strings.Split(scopesHeader, ",")
	for i := range scopes {
		scopes[i] = strings.TrimSpace(scopes[i])
	}

	return scopes, nil
}
