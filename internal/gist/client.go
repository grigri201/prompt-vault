package gist

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/go-github/v73/github"
)

// Client wraps the GitHub API client for Gist operations
type Client struct {
	github *github.Client
}

// NewClient creates a new GitHub client with authentication
func NewClient(token string) (*Client, error) {
	// Validate token
	token = strings.TrimSpace(token)
	if token == "" {
		return nil, errors.New("token is required")
	}

	// Create authenticated client
	client := github.NewClient(nil).WithAuthToken(token)
	
	return &Client{
		github: client,
	}, nil
}

// ValidateToken validates the provided token and returns the username
func (c *Client) ValidateToken(ctx context.Context) (string, error) {
	// Get authenticated user
	user, resp, err := c.github.Users.Get(ctx, "")
	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusUnauthorized {
			return "", fmt.Errorf("invalid token: %w", err)
		}
		if c.IsRateLimitError(err) {
			return "", fmt.Errorf("rate limit exceeded: %w", err)
		}
		return "", fmt.Errorf("failed to validate token: %w", err)
	}

	if user.Login == nil {
		return "", errors.New("unable to get username from API response")
	}

	return *user.Login, nil
}

// IsRateLimitError checks if an error is due to rate limiting
func (c *Client) IsRateLimitError(err error) bool {
	if err == nil {
		return false
	}

	// Check if it's a GitHub ErrorResponse
	var errResp *github.ErrorResponse
	if errors.As(err, &errResp) {
		// Check status code
		if errResp.Response != nil && errResp.Response.StatusCode == http.StatusForbidden {
			// Check for rate limit header
			remaining := errResp.Response.Header.Get("X-RateLimit-Remaining")
			if remaining == "0" {
				return true
			}
			
			// Check error message for rate limit text
			message := strings.ToLower(errResp.Message)
			if strings.Contains(message, "rate limit") {
				return true
			}
		}
	}

	// Also check the error string directly
	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "rate limit")
}

// GetAPIError extracts the GitHub API error from an error
func (c *Client) GetAPIError(err error) *github.ErrorResponse {
	if err == nil {
		return nil
	}

	var errResp *github.ErrorResponse
	if errors.As(err, &errResp) {
		return errResp
	}

	return nil
}

// GetGitHubClient returns the underlying GitHub client for advanced operations
func (c *Client) GetGitHubClient() *github.Client {
	return c.github
}

// CreateGist creates a new private Gist with the given content
func (c *Client) CreateGist(ctx context.Context, gistName, description, content string) (string, string, error) {
	// Validate inputs
	if gistName == "" {
		return "", "", errors.New("gist name is required")
	}
	if content == "" {
		return "", "", errors.New("content is required")
	}

	// Prepare the gist request
	gistReq := &github.Gist{
		Description: github.String(description),
		Public:      github.Bool(false), // Always create private gists
		Files: map[github.GistFilename]github.GistFile{
			github.GistFilename(gistName + ".yaml"): {
				Content: github.String(content),
			},
		},
	}

	// Create the gist
	gist, _, err := c.github.Gists.Create(ctx, gistReq)
	if err != nil {
		if c.IsRateLimitError(err) {
			return "", "", fmt.Errorf("rate limit exceeded: %w", err)
		}
		return "", "", fmt.Errorf("failed to create gist: %w", err)
	}

	if gist.ID == nil || gist.HTMLURL == nil {
		return "", "", errors.New("unexpected response from GitHub API")
	}

	return *gist.ID, *gist.HTMLURL, nil
}

// UpdateGist updates an existing Gist with new content
func (c *Client) UpdateGist(ctx context.Context, gistID, gistName, description, content string) (string, error) {
	// Validate inputs
	if gistID == "" {
		return "", errors.New("gist ID is required")
	}
	if gistName == "" {
		return "", errors.New("gist name is required")
	}
	if content == "" {
		return "", errors.New("content is required")
	}

	// Prepare the update request
	gistReq := &github.Gist{
		Description: github.String(description),
		Files: map[github.GistFilename]github.GistFile{
			github.GistFilename(gistName + ".yaml"): {
				Content: github.String(content),
			},
		},
	}

	// Update the gist
	gist, resp, err := c.github.Gists.Edit(ctx, gistID, gistReq)
	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusNotFound {
			return "", fmt.Errorf("gist not found: %w", err)
		}
		if resp != nil && resp.StatusCode == http.StatusForbidden && !c.IsRateLimitError(err) {
			return "", fmt.Errorf("permission denied: %w", err)
		}
		if c.IsRateLimitError(err) {
			return "", fmt.Errorf("rate limit exceeded: %w", err)
		}
		return "", fmt.Errorf("failed to update gist: %w", err)
	}

	if gist.HTMLURL == nil {
		return "", errors.New("unexpected response from GitHub API")
	}

	return *gist.HTMLURL, nil
}

// GetGist retrieves a Gist by ID
func (c *Client) GetGist(ctx context.Context, gistID string) (*github.Gist, error) {
	// Validate input
	if gistID == "" {
		return nil, errors.New("gist ID is required")
	}

	// Get the gist
	gist, resp, err := c.github.Gists.Get(ctx, gistID)
	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusNotFound {
			return nil, fmt.Errorf("gist not found: %w", err)
		}
		if c.IsRateLimitError(err) {
			return nil, fmt.Errorf("rate limit exceeded: %w", err)
		}
		return nil, fmt.Errorf("failed to get gist: %w", err)
	}

	return gist, nil
}

// DeleteGist deletes a Gist by ID
func (c *Client) DeleteGist(ctx context.Context, gistID string) error {
	// Validate input
	if gistID == "" {
		return errors.New("gist ID is required")
	}

	// Delete the gist
	resp, err := c.github.Gists.Delete(ctx, gistID)
	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusNotFound {
			return fmt.Errorf("gist not found: %w", err)
		}
		if resp != nil && resp.StatusCode == http.StatusForbidden && !c.IsRateLimitError(err) {
			return fmt.Errorf("permission denied: %w", err)
		}
		if c.IsRateLimitError(err) {
			return fmt.Errorf("rate limit exceeded: %w", err)
		}
		return fmt.Errorf("failed to delete gist: %w", err)
	}

	return nil
}